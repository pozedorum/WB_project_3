.PHONY: build run test clean migrate

build:
	docker-compose build

run:
	docker-compose up

lint:
	golangci-lint run ./...
	go vet ./...

rebuild:
	docker-compose down -v
	docker-compose up --build
	
run-detached:
	docker-compose up -d

stop:
	docker-compose down

logs:
	docker-compose logs -f app

clean:
	docker-compose down -v
	rm -f delayed-notifier

migrate:
	docker-compose exec postgres psql -U postgres -d delayed_notifier -f /docker-entrypoint-initdb.d/init.sql