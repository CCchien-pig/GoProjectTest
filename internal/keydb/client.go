package keydb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client 封裝 KeyDB 連線
type Client struct {
	Client redis.UniversalClient
}

// NewClient 建立 KeyDB 連線（支援 Cluster 與單機模式，含 TLS 支援）
func NewClient(addr string, password string, clusterMode bool, useTLS bool, caCertPath string, insecure bool) (*Client, error) {
	var rdb redis.UniversalClient

	var tlsConfig *tls.Config
	if useTLS {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: insecure, //nolint:gosec // 開發環境自簽憑證允許，生產環境務必設 false
		}
		// Finding #11: 明確警告 InsecureSkipVerify=true 僅適用開發環境
		if insecure {
			slog.Warn("KeyDB TLS InsecureSkipVerify is enabled — do NOT use this in production", "addr", addr)
		}
		if caCertPath != "" {
			caCert, err := os.ReadFile(caCertPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read KeyDB CA cert file %q: %w", caCertPath, err)
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return nil, fmt.Errorf("failed to parse KeyDB CA cert from PEM in %q", caCertPath)
			}
			tlsConfig.RootCAs = caCertPool
		}
	}

	if clusterMode {
		opts := &redis.ClusterOptions{
			Addrs:    []string{addr},
			Password: password,
		}
		if useTLS {
			opts.TLSConfig = tlsConfig
		}
		rdb = redis.NewClusterClient(opts)
	} else {
		opts := &redis.Options{
			Addr:     addr,
			Password: password,
		}
		if useTLS {
			opts.TLSConfig = tlsConfig
		}
		rdb = redis.NewClient(opts)
	}

	// 測試連線，若本地 KeyDB 尚未啟動會報錯但不一定強制中止（容錯由呼叫方處理）
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping KeyDB failed: %w", err)
	}

	return &Client{Client: rdb}, nil
}

// Close 關閉連線
func (c *Client) Close() error {
	if c.Client != nil {
		return c.Client.Close()
	}
	return nil
}
