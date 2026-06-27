build:
	go build -o bin/api ./cmd/api/

test:
	go test ./... -v -race -cover

lint:
	golangci-lint run

run:
	go run ./cmd/api/

# ScyllaDB 只在本地跑
compose-up:
	docker compose --env-file .env.dev -f .docker/docker-compose.dev.yml up -d

compose-down:
	docker compose --env-file .env.dev -f .docker/docker-compose.dev.yml down

compose-down-v:
	docker compose --env-file .env.dev -f .docker/docker-compose.dev.yml down -v

# SSH Tunnel 管理（連到 GCP 的 PostgreSQL 和 KeyDB）
tunnel-start:
	bash .docker/ssh/start_tunnels.sh start

tunnel-stop:
	bash .docker/ssh/start_tunnels.sh stop

tunnel-status:
	bash .docker/ssh/start_tunnels.sh status

# GCP 機器上的 docker-compose（PostgreSQL + KeyDB）
gcp-up:
	@echo "請在 GCP 機器上執行: docker compose -f docker-compose.gcp.yml up -d"

.PHONY: build test lint run compose-up compose-down compose-down-v tunnel-start tunnel-stop tunnel-status gcp-up
