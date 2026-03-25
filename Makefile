.PHONY: run build test db-up db-down migrate-up migrate-down docker-up docker-down

# Run the server locally (requires Postgres running)
run:
	go run ./cmd/server

# Build the binary
build:
	go build -o bin/Book-of-Crane ./cmd/server

# Run tests
test:
	go test ./... -v

# Start just the database
db-up:
	docker compose up db -d

# Stop the database
db-down:
	docker compose down

# Run all containers
docker-up:
	docker compose up --build

# Stop all containers
docker-down:
	docker compose down

# Remove database volume (fresh start)
db-reset:
	docker compose down -v
	docker compose up db -d
