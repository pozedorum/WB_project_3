package processor

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"testing"

	"golang.org/x/image/draw"
)

// Создаем тестовое изображение
func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / width),
				G: uint8(y * 255 / height),
				B: uint8((x + y) * 255 / (width + height)),
				A: 255,
			})
		}
	}
	return img
}

func TestResize(t *testing.T) {
	processor := NewDefaultProcessor()
	testImg := createTestImage(100, 100)

	// Тест уменьшения
	resized, err := processor.Resize(testImg, 50, 50)
	if err != nil {
		t.Fatalf("Resize failed: %v", err)
	}
	if resized.Bounds().Dx() != 50 || resized.Bounds().Dy() != 50 {
		t.Errorf("Expected 50x50, got %dx%d", resized.Bounds().Dx(), resized.Bounds().Dy())
	}

	// Тест увеличения
	resized, err = processor.Resize(testImg, 200, 200)
	if err != nil {
		t.Fatalf("Resize failed: %v", err)
	}
	if resized.Bounds().Dx() != 200 || resized.Bounds().Dy() != 200 {
		t.Errorf("Expected 200x200, got %dx%d", resized.Bounds().Dx(), resized.Bounds().Dy())
	}
}

func TestCreateThumbnail(t *testing.T) {
	processor := NewDefaultProcessor()

	// Горизонтальное изображение
	horizontalImg := createTestImage(200, 100)
	thumbnail, err := processor.CreateThumbnail(horizontalImg, 150)
	if err != nil {
		t.Fatalf("CreateThumbnail failed: %v", err)
	}

	bounds := thumbnail.Bounds()
	if bounds.Dx() != 150 || bounds.Dy() != 75 { // 200:100 = 150:75
		t.Errorf("Expected 150x75, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Вертикальное изображение
	verticalImg := createTestImage(100, 200)
	thumbnail, err = processor.CreateThumbnail(verticalImg, 150)
	if err != nil {
		t.Fatalf("CreateThumbnail failed: %v", err)
	}

	bounds = thumbnail.Bounds()
	if bounds.Dx() != 75 || bounds.Dy() != 150 { // 100:200 = 75:150
		t.Errorf("Expected 75x150, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestAddWatermark(t *testing.T) {
	processor := NewDefaultProcessor()
	testImg := createTestImage(100, 100)

	watermarked, err := processor.AddWatermark(testImg, "Test")
	if err != nil {
		t.Fatalf("AddWatermark failed: %v", err)
	}

	if watermarked.Bounds() != testImg.Bounds() {
		t.Error("Watermark changed image dimensions")
	}
}

func TestConvertFormat(t *testing.T) {
	processor := NewDefaultProcessor()
	testImg := createTestImage(10, 10)

	// Тест JPEG
	jpegData, err := processor.ConvertFormat(testImg, "jpeg", 90)
	if err != nil {
		t.Fatalf("JPEG conversion failed: %v", err)
	}
	if len(jpegData) == 0 {
		t.Error("JPEG data is empty")
	}

	// Тест PNG
	pngData, err := processor.ConvertFormat(testImg, "png", 0)
	if err != nil {
		t.Fatalf("PNG conversion failed: %v", err)
	}
	if len(pngData) == 0 {
		t.Error("PNG data is empty")
	}

	// Тест неподдерживаемого формата
	_, err = processor.ConvertFormat(testImg, "bmp", 90)
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
}

func TestProcessImageIntegration(t *testing.T) {
	processor := NewDefaultProcessor()
	testImg := createTestImage(100, 100)

	// Конвертируем в bytes для ProcessImage
	var buf bytes.Buffer
	err := png.Encode(&buf, testImg)
	if err != nil {
		t.Fatal(err)
	}
	imageData := buf.Bytes()

	// Тест полной обработки
	opts := ProcessingOptions{
		Width:         50,
		Height:        50,
		Quality:       90,
		Format:        "jpeg",
		WatermarkText: "Test",
		Thumbnail:     false,
	}

	result, err := processor.ProcessImage(imageData, opts)
	if err != nil {
		t.Fatalf("ProcessImage failed: %v", err)
	}

	if result.Width != 50 || result.Height != 50 {
		t.Errorf("Expected 50x50, got %dx%d", result.Width, result.Height)
	}

	if result.Size == 0 {
		t.Error("Result size is zero")
	}

	if result.Format != "jpeg" {
		t.Errorf("Expected jpeg, got %s", result.Format)
	}
}

func TestZeroSize(t *testing.T) {
	processor := NewDefaultProcessor()
	testImg := createTestImage(0, 0)

	if _, err := processor.Resize(testImg, 50, 50); err != ErrWrongBounds {
		t.Errorf("Expected error %v, got %v", ErrWrongBounds, err)
	}
	if _, err := processor.CreateThumbnail(testImg, 10); err != ErrWrongBounds {
		t.Errorf("Expected error %v, got %v", ErrWrongBounds, err)
	}
	if _, err := processor.AddWatermark(testImg, "aboba"); err != ErrWrongBounds {
		t.Errorf("Expected error %v, got %v", ErrWrongBounds, err)
	}
	if _, err := processor.ConvertFormat(testImg, "png", DefaultQuality); err != ErrWrongBounds {
		t.Errorf("Expected error %v, got %v", ErrWrongBounds, err)
	}
}

// encodeImageToFormat кодирует изображение в указанный формат
func encodeImageToFormat(img image.Image, format string, quality int) ([]byte, error) {
	var buf bytes.Buffer
	var err error

	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(&buf, img)
	case "gif":
		paletted := image.NewPaletted(img.Bounds(), palette.Plan9)
		draw.Draw(paletted, paletted.Bounds(), img, image.Point{}, draw.Src)
		err = gif.Encode(&buf, paletted, &gif.Options{})
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestProcessImageDifferentFormats(t *testing.T) {
	processor := NewDefaultProcessor()
	testImg := createTestImage(100, 100)

	testCases := []struct {
		name           string
		inputFormat    string
		outputFormat   string
		expectedFormat string
		quality        int
	}{
		{"JPEG to JPEG", "jpeg", "jpeg", "jpeg", 90},
		{"JPEG to PNG", "jpeg", "png", "png", 0},
		{"JPEG to GIF", "jpeg", "gif", "gif", 0},
		{"PNG to JPEG", "png", "jpeg", "jpeg", 85},
		{"PNG to PNG", "png", "png", "png", 0},
		{"PNG to GIF", "png", "gif", "gif", 0},
		{"GIF to JPEG", "gif", "jpeg", "jpeg", 75},
		{"GIF to PNG", "gif", "png", "png", 0},
		{"GIF to GIF", "gif", "gif", "gif", 0},
		{"Auto format (JPEG)", "jpeg", "", "jpeg", 0},
		{"Auto format (PNG)", "png", "", "png", 0},
		{"Auto format (GIF)", "gif", "", "gif", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Кодируем в исходный формат
			inputData, err := encodeImageToFormat(testImg, tc.inputFormat, 90)
			if err != nil {
				t.Fatalf("Failed to encode %s: %v", tc.inputFormat, err)
			}

			// Обрабатываем
			opts := ProcessingOptions{
				Width:   80,
				Height:  60,
				Quality: tc.quality,
				Format:  tc.outputFormat,
			}

			result, err := processor.ProcessImage(inputData, opts)
			if err != nil {
				t.Fatalf("ProcessImage failed for %s -> %s: %v",
					tc.inputFormat, tc.outputFormat, err)
			}

			// Проверяем результат
			if result.Format != tc.expectedFormat {
				t.Errorf("Expected format %s, got %s",
					tc.expectedFormat, result.Format)
			}

			if result.Width != 80 || result.Height != 60 {
				t.Errorf("Expected size 80x60, got %dx%d",
					result.Width, result.Height)
			}

			if result.Size == 0 {
				t.Error("Result size is zero")
			}

			// Проверяем, что данные можно декодировать
			_, format, err := image.Decode(bytes.NewReader(result.ProcessedData))
			if err != nil {
				t.Errorf("Failed to decode result image: %v", err)
			}

			if format != tc.expectedFormat {
				t.Errorf("Decoded format %s doesn't match expected %s",
					format, tc.expectedFormat)
			}
		})
	}
}

func TestProcessImageWithWatermarkDifferentFormats(t *testing.T) {
	processor := NewDefaultProcessor()
	testImg := createTestImage(200, 200)

	formats := []string{"jpeg", "png", "gif"}

	for _, format := range formats {
		t.Run("Watermark with "+format, func(t *testing.T) {
			inputData, err := encodeImageToFormat(testImg, format, 90)
			if err != nil {
				t.Fatalf("Failed to encode %s: %v", format, err)
			}

			opts := ProcessingOptions{
				Width:         150,
				Height:        150,
				Quality:       90,
				Format:        format,
				WatermarkText: "Test Watermark " + format,
				Thumbnail:     false,
			}

			result, err := processor.ProcessImage(inputData, opts)
			if err != nil {
				t.Fatalf("ProcessImage failed for %s with watermark: %v", format, err)
			}

			if result.Format != format {
				t.Errorf("Expected format %s, got %s", format, result.Format)
			}

			if result.Width != 150 || result.Height != 150 {
				t.Errorf("Expected size 150x150, got %dx%d", result.Width, result.Height)
			}
		})
	}
}

func TestProcessImageThumbnailDifferentFormats(t *testing.T) {
	processor := NewDefaultProcessor()
	testImg := createTestImage(300, 200) // Горизонтальное изображение

	formats := []string{"jpeg", "png", "gif"}

	for _, format := range formats {
		t.Run("Thumbnail with "+format, func(t *testing.T) {
			inputData, err := encodeImageToFormat(testImg, format, 90)
			if err != nil {
				t.Fatalf("Failed to encode %s: %v", format, err)
			}

			opts := ProcessingOptions{
				Thumbnail: true,
				Format:    format,
				Quality:   90,
			}

			result, err := processor.ProcessImage(inputData, opts)
			if err != nil {
				t.Fatalf("ProcessImage failed for %s thumbnail: %v", format, err)
			}

			// Проверяем, что миниатюра сохранила пропорции
			expectedRatio := float64(300) / float64(200) // 1.5
			actualRatio := float64(result.Width) / float64(result.Height)

			if math.Abs(actualRatio-expectedRatio) > 0.1 {
				t.Errorf("Thumbnail ratio mismatch: expected ~%.2f, got %.2f",
					expectedRatio, actualRatio)
			}

			if result.Width > 200 || result.Height > 200 {
				t.Errorf("Thumbnail too large: %dx%d", result.Width, result.Height)
			}
		})
	}
}

func TestProcessImageQualitySettings(t *testing.T) {
	processor := NewDefaultProcessor()
	testImg := createTestImage(100, 100)

	inputData, err := encodeImageToFormat(testImg, "jpeg", 90)
	if err != nil {
		t.Fatal(err)
	}

	qualityLevels := []int{10, 50, 90, 100}
	var sizes []int64

	for _, quality := range qualityLevels {
		t.Run(fmt.Sprintf("Quality %d", quality), func(t *testing.T) {
			opts := ProcessingOptions{
				Format:  "jpeg",
				Quality: quality,
			}

			result, err := processor.ProcessImage(inputData, opts)
			if err != nil {
				t.Fatalf("ProcessImage failed for quality %d: %v", quality, err)
			}

			sizes = append(sizes, result.Size)

			// Проверяем, что изображение валидно
			_, err = jpeg.Decode(bytes.NewReader(result.ProcessedData))
			if err != nil {
				t.Errorf("Invalid JPEG with quality %d: %v", quality, err)
			}
		})
	}

	// Проверяем, что размер файла уменьшается с понижением качества
	for i := 1; i < len(sizes); i++ {
		if qualityLevels[i] < qualityLevels[i-1] && sizes[i] >= sizes[i-1] {
			t.Errorf("Expected smaller file size for lower quality: %d -> %d bytes",
				sizes[i-1], sizes[i])
		}
	}
}

func TestProcessImageUnsupportedFormat(t *testing.T) {
	processor := NewDefaultProcessor()

	// Создаем данные в неподдерживаемом формате (просто случайные байты)
	invalidData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10} // Неполный JPEG

	opts := ProcessingOptions{
		Format: "jpeg",
	}

	_, err := processor.ProcessImage(invalidData, opts)
	if err == nil {
		t.Error("Expected error for invalid image data")
	}
}

func TestProcessImageEmptyData(t *testing.T) {
	processor := NewDefaultProcessor()

	opts := ProcessingOptions{
		Format: "jpeg",
	}

	_, err := processor.ProcessImage([]byte{}, opts)
	if err == nil {
		t.Error("Expected error for empty image data")
	}
}

func BenchmarkProcessImageJPEG(b *testing.B) {
	processor := NewDefaultProcessor()
	testImg := createTestImage(1024, 768)

	inputData, err := encodeImageToFormat(testImg, "jpeg", 90)
	if err != nil {
		b.Fatal(err)
	}

	opts := ProcessingOptions{
		Width:   800,
		Height:  600,
		Quality: 90,
		Format:  "jpeg",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.ProcessImage(inputData, opts)
		if err != nil {
			b.Fatalf("ProcessImage failed: %v", err)
		}
	}
}

func BenchmarkProcessImagePNG(b *testing.B) {
	processor := NewDefaultProcessor()
	testImg := createTestImage(1024, 768)

	inputData, err := encodeImageToFormat(testImg, "png", 0)
	if err != nil {
		b.Fatal(err)
	}

	opts := ProcessingOptions{
		Width:  800,
		Height: 600,
		Format: "png",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.ProcessImage(inputData, opts)
		if err != nil {
			b.Fatalf("ProcessImage failed: %v", err)
		}
	}
}
