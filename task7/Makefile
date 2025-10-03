.PHONY: build run test clean migrate test help

help:

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

stop:
	docker compose down

logs:
	docker compose logs -f app

clean:
	docker compose down -v
	rm -f sales-tracker

sql_connect:
	docker exec -it task6-postgres-1 psql -U postgres -d WarehouseControl
