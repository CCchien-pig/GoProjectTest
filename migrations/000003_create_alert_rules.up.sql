-- 000003_create_alert_rules.up.sql
CREATE TABLE IF NOT EXISTS alert_rules (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id     UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    metric_name   VARCHAR(100) NOT NULL,           -- e.g. temperature, voltage
    operator      VARCHAR(10)  NOT NULL,           -- gt, lt, gte, lte, eq
    threshold     DOUBLE PRECISION NOT NULL,
    severity      VARCHAR(20) NOT NULL DEFAULT 'warning', -- info / warning / critical
    is_enabled    BOOLEAN NOT NULL DEFAULT true,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
