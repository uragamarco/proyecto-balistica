package image_processor

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"
	"github.com/uragamarco/proyecto-balistica/internal/services/python_features"
)

// ImageProcessor provee funcionalidades para procesamiento de imágenes balísticas
type ImageProcessor struct {
	config         *Config
	pythonFeatures *python_features.Service
}

// Config contiene parámetros de procesamiento de imágenes
type Config struct {
	Contrast               float64
	SharpenSigma           float64
	EdgeThreshold          int
	GLCMOffsetDistance     int         // Distancia para cálculo de GLCM (1-3 píxeles)
	ForegroundThreshold    uint8       // Umbral para detección de objeto (ej: 128)
	EdgeDetectionThreshold float64     // Sensibilidad para bordes
	TempDir                string      // Directorio para archivos temporales
	Logger                 *log.Logger // Logger para registrar eventos
}

// NewImageProcessor crea una nueva instancia del procesador de imágenes
func NewImageProcessor(cfg *Config, pyService *python_features.Service) *ImageProcessor {
	return &ImageProcessor{
		config:         cfg,
		pythonFeatures: pyService,
	}
}

// Process aplica transformaciones a la imagen para análisis balístico
func (ip *ImageProcessor) Process(img image.Image) (image.Image, error) {
	// 1. Convertir a escala de grises
	grayImg := imaging.Grayscale(img)

	// 2. Ajustar contraste
	contrastImg := imaging.AdjustContrast(grayImg, ip.config.Contrast)

	// 3. Enfocar imagen
	sharpenedImg := imaging.Sharpen(contrastImg, ip.config.SharpenSigma)

	// 4. Detección de bordes
	edges := ip.detectEdges(sharpenedImg)

	return edges, nil
}

// ExtractFeatures extrae características balísticas combinando métodos locales y Python
func (ip *ImageProcessor) ExtractFeatures(img image.Image, originalPath string) (map[string]float64, error) {
	features := make(map[string]float64)

	// 1. Características locales (Go)
	goFeatures, err := ip.extractLocalFeatures(img)
	if err != nil {
		return nil, err
	}

	for k, v := range goFeatures {
		features[k] = v
	}

	// 2. Características avanzadas (Python)
	if ip.pythonFeatures != nil {
		// Guardar imagen procesada temporalmente para Python
		tempPath := filepath.Join(ip.config.TempDir, "processed_"+time.Now().Format("20060102150405")+".png")
		if err := imaging.Save(img, tempPath); err != nil {
			return nil, fmt.Errorf("error guardando imagen temporal: %w", err)
		}
		defer os.Remove(tempPath)

		pyFeatures, err := ip.pythonFeatures.ExtractAdvancedFeatures(tempPath)
		if err != nil {
			return nil, fmt.Errorf("error en extracción Python: %w", err)
		}

		// Combinar características con nombres descriptivos
		for i, hu := range pyFeatures.HuMoments {
			features[fmt.Sprintf("hu_moment_%d", i+1)] = hu
		}

		for i, str := range pyFeatures.Striations {
			features[fmt.Sprintf("striation_%d", i+1)] = str
		}

		features["contour_area"] = pyFeatures.ContourArea
		features["contour_length"] = pyFeatures.ContourLen
	}

	return features, nil
}

// extractLocalFeatures extrae características usando solo métodos Go
func (ip *ImageProcessor) extractLocalFeatures(img image.Image) (map[string]float64, error) {
	features := make(map[string]float64)

	// 1. Características de textura (GLCM)
	glcmFeatures := ip.calculateGLCMFeatures(img)
	features["glcm_contrast"] = glcmFeatures[0]
	features["glcm_energy"] = glcmFeatures[1]
	features["glcm_homogeneity"] = glcmFeatures[2]

	// 2. Características de forma
	shapeFeatures := ip.calculateShapeFeatures(img)
	features["circularity"] = shapeFeatures[0]
	features["aspect_ratio"] = shapeFeatures[1]

	return features, nil
}

// detectEdges implementa detección de bordes usando operador Sobel
func (ip *ImageProcessor) detectEdges(img image.Image) image.Image {
	bounds := img.Bounds()
	edgeImg := image.NewGray(bounds)

	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			gx, gy := ip.sobelOperator(img, x, y)
			magnitude := math.Sqrt(float64(gx*gx + gy*gy))

			if magnitude > float64(ip.config.EdgeThreshold) {
				edgeImg.SetGray(x, y, color.Gray{Y: 255})
			} else {
				edgeImg.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}

	return edgeImg
}

// sobelOperator aplica el operador Sobel para detección de bordes
func (ip *ImageProcessor) sobelOperator(img image.Image, x, y int) (int, int) {
	var gx, gy int

	kernelX := [3][3]int{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}
	kernelY := [3][3]int{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}

	for ky := -1; ky <= 1; ky++ {
		for kx := -1; kx <= 1; kx++ {
			r, _, _, _ := img.At(x+kx, y+ky).RGBA()
			gray := int(r >> 8)
			gx += gray * kernelX[ky+1][kx+1]
			gy += gray * kernelY[ky+1][kx+1]
		}
	}

	return gx, gy
}

//-------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------

// getGrayValue convierte un pixel a valor de gris (0-255)
func getGrayValue(c color.Color) uint8 {
	r, g, b, _ := c.RGBA()
	// Fórmula estándar para conversión RGB a escala de grises
	gray := 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(b>>8)
	return uint8(gray)
}

// calculateContrast calcula el contraste desde la GLCM
func calculateContrast(glcm map[[2]uint8]int) float64 {
	var contrast float64
	for pair, count := range glcm {
		diff := int(pair[0]) - int(pair[1])
		contrast += float64(count) * float64(diff*diff)
	}
	return contrast / float64(sumGLCM(glcm))
}

// calculateEnergy calcula la energía desde la GLCM
func calculateEnergy(glcm map[[2]uint8]int) float64 {
	var energy float64
	total := sumGLCM(glcm)
	for _, count := range glcm {
		prob := float64(count) / float64(total)
		energy += prob * prob
	}
	return energy
}

// calculateHomogeneity calcula la homogeneidad desde la GLCM
func calculateHomogeneity(glcm map[[2]uint8]int) float64 {
	var homogeneity float64
	total := sumGLCM(glcm)
	for pair, count := range glcm {
		diff := int(pair[0]) - int(pair[1])
		homogeneity += float64(count) / (1.0 + float64(diff*diff))
	}
	return homogeneity / float64(total)
}

// sumGLCM suma todos los valores de la GLCM
func sumGLCM(glcm map[[2]uint8]int) int {
	total := 0
	for _, count := range glcm {
		total += count
	}
	return total
}

// calculateGLCMFeatures calcula características de textura usando matriz de co-ocurrencia
func (ip *ImageProcessor) calculateGLCMFeatures(img image.Image) []float64 {
	bounds := img.Bounds()
	glcm := make(map[[2]uint8]int)                            // Matriz de co-ocurrencia
	offsets := []image.Point{{0, 1}, {1, 0}, {1, 1}, {1, -1}} // Vecindarios

	// 1. Calcular GLCM
	for y := bounds.Min.Y; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X; x < bounds.Max.X-1; x++ {
			pixel1 := getGrayValue(img.At(x, y))
			for _, offset := range offsets {
				pixel2 := getGrayValue(img.At(x+offset.X, y+offset.Y))
				glcm[[2]uint8{pixel1, pixel2}]++ // Incrementar co-ocurrencias
			}
		}
	}

	// 2. Extraer características (ejemplo simplificado)
	contrast := calculateContrast(glcm)
	energy := calculateEnergy(glcm)
	homogeneity := calculateHomogeneity(glcm)

	return []float64{contrast, energy, homogeneity}
}

//-------------------------------------------------------------------------------------------------
//-------------------------------------------------------------------------------------------------

// isForeground determina si un pixel es parte del objeto balístico
func isForeground(c color.Color) bool {
	// Umbral para considerar un pixel como parte del objeto
	gray := getGrayValue(c)
	return gray < 128 // Ajustar según necesidad
}

// isEdgePixel determina si un pixel es borde del objeto
func isEdgePixel(img image.Image, x, y int) bool {
	if !isForeground(img.At(x, y)) {
		return false
	}

	// Verificar vecinos (4-connectivity)
	neighbors := []image.Point{
		{0, 1}, {1, 0}, {0, -1}, {-1, 0},
	}

	for _, n := range neighbors {
		if !isForeground(img.At(x+n.X, y+n.Y)) {
			return true
		}
	}
	return false
}

// calculateAspectRatio calcula la relación de aspecto del objeto
func calculateAspectRatio(img image.Image) float64 {
	bounds := img.Bounds()
	var (
		minX = bounds.Max.X
		maxX = bounds.Min.X
		minY = bounds.Max.Y
		maxY = bounds.Min.Y
	)

	// Encontrar los límites reales del objeto
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if isForeground(img.At(x, y)) {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	width := maxX - minX
	height := maxY - minY

	if height == 0 {
		return 0
	}
	return float64(width) / float64(height)
}

// calculateShapeFeatures extrae características geométricas relevantes
func (ip *ImageProcessor) calculateShapeFeatures(img image.Image) []float64 {
	bounds := img.Bounds()
	var area, perimeter float64

	// 1. Detección de contornos y cálculo de área/perímetro
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if isForeground(img.At(x, y)) {
				area++
				if isEdgePixel(img, x, y) {
					perimeter++
				}
			}
		}
	}

	// 2. Calcular métricas de forma
	circularity := (4 * math.Pi * area) / (perimeter * perimeter)
	aspectRatio := calculateAspectRatio(img)

	return []float64{circularity, aspectRatio}
}

func (ip *ImageProcessor) PythonFeaturesStatus() (bool, string) {
	if ip.pythonFeatures == nil {
		return false, "Python integration disabled"
	}

	err := ip.pythonFeatures.HealthCheck()
	if err != nil {
		return false, "Python error: " + err.Error()
	}

	return true, "Python integration active"
}
