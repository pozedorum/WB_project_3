
curl -X POST http://localhost:8080/notify \
  -H "Content-Type: application/json" \
  -d '{
    "author": "Arkadiy",
    "content": "Это первый комментарий"
  }'

curl -X POST http://localhost:8080/notify \
  -H "Content-Type: application/json" \
  -d '{
    "author": "Arkadiy",
    "content": "Это второй комментарий"
  }'

curl -X POST http://localhost:8080/notify \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "8aa5c2e8-607a-4483-97fb-1af7c286649e",
    "author": "Jane Smith", 
    "content": "Это ответ на первый комментарий"
  }'

curl -X POST http://localhost:8080/notify \
  -H "Content-Type: application/json" \
  -d '{
    "parent_id": "cceb9839-e09e-47a7-b288-3e49e145e1bc",
    "author": "Jane Smith", 
    "content": "Это ответ на второй комментарий"
  }'
  curl http://localhost:8080/notify
  curl http://localhost:8080/notify/all
  curl http://localhost:8080/notify/a5ef5b85-f3cd-4eaf-a98c-c6082a51ce21