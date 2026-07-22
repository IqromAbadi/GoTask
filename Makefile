.PHONY: run build test test-cover migrate-up migrate-down migrate-create sqlc docker-up docker-down lint fmt

APP_NAME=gotask
MIGRATIONS_DIR=db/migrations
DATABASE_URL?=postgres://gotask:gotask@localhost:5432/gotask?sslmode=disable

# Run the application
run:
	go run ./cmd/api

# Build the binary
build:
	go build -o bin/$(APP_NAME) ./cmd/api

# Run tests
test:
	go test ./... -count=1

# Run tests with coverage
test-cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -html=coverage.out -o coverage.html

# Run database migrations up
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

# Run database migrations down
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down

# Create a new migration
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# Generate sqlc code
sqlc:
	sqlc generate -f db/sqlc.yaml

# Start Docker Compose
docker-up:
	docker compose up -d

# Stop Docker Compose
docker-down:
	docker compose down

# Run linter
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...

# Tidy dependencies
tidy:
	go mod tidy
