.PHONY: help run build test lint migrate-up migrate-down migrate-create docker-up docker-down clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ====================================================================
# Development
# ====================================================================

run: ## Run backend locally (needs PostgreSQL, Redis, NATS)
	cd backend && go run ./cmd/server/

build: ## Build backend binary
	cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server/

test: ## Run tests
	cd backend && go test ./... -v -count=1

lint: ## Run linters
	cd backend && go vet ./...

fmt: ## Format code
	cd backend && go fmt ./...

mod: ## Tidy go modules
	cd backend && go mod tidy

# ====================================================================
# Database Migrations
# ====================================================================

migrate-up: ## Run all pending migrations
	cd backend && migrate -database "postgres://sabtbrooker:sabtbrooker@localhost:5433/sabtbrooker?sslmode=disable" -path migrations up

migrate-down: ## Rollback last migration
	cd backend && migrate -database "postgres://sabtbrooker:sabtbrooker@localhost:5433/sabtbrooker?sslmode=disable" -path migrations down 1

migrate-create: ## Create new migration (usage: make migrate-create NAME=add_foo)
	cd backend && migrate create -ext sql -dir migrations -seq $(NAME)

migrate-install: ## Install golang-migrate CLI
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# ====================================================================
# Docker
# ====================================================================

docker-up: ## Start all services
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-build: ## Build backend image
	docker-compose build backend

docker-logs: ## Tail logs
	docker-compose logs -f backend

docker-reset: ## Reset all data (CAREFUL!)
	docker-compose down -v
	docker-compose up -d

# ====================================================================
# Utilities
# ====================================================================

clean: ## Remove build artifacts
	cd backend && rm -rf bin/

db-shell: ## Open PostgreSQL shell
	docker exec -it sabtbrooker-db psql -U sabtbrooker -d sabtbrooker

redis-shell: ## Open Redis CLI
	docker exec -it sabtbrooker-redis redis-cli

nats-mon: ## Open NATS monitoring
	open http://localhost:8222

minio-console: ## Open MinIO console
	open http://localhost:9001
