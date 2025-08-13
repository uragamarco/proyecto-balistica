package api

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uragamarco/proyecto-balistica/internal/models"
	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
)

// MockImageProcessor para pruebas
type MockImageProcessor struct {
	mock.Mock
}

func (m *MockImageProcessor) Process(img image.Image) (image.Image, error) {
	args := m.Called(img)
	return args.Get(0).(image.Image), args.Error(1)
}

func (m *MockImageProcessor) ExtractFeatures(img image.Image, tempPath string) (map[string]float64, error) {
	args := m.Called(img, tempPath)
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockImageProcessor) PythonFeaturesStatus() (bool, string) {
	args := m.Called()
	return args.Bool(0), args.String(1)
}

// MockChromaService para pruebas
type MockChromaService struct {
	mock.Mock
}

func (m *MockChromaService) Analyze(img image.Image) (*chroma.AnalysisResult, error) {
	args := m.Called(img)
	return args.Get(0).(*chroma.AnalysisResult), args.Error(1)
}

func TestProcessImageHandler(t *testing.T) {
	// Configurar mocks
	mockImgProc := new(MockImageProcessor)
	mockChroma := new(MockChromaService)

	// Configurar valores esperados
	processedImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	features := map[string]float64{
		"hu_moment_1":  0.123,
		"contour_area": 5000.0,
	}
	chromaResult := &chroma.AnalysisResult{
		DominantColors: []chroma.ColorData{
			{Color: color.RGBA{R: 255, G: 255, B: 255, A: 255}, Frequency: 0.8},
		},
		ColorVariance: 0.1,
	}

	// Configurar expectativas
	mockImgProc.On("Process", mock.Anything).Return(processedImg, nil)
	mockImgProc.On("ExtractFeatures", processedImg, mock.Anything).Return(features, nil)
	mockImgProc.On("PythonFeaturesStatus").Return(true, "")
	mockChroma.On("Analyze", processedImg).Return(chromaResult, nil)

	// Crear handler
	handlers := NewHandlers(mockImgProc, mockChroma)

	// Crear router Gin para pruebas
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/process", handlers.ProcessImage)

	// Crear archivo temporal de imagen
	file, err := os.CreateTemp("", "test_image_*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	// Escribir una imagen PNG mínima
	_, err = file.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	// Preparar solicitud multipart
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", filepath.Base(file.Name()))
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	if err != nil {
		t.Fatal(err)
	}
	writer.Close()

	// Crear solicitud HTTP
	req, _ := http.NewRequest("POST", "/process", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Ejecutar solicitud
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Verificar respuesta
	assert.Equal(t, http.StatusOK, resp.Code)

	var analysis models.BallisticAnalysis
	err = json.Unmarshal(resp.Body.Bytes(), &analysis)
	assert.NoError(t, err)

	// Verificar campos básicos
	assert.Equal(t, 2, len(analysis.Features))
	assert.Equal(t, 1, len(analysis.ChromaData.DominantColors))
	assert.True(t, analysis.Metadata.PythonFeaturesUsed)
	assert.NotEmpty(t, analysis.Metadata.ImageHash)
	assert.Greater(t, analysis.Metadata.Confidence, 0.0)

	// Verificar llamadas a mocks
	mockImgProc.AssertExpectations(t)
	mockChroma.AssertExpectations(t)
}

func TestProcessImageHandler_InvalidMethod(t *testing.T) {
	mockImgProc := new(MockImageProcessor)
	mockChroma := new(MockChromaService)
	handlers := NewHandlers(mockImgProc, mockChroma)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/process", handlers.ProcessImage)

	req, _ := http.NewRequest("GET", "/process", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
}

func TestProcessImageHandler_NoImage(t *testing.T) {
	mockImgProc := new(MockImageProcessor)
	mockChroma := new(MockChromaService)
	handlers := NewHandlers(mockImgProc, mockChroma)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/process", handlers.ProcessImage)

	// Solicitud sin archivo de imagen
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.Close()

	req, _ := http.NewRequest("POST", "/process", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "Error retrieving image")
}

func TestCompareSamplesHandler(t *testing.T) {
	mockImgProc := new(MockImageProcessor)
	mockChroma := new(MockChromaService)
	handlers := NewHandlers(mockImgProc, mockChroma)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/compare", handlers.CompareSamples)

	comparisonRequest := struct {
		Sample1 map[string]float64 `json:"sample1"`
		Sample2 map[string]float64 `json:"sample2"`
	}{
		Sample1: map[string]float64{"feature1": 1.0, "feature2": 2.0},
		Sample2: map[string]float64{"feature1": 1.1, "feature2": 2.1},
	}

	body, _ := json.Marshal(comparisonRequest)
	req, _ := http.NewRequest("POST", "/compare", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)

	var result models.ComparisonResult
	err := json.Unmarshal(resp.Body.Bytes(), &result)
	assert.NoError(t, err)

	assert.Greater(t, result.Similarity, 0.8)
	assert.Less(t, result.Similarity, 1.0)
	assert.Len(t, result.DiffPerFeature, 2)
}

func TestHealthCheckHandler(t *testing.T) {
	mockImgProc := new(MockImageProcessor)
	mockChroma := new(MockChromaService)
	handlers := NewHandlers(mockImgProc, mockChroma)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "status")
}
