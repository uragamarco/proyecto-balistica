package api

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uragamarco/proyecto-balistica/internal/models"
	"github.com/uragamarco/proyecto-balistica/internal/services/chroma"
	"github.com/uragamarco/proyecto-balistica/internal/services/classification"
	"github.com/uragamarco/proyecto-balistica/internal/services/image_processor"
	"github.com/uragamarco/proyecto-balistica/internal/storage"
	"go.uber.org/zap"
)

// MockImageProcessor para pruebas
type MockImageProcessor struct {
	mock.Mock
	*image_processor.ImageProcessor
}

func (m *MockImageProcessor) Process(img image.Image) (image.Image, error) {
	args := m.Called(img)
	return args.Get(0).(image.Image), args.Error(1)
}

func (m *MockImageProcessor) ExtractFeatures(img image.Image, tempPath string) (map[string]float64, map[string]interface{}, error) {
	args := m.Called(img, tempPath)
	return args.Get(0).(map[string]float64), args.Get(1).(map[string]interface{}), args.Error(2)
}

func (m *MockImageProcessor) PythonFeaturesStatus() (bool, string) {
	args := m.Called()
	return args.Bool(0), args.String(1)
}

// MockChromaService para pruebas
type MockChromaService struct {
	mock.Mock
	*chroma.Service
}

func (m *MockChromaService) Analyze(img image.Image) (*chroma.ChromaAnalysis, error) {
	args := m.Called(img)
	return args.Get(0).(*chroma.ChromaAnalysis), args.Error(1)
}

// MockStorageService mock para el servicio de almacenamiento
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) SaveAnalysis(imagePath string, features map[string]float64, metadata *models.AnalysisMetadata) (*storage.BallisticAnalysis, error) {
	args := m.Called(imagePath, features, metadata)
	return args.Get(0).(*storage.BallisticAnalysis), args.Error(1)
}

func (m *MockStorageService) GetAnalysis(id string) (*storage.BallisticAnalysis, error) {
	args := m.Called(id)
	return args.Get(0).(*storage.BallisticAnalysis), args.Error(1)
}

func (m *MockStorageService) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestProcessImageHandler(t *testing.T) {
	// Crear handler
	logger, _ := zap.NewDevelopment()

	// Configurar mocks
	mockImgProc := &MockImageProcessor{
		ImageProcessor: image_processor.NewImageProcessor(&image_processor.Config{
			Logger: logger,
		}, nil),
	}
	mockChroma := &MockChromaService{
		Service: chroma.NewService(&chroma.Config{
			SampleSize:     100,
			ColorThreshold: 0.05,
		}),
	}

	// Configurar valores esperados
	processedImg := image.NewRGBA(image.Rect(0, 0, 100, 100))
	features := map[string]float64{
		"hu_moment_1":  0.123,
		"contour_area": 5000.0,
	}
	chromaResult := &chroma.ChromaAnalysis{
		DominantColors: []chroma.ColorData{
			{Color: color.RGBA{R: 255, G: 255, B: 255, A: 255}, Frequency: 0.8},
		},
		ColorVariance: map[string]float64{"overall": 0.1},
	}

	// Configurar expectativas
	mockImgProc.On("Process", mock.Anything).Return(processedImg, nil)

	// Metadatos de prueba
	metadata := map[string]interface{}{
		"filename":     "test_image.png",
		"content_type": "image/png",
		"file_size":    int64(1024),
	}

	mockImgProc.On("ExtractFeatures", processedImg, mock.Anything).Return(features, metadata, nil)
	mockImgProc.On("PythonFeaturesStatus").Return(true, "")
	mockChroma.On("Analyze", processedImg).Return(chromaResult, nil)
	mockStorage := &MockStorageService{}
	// Configurar mock para que no falle
	mockStorage.On("SaveAnalysis", mock.Anything, mock.Anything, mock.Anything).Return(&storage.BallisticAnalysis{}, nil)
	// Mock classification service
	mockClassification := &classification.ClassificationService{}
	handlers := NewHandlers(logger, mockImgProc, mockChroma, &storage.StorageService{}, mockClassification)

	// Crear router Gin para pruebas
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/process", func(c *gin.Context) {
		handlers.ProcessImage(c.Writer, c.Request)
	})


	// Crear una imagen PNG válida en memoria
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	// Llenar con color blanco
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}

	// Codificar como PNG
	var imgBuffer bytes.Buffer
	err := png.Encode(&imgBuffer, img)
	if err != nil {
		t.Fatal(err)
	}

	// Preparar solicitud multipart
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	// Crear header personalizado para especificar content-type
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{`form-data; name="image"; filename="test_image.png"`}
	h["Content-Type"] = []string{"image/png"}
	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	_, err = part.Write(imgBuffer.Bytes())
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
	logger, _ := zap.NewDevelopment()
	mockImgProc := &MockImageProcessor{
		ImageProcessor: image_processor.NewImageProcessor(&image_processor.Config{
			Logger: logger,
		}, nil),
	}
	mockChroma := &MockChromaService{
		Service: chroma.NewService(&chroma.Config{
			SampleSize:     100,
			ColorThreshold: 0.05,
		}),
	}
	// Mock classification service
	mockClassification := &classification.ClassificationService{}
	handlers := NewHandlers(logger, mockImgProc, mockChroma, &storage.StorageService{}, mockClassification)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/process", func(c *gin.Context) {
		handlers.ProcessImage(c.Writer, c.Request)
	})
	// Agregar handler GET para devolver 405
	router.GET("/process", func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
	})

	req, _ := http.NewRequest("GET", "/process", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
}

func TestProcessImageHandler_NoImage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockImgProc := &MockImageProcessor{
		ImageProcessor: image_processor.NewImageProcessor(&image_processor.Config{
			Logger: logger,
		}, nil),
	}
	mockChroma := &MockChromaService{
		Service: chroma.NewService(&chroma.Config{
			SampleSize:     100,
			ColorThreshold: 0.05,
		}),
	}
	// Mock classification service
	mockClassification := &classification.ClassificationService{}
	handlers := NewHandlers(logger, mockImgProc, mockChroma, &storage.StorageService{}, mockClassification)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/process", func(c *gin.Context) {
		handlers.ProcessImage(c.Writer, c.Request)
	})

	// Solicitud sin archivo de imagen
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.Close()

	req, _ := http.NewRequest("POST", "/process", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "Error al obtener la imagen")
}

func TestCompareSamplesHandler(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockImgProc := &MockImageProcessor{
		ImageProcessor: image_processor.NewImageProcessor(&image_processor.Config{
			Logger: logger,
		}, nil),
	}
	mockChroma := &MockChromaService{
		Service: chroma.NewService(&chroma.Config{
			SampleSize:     100,
			ColorThreshold: 0.05,
		}),
	}
	// Mock classification service
	mockClassification := &classification.ClassificationService{}
	handlers := NewHandlers(logger, mockImgProc, mockChroma, &storage.StorageService{}, mockClassification)

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/compare", func(c *gin.Context) {
		handlers.CompareSamples(c.Writer, c.Request)
	})

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
