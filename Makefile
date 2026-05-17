.PHONY: build build-server build-frontend docker-build docker-up docker-down test clean dev-frontend

# Build targets
build: build-frontend build-server

build-frontend:
	@echo "Building frontend..."
	cd web && npm install && npm run build
	@echo "Copying frontend dist to cmd/server/web/dist..."
	rm -rf cmd/server/web/dist
	cp -r web/dist cmd/server/web/dist
	@echo "Frontend build complete."

build-server: build-frontend
	CGO_ENABLED=1 go build -o bin/server ./cmd/server

build-worker:
	CGO_ENABLED=1 go build -o bin/worker ./cmd/worker

# Docker targets
docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-scale:
	docker compose up -d --scale worker=$(N)

# Development
dev-frontend:
	cd web && npm run dev

run-server: build-server
	./bin/server -config configs/server.yaml

run-worker: build-worker
	./bin/worker -config configs/worker.yaml

# Testing
test:
	go test ./...

# Utility
clean:
	rm -rf bin/
	rm -rf web/dist
	rm -rf cmd/server/web/dist
	docker compose down -v

# Quick start
start: docker-build docker-up
	@echo "API server running at http://localhost:8080"
	@echo "NATS monitoring at http://localhost:8222"
	@echo "Scale workers: make docker-scale N=5"
