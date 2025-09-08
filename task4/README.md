curl -X POST http://localhost:8080/upload -F "image=@./test/cat3.jpeg" -F "width=1000" -F "height=300" -F "watermark=MyCat"

curl -X GET http://localhost:8080/image/1757331332656050490_fa0806f3.jpeg --output ./test/result3.jpeg
