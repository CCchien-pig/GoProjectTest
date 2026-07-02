-- 000004_add_indexes.up.sql
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- pg_trgm GIN indexes for fuzzy search
CREATE INDEX IF NOT EXISTS idx_devices_device_code_trgm ON devices USING gin (device_code gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_devices_name_trgm ON devices USING gin (name gin_trgm_ops);

-- B-Tree indexes for quick filtering
CREATE INDEX IF NOT EXISTS idx_devices_device_type ON devices (device_type);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices (status);
CREATE INDEX IF NOT EXISTS idx_devices_location ON devices (location);
