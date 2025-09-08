#!/bin/bash

# Скрипт для тестирования обработки изображений с разными параметрами
BASE_URL="http://localhost:8080"
TEST_IMAGE_PATH="./test/cat3.jpeg"

echo "Starting image processing tests..."

# Функция для загрузки изображения с параметрами
upload_image() {
    local description="$1"
    shift
    local params=("$@")
    
    echo "$description"
    echo "Parameters: $*"
    
    # Создаем временный файл для параметров curl
    local temp_file=$(mktemp)
    
    # Формируем команду curl
    local curl_cmd="curl -s -X POST \"$BASE_URL/upload\" -F \"image=@$TEST_IMAGE_PATH\""
    
    for param in "${params[@]}"; do
        curl_cmd="$curl_cmd -F \"$param\""
    done
    
    curl_cmd="$curl_cmd -w \"HTTP_STATUS:%{http_code}\""
    
    # Выполняем команду
    response=$(bash -c "$curl_cmd")
    
    # Извлекаем тело ответа и статус
    body=$(echo "$response" | sed -e 's/HTTP_STATUS\:.*//g')
    status=$(echo "$response" | tr -d '\n' | sed -e 's/.*HTTP_STATUS://')
    
    if [ "$status" -eq 202 ]; then
        image_id=$(echo "$body" | jq -r '.image_id')
        echo "✓ Success! Image ID: $image_id"
        echo "$image_id" >> test_image_ids.txt
    else
        echo "✗ Failed! Status: $status"
        echo "Response: $body"
    fi
    echo "----------------------------------------"
    
    rm -f "$temp_file"
}

# Очищаем файл с ID тестовых изображений
> test_image_ids.txt

echo "=== Test 1: Resize only ==="
upload_image "Resize to 300x200" "width=300" "height=200"

echo "=== Test 2: Resize with watermark ==="
upload_image "Resize with watermark" "width=500" "height=300" "watermark=TestWatermark"

echo "=== Test 3: Thumbnail ==="
upload_image "Create thumbnail" "thumbnail=true"

echo "=== Test 4: Format conversion to PNG ==="
upload_image "Convert to PNG" "format=png" "quality=90"

echo "=== Test 5: Complex processing ==="
upload_image "Complex processing" "width=800" "height=600" "quality=95" "watermark=ComplexTest" "format=jpeg"

echo "=== Test 6: Original image (no processing) ==="
upload_image "Original image" "dummy=1"  # Пустой параметр чтобы не было ошибок

echo ""
echo "=== Waiting for processing to complete... ==="
sleep 5

echo ""
echo "=== Checking processing status ==="
while read image_id; do
    if [ -n "$image_id" ]; then
        status_response=$(curl -s "$BASE_URL/image/$image_id")
        if echo "$status_response" | jq -e '.status' >/dev/null 2>&1; then
            status=$(echo "$status_response" | jq -r '.status')
            case "$status" in
                "completed") echo "✓ $image_id: COMPLETED" ;;
                "processing") echo "⏳ $image_id: PROCESSING" ;;
                *) echo "? $image_id: $status" ;;
            esac
        else
            echo "✓ $image_id: COMPLETED (binary data received)"
        fi
    fi
done < test_image_ids.txt

echo ""
echo "=== Test completed ==="
echo "Image IDs saved to test_image_ids.txt"