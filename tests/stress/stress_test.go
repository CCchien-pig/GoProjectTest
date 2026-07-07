//go:build stress

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// safePercentileIndex 安全計算百分位數索引，防止越界 (Finding #10)
func safePercentileIndex(total int, pct float64) int {
	idx := int(float64(total) * pct)
	if idx >= total {
		idx = total - 1
	}
	return idx
}

var (
	baseURL = getEnv("API_BASE_URL", "http://localhost:8080")
	client  = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
)

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}

type DeviceReq struct {
	DeviceCode string `json:"device_code"`
	Name       string `json:"name"`
	DeviceType string `json:"device_type"`
	Location   string `json:"location"`
	Status     string `json:"status"`
}

type DeviceResp struct {
	Code int `json:"code"`
	Data struct {
		ID uuid.UUID `json:"id"`
	} `json:"data"`
}

type TelemetryPoint struct {
	RecordedAt time.Time         `json:"recorded_at"`
	MetricName string            `json:"metric_name"`
	Value      float64           `json:"value"`
	Unit       string            `json:"unit"`
	Tags       map[string]string `json:"tags"`
}

type BatchTelemetryReq struct {
	Points []TelemetryPoint `json:"points"`
}

func TestStress(t *testing.T) {
	// Skip if short test
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// 1. Verify health
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		t.Fatalf("API Server is not running at %s: %v", baseURL, err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("API Server is not healthy, status: %d", resp.StatusCode)
	}

	deviceCount, _ := strconv.Atoi(getEnv("STRESS_DEVICE_COUNT", "100")) // 預設 100 方便快速本地驗證，可設成 1000
	pointsPerDevice, _ := strconv.Atoi(getEnv("STRESS_POINTS_PER_DEVICE", "10"))
	concurrency, _ := strconv.Atoi(getEnv("STRESS_CONCURRENCY", "10"))
	durationSec, _ := strconv.Atoi(getEnv("STRESS_DURATION_SEC", "10"))

	fmt.Printf("開始壓力測試 - 伺服器: %s\n", baseURL)
	fmt.Printf("設備數: %d, 每個設備寫入點數: %d, 並發數: %d, 混合測試時長: %ds\n", deviceCount, pointsPerDevice, concurrency, durationSec)

	// Step 1: 批量建立設備
	fmt.Println("Step 1: 正在批量建立設備...")
	deviceIDs := make([]uuid.UUID, 0, deviceCount)
	var mu sync.Mutex
	var wg sync.WaitGroup

	sem := make(chan struct{}, concurrency)
	for i := 0; i < deviceCount; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			req := DeviceReq{
				DeviceCode: fmt.Sprintf("STRESS-DEV-%d-%d", time.Now().UnixNano(), index),
				Name:       fmt.Sprintf("Stress Device %d", index),
				DeviceType: "sensor",
				Location:   "TPE-Stress",
				Status:     "active",
			}
			bodyBytes, _ := json.Marshal(req)
			hResp, err := client.Post(baseURL+"/api/v1/devices", "application/json", bytes.NewBuffer(bodyBytes))
			if err != nil {
				return
			}
			defer func() { _ = hResp.Body.Close() }()

			if hResp.StatusCode == http.StatusOK {
				var r DeviceResp
				if err := json.NewDecoder(hResp.Body).Decode(&r); err == nil {
					mu.Lock()
					deviceIDs = append(deviceIDs, r.Data.ID)
					mu.Unlock()
				}
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("成功建立 %d / %d 個設備。\n", len(deviceIDs), deviceCount)
	if len(deviceIDs) == 0 {
		t.Fatal("沒有成功建立任何設備，中止測試")
	}

	// Step 2: 寫入遙測數據
	fmt.Println("Step 2: 正在寫入遙測歷史數據...")
	for _, devID := range deviceIDs {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			points := make([]TelemetryPoint, pointsPerDevice)
			for j := 0; j < pointsPerDevice; j++ {
				points[j] = TelemetryPoint{
					RecordedAt: time.Now().Add(time.Duration(-j) * time.Minute),
					MetricName: "temperature",
					Value:      20.0 + rand.Float64()*30.0,
					Unit:       "C",
					Tags:       map[string]string{"env": "stress"},
				}
			}
			req := BatchTelemetryReq{Points: points}
			bodyBytes, _ := json.Marshal(req)
			url := fmt.Sprintf("%s/api/v1/devices/%s/telemetry", baseURL, id.String())
			hResp, err := client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
			if err != nil {
				return
			}
			_ = hResp.Body.Close()
		}(devID)
	}
	wg.Wait()
	fmt.Println("歷史遙測數據寫入完成。")

	// Step 3: 混合讀寫負載 (70% 讀 / 30% 寫)
	fmt.Println("Step 3: 開始混合負載測試 (70% 讀 / 30% 寫)...")
	stopChan := make(chan struct{})
	latencies := make([]time.Duration, 0, 10000)
	var latMu sync.Mutex

	startTime := time.Now()
	for i := 0; i < concurrency; i++ {
		// Finding #9: 每個 goroutine 使用獨立的 rand 實例，陥免全局 mutex 競爭
		go func(r *rand.Rand) {
			for {
				select {
				case <-stopChan:
					return
				default:
					devID := deviceIDs[r.Intn(len(deviceIDs))]
					isRead := r.Float64() < 0.70

					reqStartTime := time.Now()
					var err error
					var hResp *http.Response

					if isRead {
						// 讀取設備即時狀態
						url := fmt.Sprintf("%s/api/v1/devices/%s/status", baseURL, devID.String())
						hResp, err = client.Get(url)
					} else {
						// 寫入單筆遙測
						url := fmt.Sprintf("%s/api/v1/devices/%s/telemetry", baseURL, devID.String())
						req := BatchTelemetryReq{
							Points: []TelemetryPoint{
								{
									RecordedAt: time.Now(),
									MetricName: "temperature",
									Value:      25.0 + r.Float64()*10.0,
									Unit:       "C",
								},
							},
						}
						bodyBytes, _ := json.Marshal(req)
						hResp, err = client.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
					}

					if err == nil {
						_ = hResp.Body.Close()
						duration := time.Since(reqStartTime)
						latMu.Lock()
						latencies = append(latencies, duration)
						latMu.Unlock()
					}
				}
			}
		}(rand.New(rand.NewSource(time.Now().UnixNano() + int64(i))))
	}

	time.Sleep(time.Duration(durationSec) * time.Second)
	close(stopChan)

	totalDuration := time.Since(startTime)
	totalRequests := len(latencies)
	qps := float64(totalRequests) / totalDuration.Seconds()

	fmt.Printf("\n--- 壓力測試報告 ---\n")
	fmt.Printf("總測試時長: %v\n", totalDuration)
	fmt.Printf("總成功請求數: %d\n", totalRequests)
	fmt.Printf("每秒請求數 (QPS): %.2f req/s\n", qps)

	if totalRequests > 0 {
		sort.Slice(latencies, func(i, j int) bool {
			return latencies[i] < latencies[j]
		})

		// Finding #10: 使用 safePercentileIndex 防止越界
		p50 := latencies[safePercentileIndex(totalRequests, 0.50)]
		p95 := latencies[safePercentileIndex(totalRequests, 0.95)]
		p99 := latencies[safePercentileIndex(totalRequests, 0.99)]

		fmt.Printf("P50 延遲: %v\n", p50)
		fmt.Printf("P95 延遲: %v\n", p95)
		fmt.Printf("P99 延遲: %v\n", p99)
	} else {
		t.Fatal("未完成任何請求")
	}
}
