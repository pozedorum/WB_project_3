# 1. Регистрация
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123","name":"Test User"}'

# 2. Логин (сохраните токен)
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@test.com","password":"test123"}' | jq -r '.token')

# 3. Создать мероприятие
curl -X POST http://localhost:8080/events \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Event","date":"2025-12-25T19:00:00Z","total_seats":50,"life_span":"1h"}'

# 4. Получить все мероприятия
curl -X GET http://localhost:8080/events

# 5. Забронировать место
curl -X POST http://localhost:8080/events/1/book \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"seat_count":2}'

# 6. Подтвердить бронь (используйте booking_code из ответа предыдущего запроса)
curl -X POST http://localhost:8080/events/1/confirm \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"booking_code":"ee52e89c-5570-49ad-b6d0-94e4ddcf4504"}'