package keydb

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client 封裝 KeyDB 連線
type Client struct {
	Client redis.UniversalClient
}

// NewClient 建立 KeyDB 連線（支援 Cluster 與單機模式）
func NewClient(addr string, password string, clusterMode bool) (*Client, error) {
	var rdb redis.UniversalClient

	if clusterMode {
		rdb = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    []string{addr},
			Password: password,
		})
	} else {
		rdb = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
		})
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
