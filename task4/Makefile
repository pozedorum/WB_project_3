.PHONY: build up down logs clean

build:
	docker compose build

lint:
	golangci-lint run ./...
	go vet ./...

rebuild:
	docker compose down -v
	docker compose up --build

up:
	dockercompose up -d

down:
	docker compose down

logs:
	docker compose logs -f image-processor

clean:
	docker compose down -v
	docker system prune -f

test:
	go test ./... -v