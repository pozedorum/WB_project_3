.PHONY: build run test clean migrate

build:
	docker compose build

run:
	docker compose up

lint:
	golangci-lint run ./...
	go vet ./...

rebuild:
	docker compose down -v
	docker compose up --build

service_rebuild:
	docker compose stop app
	docker compose up --build -d app && docker compose logs -f app

run-detached:
	docker compose up -d

stop:
	docker compose down

logs:
	docker compose logs -f app

clean:
	docker compose down -v
	rm -f commentTree
	rm -rf postgres_data/*

migrate:
	docker compose exec postgres psql -U postgres -d comments_db -f /docker-entrypoint-initdb.d/init.sql