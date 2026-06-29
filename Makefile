build:
	go build -o bin/api ./cmd/api/

test:
	go test ./... -v -race -cover

lint:
	golangci-lint run

run:
	go run ./cmd/api/

# 本地開發環境：同時啟動 PostgreSQL, KeyDB, ScyllaDB
compose-up:
	docker compose --env-file .env.dev -f .docker/docker-compose.dev.yml up -d

compose-down:
	docker compose --env-file .env.dev -f .docker/docker-compose.dev.yml down

compose-down-v:
	docker compose --env-file .env.dev -f .docker/docker-compose.dev.yml down -v

.PHONY: build test lint run compose-up compose-down compose-down-v
