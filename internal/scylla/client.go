package scylla

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

// Client 封裝 ScyllaDB 的 session 操作
type Client struct {
	Session *gocql.Session
}

// NewClient 建立 ScyllaDB 連線，並確認 Schema 已經建立
func NewClient(hosts []string, keyspace string) (*Client, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second

	// 先不指定 keyspace 建立 session，以便在此建立 keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create initial session: %w", err)
	}

	client := &Client{Session: session}
	if err := client.EnsureSchema(keyspace); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to ensure schema: %w", err)
	}

	session.Close()

	// 重新指定 keyspace 連線
	cluster.Keyspace = keyspace
	finalSession, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create final session with keyspace: %w", err)
	}
	client.Session = finalSession

	return client, nil
}

// Close 關閉 ScyllaDB 連線
func (c *Client) Close() {
	if c.Session != nil && !c.Session.Closed() {
		c.Session.Close()
	}
}

// EnsureSchema 確保 keyspace 和 table 存在
func (c *Client) EnsureSchema(keyspace string) error {
	// 建立 keyspace (SimpleStrategy 在本地開發最適合)
	err := c.Session.Query(fmt.Sprintf(`
		CREATE KEYSPACE IF NOT EXISTS %s
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
	`, keyspace)).Exec()
	if err != nil {
		return err
	}

	// 遙測時序資料表
	err = c.Session.Query(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.telemetry (
			device_id uuid,
			date date,
			recorded_at timestamp,
			metric_name text,
			value double,
			unit text,
			tags map<text, text>,
			PRIMARY KEY ((device_id, date), recorded_at, metric_name)
		) WITH CLUSTERING ORDER BY (recorded_at DESC, metric_name ASC)
		AND default_time_to_live = 7776000;
	`, keyspace)).Exec()
	if err != nil {
		return err
	}

	// 告警事件時序資料表
	err = c.Session.Query(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.alert_events (
			device_id uuid,
			month text,
			triggered_at timestamp,
			rule_id uuid,
			metric_name text,
			metric_value double,
			threshold double,
			severity text,
			acknowledged boolean,
			PRIMARY KEY ((device_id, month), triggered_at, rule_id)
		) WITH CLUSTERING ORDER BY (triggered_at DESC, rule_id ASC)
		AND default_time_to_live = 31536000;
	`, keyspace)).Exec()
	return err
}
