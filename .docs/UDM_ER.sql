CREATE TABLE "roles" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" "VARCHAR(50)" UNIQUE NOT NULL,
  "description" "VARCHAR(200)",
  "created_at" "TIMESTAMPTZ" NOT NULL DEFAULT (now()),
  "updated_at" "TIMESTAMPTZ" NOT NULL DEFAULT (now())
);

CREATE TABLE "permissions" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "name" "VARCHAR(50)" UNIQUE NOT NULL,
  "description" "VARCHAR(200)",
  "created_at" "TIMESTAMPTZ" NOT NULL DEFAULT (now()),
  "updated_at" "TIMESTAMPTZ" NOT NULL DEFAULT (now())
);

CREATE TABLE "role_permissions" (
  "role_id" UUID NOT NULL,
  "permission_id" UUID NOT NULL,
  PRIMARY KEY ("role_id", "permission_id")
);

CREATE TABLE "users" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "username" "VARCHAR(100)" UNIQUE NOT NULL,
  "email" "VARCHAR(255)" UNIQUE NOT NULL,
  "password_hash" "VARCHAR(255)" NOT NULL,
  "role_id" UUID NOT NULL,
  "is_active" BOOLEAN NOT NULL DEFAULT true,
  "created_at" "TIMESTAMPTZ" NOT NULL DEFAULT (now()),
  "updated_at" "TIMESTAMPTZ" NOT NULL DEFAULT (now())
);

CREATE TABLE "devices" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "device_code" "VARCHAR(50)" UNIQUE NOT NULL,
  "name" "VARCHAR(200)" NOT NULL,
  "device_type" "VARCHAR(50)" NOT NULL,
  "location" "VARCHAR(200)",
  "metadata" JSONB DEFAULT '{}',
  "status" "VARCHAR(20)" NOT NULL DEFAULT 'inactive',
  "created_at" "TIMESTAMPTZ" NOT NULL DEFAULT (now()),
  "updated_at" "TIMESTAMPTZ" NOT NULL DEFAULT (now())
);

CREATE TABLE "user_devices" (
  "user_id" UUID NOT NULL,
  "device_id" UUID NOT NULL,
  "created_at" "TIMESTAMPTZ" NOT NULL DEFAULT (now()),
  PRIMARY KEY ("user_id", "device_id")
);

CREATE TABLE "alert_rules" (
  "id" UUID PRIMARY KEY DEFAULT (gen_random_uuid()),
  "device_id" UUID NOT NULL,
  "metric_name" "VARCHAR(100)" NOT NULL,
  "operator" "VARCHAR(10)" NOT NULL,
  "threshold" "DOUBLEPRECISION" NOT NULL,
  "severity" "VARCHAR(20)" NOT NULL DEFAULT 'warning',
  "is_enabled" BOOLEAN NOT NULL DEFAULT true,
  "created_at" "TIMESTAMPTZ" NOT NULL DEFAULT (now()),
  "updated_at" "TIMESTAMPTZ" NOT NULL DEFAULT (now())
);

CREATE INDEX "idx_users_role_id" ON "users" ("role_id");

CREATE INDEX "idx_devices_device_code_trgm" ON "devices" USING GIN ("device_code");

CREATE INDEX "idx_devices_name_trgm" ON "devices" USING GIN ("name");

CREATE INDEX "idx_devices_device_type" ON "devices" ("device_type");

CREATE INDEX "idx_devices_status" ON "devices" ("status");

CREATE INDEX "idx_devices_location" ON "devices" ("location");

CREATE INDEX "idx_user_devices_user_id" ON "user_devices" ("user_id");

CREATE INDEX "idx_user_devices_device_id" ON "user_devices" ("device_id");

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "role_permissions" ADD FOREIGN KEY ("permission_id") REFERENCES "permissions" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "users" ADD FOREIGN KEY ("role_id") REFERENCES "roles" ("id") DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_devices" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "user_devices" ADD FOREIGN KEY ("device_id") REFERENCES "devices" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE "alert_rules" ADD FOREIGN KEY ("device_id") REFERENCES "devices" ("id") ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;
