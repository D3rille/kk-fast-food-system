# Makefile for NextGen Kiosk-to-Kitchen Fast Food System

# Default database URL for local environment
DATABASE_URL ?= postgres://app:devpassword123@localhost:5432/fastfood?sslmode=disable

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build           Build the Go API server binary"
	@echo "  run             Run the Go API server locally"
	@echo "  test            Run Go tests"
	@echo "  lint            Run golangci-lint"
	@echo "  vet             Run go vet"
	@echo "  fmt             Format Go code"
	@echo "  migrate-up      Apply database migrations"
	@echo "  migrate-down    Revert database migrations"
	@echo "  migrate-create  Create a new migration (usage: make migrate-create name=name_of_migration)"
	@echo "  docker-up       Build and start all services in Docker Compose"
	@echo "  docker-down     Stop and remove Docker Compose containers"
	@echo "  mock-seed       Seed the database with mock data for manual testing"
	@echo "  clean           Clean built binaries"

.PHONY: build
build:
	go build -o bin/server ./cmd/api

.PHONY: run
run:
	go run ./cmd/api

.PHONY: test
test:
	go test -v -race -count=1 ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: migrate-up
migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

.PHONY: migrate-create
migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

.PHONY: docker-up
docker-up:
	docker compose up -d --build

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-rerun
docker-rerun:
	docker compose down && docker compose up -d --build

.PHONY: mock-seed
mock-seed:
	DATABASE_URL=$(DATABASE_URL) go run ./cmd/seed

.PHONY: clean
clean:
	rm -rf bin/
