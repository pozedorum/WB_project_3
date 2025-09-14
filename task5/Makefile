PHONY: build run test clean migrate test setup test-booking test-expiry test-concurrency help

help:
	@echo "Доступные цели:"
	@echo "  make setup     - Настройка тестовой среды (регистрация пользователей)"
	@echo "  make test      - Запуск всех тестов"
	@echo "  make test-booking - Тестирование процесса бронирования"
	@echo "  make test-expiry - Тестирование истечения срока брони"
	@echo "  make test-concurrency - Тестирование конкурентного доступа"
	@echo "  make clean     - Очистка временных файлов"
	@echo "  make help      - Показать эту справку"

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

clean: test_clean
	docker compose down -v
	rm -f event_booker
	@echo "🧹 Очистка временных файлов..."
	@rm -f .admin_token .user1_token .user2_token .event_id .booking1_code .booking2_code 2>/dev/null || true

sql_connect:
	docker exec -it task5-postgres-1 psql -U postgres -d eventbooker

# Переменные
SERVER_URL := http://localhost:8080
ADMIN_EMAIL := admin@eventbooker.com
USER1_EMAIL := user1@eventbooker.com
USER2_EMAIL := user2@eventbooker.com
PASSWORD := password123
EVENT_NAME := "Test Event with Short Expiry"
SHORT_LIFESPAN := "1m"  # Короткое время для тестирования экспирации

# Цели тестирования
test: setup test-booking test-expiry test-concurrency
	@echo "✅ Все тесты пройдены успешно!"

setup: test_clean
	@echo "🔄 Настройка тестовой среды..."
	
	# Регистрация администратора
	@echo "👤 Регистрируем администратора..."
	@curl -s -X POST $(SERVER_URL)/register \
		-H "Content-Type: application/json" \
		-d '{"email":"$(ADMIN_EMAIL)","password":"$(PASSWORD)","name":"Admin User"}' | \
		jq -r '.user.id' > .admin_id
	@echo "Admin ID: $$(cat .admin_id)"
	@sleep 1
	
	# Регистрация пользователя 1
	@echo "👤 Регистрируем пользователя 1..."
	@curl -s -X POST $(SERVER_URL)/register \
		-H "Content-Type: application/json" \
		-d '{"email":"$(USER1_EMAIL)","password":"$(PASSWORD)","name":"Test User 1"}' | \
		jq -r '.token' > .user1_token
	@echo "User1 Token: $$(cat .user1_token)"
	@sleep 1
	
	# Регистрация пользователя 2  
	@echo "👤 Регистрируем пользователя 2..."
	@curl -s -X POST $(SERVER_URL)/register \
		-H "Content-Type: application/json" \
		-d '{"email":"$(USER2_EMAIL)","password":"$(PASSWORD)","name":"Test User 2"}' | \
		jq -r '.token' > .user2_token
	@echo "User2 Token: $$(cat .user2_token)"
	@sleep 1
	
	# Логин администратора для получения токена	
	@echo "🔐 Логин администратора..."
	@curl -s -X POST $(SERVER_URL)/login \
		-H "Content-Type: application/json" \
		-d '{"email":"$(ADMIN_EMAIL)","password":"$(PASSWORD)"}' | \
		jq -r '.token' > .admin_token
	@echo "Admin Token: $$(cat .admin_token)"
	@sleep 1
	
	@echo "✅ Настройка завершена!"

test-booking:
	@echo "🎫 Тестирование процесса бронирования..."
	
	# Создаем мероприятие от имени администратора
	@echo "📅 Создаем тестовое мероприятие..."
	@curl -s -X POST $(SERVER_URL)/events \
		-H "Authorization: Bearer $$(cat .admin_token)" \
		-H "Content-Type: application/json" \
		-d '{"name":$(EVENT_NAME),"date":"2025-12-25T19:00:00Z","total_seats":10,"life_span":"15m"}' | \
		jq -r '.event.id' > .event_id
	@echo "Event ID: $$(cat .event_id)"
	@sleep 1
	
	# Бронируем места от имени пользователя 1
	@echo "🎟️ Бронируем 2 места для пользователя 1..."
	@curl -s -X POST $(SERVER_URL)/events/$$(cat .event_id)/book \
		-H "Authorization: Bearer $$(cat .user1_token)" \
		-H "Content-Type: application/json" \
		-d '{"seat_count":2}' | \
		jq -r '.booking.booking_code' > .booking1_code
	@echo "Booking Code 1: $$(cat .booking1_code)"
	@sleep 1
	
	# Бронируем места от имени пользователя 2
	@echo "🎟️ Бронируем 3 места для пользователя 2..."
	@curl -s -X POST $(SERVER_URL)/events/$$(cat .event_id)/book \
		-H "Authorization: Bearer $$(cat .user2_token)" \
		-H "Content-Type: application/json" \
		-d '{"seat_count":3}' | \
		jq -r '.booking.booking_code' > .booking2_code
	@echo "Booking Code 2: $$(cat .booking2_code)"
	@sleep 1
	
	# Подтверждаем бронь пользователя 1
	@echo "✅ Подтверждаем бронь пользователя 1..."
	@curl -s -X POST $(SERVER_URL)/events/$$(cat .event_id)/confirm \
		-H "Authorization: Bearer $$(cat .user1_token)" \
		-H "Content-Type: application/json" \
		-d '{"booking_code": "$$(cat .booking1_code)"}'
	@sleep 1
	
	# Проверяем доступные места (должно быть 5)
	@echo "🔍 Проверяем доступные места..."
	@curl -s -X GET $(SERVER_URL)/events/$$(cat .event_id) | \
		jq -r '.event.available_seats' > .available_seats
	@if [ "$$(cat .available_seats)" -eq 5 ]; then \
		echo "✅ Тест бронирования пройден!"; \
	else \
		echo "❌ Ошибка: ожидалось 5 мест, получено $$(cat .available_seats)"; \
		exit 1; \
	fi

test-expiry:
	@echo "⏰ Тестирование истечения срока бронирования..."
	
	# Создаем мероприятие с коротким сроком жизни брони
	@echo "📅 Создаем мероприятие с коротким сроком жизни (1 минута)..."
	@curl -s -X POST $(SERVER_URL)/events \
		-H "Authorization: Bearer $$(cat .admin_token)" \
		-H "Content-Type: application/json" \
		-d '{"name":"Short Expiry Event","date":"2025-12-25T20:00:00Z","total_seats":5,"life_span":"$(SHORT_LIFESPAN)"}' | \
		jq -r '.event.id' > .short_event_id
	@echo "Short Event ID: $$(cat .short_event_id)"
	@sleep 1
	
	# Бронируем место, но НЕ подтверждаем его
	@echo "🎟️ Бронируем место без подтверждения..."
	@curl -s -X POST $(SERVER_URL)/events/$$(cat .short_event_id)/book \
		-H "Authorization: Bearer $$(cat .user2_token)" \
		-H "Content-Type: application/json" \
		-d '{"seat_count":1}' | \
		jq -r '.booking.booking_code' > .expiry_booking_code
	@echo "Expiry Booking Code: $$(cat .expiry_booking_code)"
	@sleep 1
	
	# Ждем 70 секунд для истечения срока
	@echo "⏳ Ждем 70 секунд для истечения срока брони..."
	@sleep 70
	
	# Проверяем, что место освободилось
	@echo "🔍 Проверяем освободилось ли место..."
	@curl -s -X GET $(SERVER_URL)/events/$$(cat .short_event_id) | \
		jq -r '.event.available_seats' > .available_after_expiry
	@if [ "$$(cat .available_after_expiry)" -eq 5 ]; then \
		echo "✅ Тест истечения срока пройден!"; \
	else \
		echo "❌ Ошибка: бронь не была отменена автоматически"; \
		exit 1; \
	fi

test-concurrency:
	@echo "⚡ Тестирование конкурентного бронирования..."
	
	# Создаем мероприятие с ограниченным количеством мест
	@echo "📅 Создаем мероприятие с 3 местами..."
	@curl -s -X POST $(SERVER_URL)/events \
		-H "Authorization: Bearer $$(cat .admin_token)" \
		-H "Content-Type: application/json" \
		-d '{"name":"Concurrency Test Event","date":"2025-12-25T21:00:00Z","total_seats":3,"life_span":"5m"}' | \
		jq -r '.event.id' > .concurrent_event_id
	@echo "Concurrent Event ID: $$(cat .concurrent_event_id)"
	@sleep 1
	
	# Пытаемся забронировать больше мест, чем доступно
	@echo "🎟️ Пытаемся забронировать 4 места (больше чем доступно)..."
	@curl -s -X POST $(SERVER_URL)/events/$$(cat .concurrent_event_id)/book \
		-H "Authorization: Bearer $$(cat .user1_token)" \
		-H "Content-Type: application/json" \
		-d '{"seat_count":4}' | \
		jq -r '.error' > .error_response
	@if [ "$$(cat .error_response)" = "Not enough available seats" ]; then \
		echo "✅ Тест проверки доступности мест пройден!"; \
	else \
		echo "❌ Ошибка: система позволила забронировать больше мест чем доступно"; \
		exit 1; \
	fi
	@sleep 1
	
	# Тестируем конкурентное бронирование одних и тех же мест
	@echo "🎯 Тестируем конкурентное бронирование..."
	@curl -s -X POST $(SERVER_URL)/events/$$(cat .concurrent_event_id)/book \
		-H "Authorization: Bearer $$(cat .user1_token)" \
		-H "Content-Type: application/json" \
		-d '{"seat_count":2}' > /tmp/user1_booking.log &
	@sleep 0.5
	
	@curl -s -X POST $(SERVER_URL)/events/$$(cat .concurrent_event_id)/book \
		-H "Authorization: Bearer $$(cat .user2_token)" \
		-H "Content-Type: application/json" \
		-d '{"seat_count":2}' > /tmp/user2_booking.log &
	@sleep 0.5
	
	# Ждем завершения параллельных запросов
	@wait
	@sleep 1
	
	# Проверяем итоговое количество занятых мест
	@curl -s -X GET $(SERVER_URL)/events/$$(cat .concurrent_event_id) | \
		jq -r '.event.available_seats' > .final_seats
	@if [ "$$(cat .final_seats)" -ge 0 ] && [ "$$(cat .final_seats)" -le 3 ]; then \
		echo "✅ Тест конкурентности пройден!"; \
	else \
		echo "❌ Ошибка: проблемы с конкурентным доступом"; \
		exit 1; \
	fi

test_clean:
	@echo "🧹 Очистка временных файлов..."
	@rm -f .admin_token .user1_token .user2_token .event_id .booking1_code .booking2_code \
	       .admin_id .short_event_id .expiry_booking_code .available_seats \
	       .available_after_expiry .concurrent_event_id .error_response .final_seats \
	       /tmp/user1_booking.log /tmp/user2_booking.log 2>/dev/null || true

.PHONY: test setup test-booking test-expiry test-concurrency test_clean