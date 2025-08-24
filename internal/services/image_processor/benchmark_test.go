package image_processor

import (
	"fmt"
	"image"
	"image/color"
	"testing"
	"time"

	"go.uber.org/zap"
)

// createTestImage crea una imagen de prueba para benchmarks
func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// Crear patrón de prueba con círculos y líneas
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Crear patrón circular
			centerX, centerY := width/2, height/2
			distance := float64((x-centerX)*(x-centerX) + (y-centerY)*(y-centerY))
			radius := float64(width / 4)
			
			if distance < radius*radius {
				// Interior del círculo - objeto
				intensity := uint8(200 + (x+y)%55) // Variación de textura
				img.Set(x, y, color.RGBA{intensity, intensity, intensity, 255})
			} else {
				// Fondo
				intensity := uint8(50 + (x*y)%50)
				img.Set(x, y, color.RGBA{intensity, intensity, intensity, 255})
			}
		}
	}
	
	return img
}

// BenchmarkOriginalProcessor benchmark del procesador original
func BenchmarkOriginalProcessor(b *testing.B) {
	config := &Config{
		Contrast:               1.2,
		SharpenSigma:           1.0,
		EdgeThreshold:          50,
		GLCMOffsetDistance:     1,
		ForegroundThreshold:    128,
		EdgeDetectionThreshold: 0.1,
		TempDir:                "/tmp",
		Logger:                 zap.NewNop(),
	}
	
	processor := NewImageProcessor(config, nil)
	img := createTestImage(800, 600)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, _, err := processor.ExtractFeatures(img, "test_image.jpg")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkOptimizedProcessor benchmark del procesador optimizado
func BenchmarkOptimizedProcessor(b *testing.B) {
	config := &Config{
		Contrast:               1.2,
		SharpenSigma:           1.0,
		EdgeThreshold:          50,
		GLCMOffsetDistance:     1,
		ForegroundThreshold:    128,
		EdgeDetectionThreshold: 0.1,
		TempDir:                "/tmp",
		Logger:                 zap.NewNop(),
	}
	
	processor := NewOptimizedImageProcessor(config, nil)
	defer processor.Cleanup()
	
	img := createTestImage(800, 600)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, _, err := processor.ExtractFeaturesOptimized(img, "test_image.jpg")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkOriginalProcessorLargeImage benchmark con imagen grande
func BenchmarkOriginalProcessorLargeImage(b *testing.B) {
	config := &Config{
		Contrast:               1.2,
		SharpenSigma:           1.0,
		EdgeThreshold:          50,
		GLCMOffsetDistance:     1,
		ForegroundThreshold:    128,
		EdgeDetectionThreshold: 0.1,
		TempDir:                "/tmp",
		Logger:                 zap.NewNop(),
	}
	
	processor := NewImageProcessor(config, nil)
	img := createTestImage(2048, 1536) // Imagen grande
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, _, err := processor.ExtractFeatures(img, "large_test_image.jpg")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkOptimizedProcessorLargeImage benchmark optimizado con imagen grande
func BenchmarkOptimizedProcessorLargeImage(b *testing.B) {
	config := &Config{
		Contrast:               1.2,
		SharpenSigma:           1.0,
		EdgeThreshold:          50,
		GLCMOffsetDistance:     1,
		ForegroundThreshold:    128,
		EdgeDetectionThreshold: 0.1,
		TempDir:                "/tmp",
		Logger:                 zap.NewNop(),
	}
	
	processor := NewOptimizedImageProcessor(config, nil)
	defer processor.Cleanup()
	
	img := createTestImage(2048, 1536) // Imagen grande
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_, _, err := processor.ExtractFeaturesOptimized(img, "large_test_image.jpg")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCachePerformance benchmark del sistema de cache
func BenchmarkCachePerformance(b *testing.B) {
	config := &Config{
		Contrast:               1.2,
		SharpenSigma:           1.0,
		EdgeThreshold:          50,
		GLCMOffsetDistance:     1,
		ForegroundThreshold:    128,
		EdgeDetectionThreshold: 0.1,
		TempDir:                "/tmp",
		Logger:                 zap.NewNop(),
	}
	
	processor := NewOptimizedImageProcessor(config, nil)
	defer processor.Cleanup()
	
	img := createTestImage(800, 600)
	
	// Primera ejecución para llenar cache
	_, _, err := processor.ExtractFeaturesOptimized(img, "cached_test_image.jpg")
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// Benchmarks con cache hit
	for i := 0; i < b.N; i++ {
		_, _, err := processor.ExtractFeaturesOptimized(img, "cached_test_image.jpg")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParallelProcessing benchmark del procesamiento paralelo
func BenchmarkParallelProcessing(b *testing.B) {
	config := &Config{
		Contrast:               1.2,
		SharpenSigma:           1.0,
		EdgeThreshold:          50,
		GLCMOffsetDistance:     1,
		ForegroundThreshold:    128,
		EdgeDetectionThreshold: 0.1,
		TempDir:                "/tmp",
		Logger:                 zap.NewNop(),
	}
	
	processor := NewOptimizedImageProcessor(config, nil)
	defer processor.Cleanup()
	
	// Crear múltiples imágenes para procesamiento concurrente
	images := make([]image.Image, 10)
	for i := range images {
		images[i] = createTestImage(400, 300)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// Procesar múltiples imágenes en paralelo
		for j, img := range images {
			go func(idx int, image image.Image) {
				_, _, err := processor.ExtractFeaturesOptimized(image, fmt.Sprintf("parallel_test_%d.jpg", idx))
				if err != nil {
					b.Error(err)
				}
			}(j, img)
		}
		
		// Esperar un poco para que terminen
		time.Sleep(100 * time.Millisecond)
	}
}

// TestOptimizedProcessorAccuracy test de precisión del procesador optimizado
func TestOptimizedProcessorAccuracy(t *testing.T) {
	config := &Config{
		Contrast:               1.2,
		SharpenSigma:           1.0,
		EdgeThreshold:          50,
		GLCMOffsetDistance:     1,
		ForegroundThreshold:    128,
		EdgeDetectionThreshold: 0.1,
		TempDir:                "/tmp",
		Logger:                 zap.NewNop(),
	}
	
	original := NewImageProcessor(config, nil)
	optimized := NewOptimizedImageProcessor(config, nil)
	defer optimized.Cleanup()
	
	img := createTestImage(400, 300)
	
	// Extraer características con ambos procesadores
	originalFeatures, _, err1 := original.ExtractFeatures(img, "accuracy_test.jpg")
	if err1 != nil {
		t.Fatal(err1)
	}
	
	optimizedFeatures, _, err2 := optimized.ExtractFeaturesOptimized(img, "accuracy_test.jpg")
	if err2 != nil {
		t.Fatal(err2)
	}
	
	// Comparar características clave (permitir pequeñas diferencias por paralelización)
	tolerance := 0.1
	keyFeatures := []string{"glcm_contrast", "glcm_energy", "glcm_homogeneity", "circularity", "aspect_ratio"}
	
	for _, feature := range keyFeatures {
		origVal, origExists := originalFeatures[feature]
		optVal, optExists := optimizedFeatures[feature]
		
		if origExists != optExists {
			t.Errorf("Feature %s existence mismatch: original=%v, optimized=%v", feature, origExists, optExists)
			continue
		}
		
		if origExists {
			diff := abs(origVal - optVal)
			if diff > tolerance {
				t.Errorf("Feature %s value difference too large: original=%.4f, optimized=%.4f, diff=%.4f", 
					feature, origVal, optVal, diff)
			}
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}