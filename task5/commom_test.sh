#!/bin/bash
# 1. Регистрация
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test2@test.com","password":"test123","name":"Test User2"}'

# 2. Логин (сохраните токен)
LOGIN_TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test2@test.com","password":"test123"}' | jq -r '.token')
echo -e "\nLogin token: $LOGIN_TOKEN\n"

curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"biba@yandex.ru","password":"123456"}' | jq -r '.token'


# 3. Создать мероприятие
curl -X POST http://localhost:8080/events \
  -H "Authorization: Bearer $LOGIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"EVENT FOR CANCELING","date":"2025-12-25T19:00:00Z","total_seats":50,"life_span":"1m"}'

# 4. Получить все мероприятия
curl -X GET http://localhost:8080/events

# 5. Забронировать место
# Сохраняем booking_code из ответа
BOOKING_CODE=$(curl -s -X POST http://localhost:8080/events/3/book \
  -H "Authorization: Bearer $LOGIN_TOKEN2" \
  -H "Content-Type: application/json" \
  -d '{"seat_count":2}' | jq -r '.booking.booking_code')

echo -e "\nBooking code: $BOOKING_CODE\n"

# # Подтверждаем бронь (правильная подстановка переменной)
curl -X POST http://localhost:8080/events/1/confirm \
  -H "Authorization: Bearer $LOGIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"booking_code\": \"$BOOKING_CODE\"}"
