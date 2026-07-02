-- 000002_create_devices.up.sql
CREATE TABLE IF NOT EXISTS devices (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_code   VARCHAR(50)  UNIQUE NOT NULL,   -- 設備編號，如 SENSOR-TPE-001
    name          VARCHAR(200) NOT NULL,
    device_type   VARCHAR(50)  NOT NULL,           -- sensor / controller / gateway
    location      VARCHAR(200),                    -- 廠區/區域
    metadata      JSONB DEFAULT '{}',              -- 彈性欄位（韌體版本、IP、型號等）
    owner_id      UUID REFERENCES users(id),
    status        VARCHAR(20)  NOT NULL DEFAULT 'inactive', -- active / inactive / maintenance
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
