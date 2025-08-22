package storage

import (
	"os"
	"testing"

	"github.com/uragamarco/proyecto-balistica/internal/models"
	"go.uber.org/zap"
)

func TestNewDatabase(t *testing.T) {
	// Crear un archivo de base de datos temporal
	tempFile := "test_ballistics.db"
	defer os.Remove(tempFile)

	logger := zap.NewNop()

	// Crear nueva base de datos
	db, err := NewDatabase(tempFile, logger)
	if err != nil {
		t.Fatalf("Error al crear base de datos: %v", err)
	}
	defer db.Close()

	// Verificar que la conexión funciona
	if err := db.db.Ping(); err != nil {
		t.Fatalf("Error al hacer ping a la base de datos: %v", err)
	}
}

func TestDatabaseTables(t *testing.T) {
	// Crear un archivo de base de datos temporal
	tempFile := "test_ballistics_tables.db"
	defer os.Remove(tempFile)

	logger := zap.NewNop()

	// Crear nueva base de datos
	db, err := NewDatabase(tempFile, logger)
	if err != nil {
		t.Fatalf("Error al crear base de datos: %v", err)
	}
	defer db.Close()

	// Verificar que las tablas existen
	tables := []string{"ballistic_analysis", "ballistic_comparisons", "ballistic_classifications"}

	for _, table := range tables {
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
		var name string
		err := db.db.QueryRow(query, table).Scan(&name)
		if err != nil {
			t.Fatalf("Tabla %s no existe: %v", table, err)
		}
		if name != table {
			t.Fatalf("Tabla %s no encontrada", table)
		}
	}
}

func TestAnalysisRepository(t *testing.T) {
	// Crear un archivo de base de datos temporal
	tempFile := "test_analysis_repo.db"
	defer os.Remove(tempFile)

	logger := zap.NewNop()

	// Crear nueva base de datos
	db, err := NewDatabase(tempFile, logger)
	if err != nil {
		t.Fatalf("Error al crear base de datos: %v", err)
	}
	defer db.Close()

	// Crear repositorio
	repo := NewAnalysisRepository(db, logger)

	// Datos de prueba
	imagePath := "/test/image.png"
	features := map[string]float64{
		"test_feature": 1.0,
		"numeric_feature": 123.45,
	}
	metadata := &models.AnalysisMetadata{
		ImageHash: "test_hash",
		Confidence: 0.95,
		PythonFeaturesUsed: true,
	}

	// Guardar análisis
	analysis, err := repo.SaveAnalysis(imagePath, features, metadata)
	if err != nil {
		t.Fatalf("Error al guardar análisis: %v", err)
	}

	// Verificar que se guardó correctamente
	if analysis.ID == "" {
		t.Fatal("ID de análisis vacío")
	}
	if analysis.ImagePath != imagePath {
		t.Fatalf("ImagePath incorrecto: esperado %s, obtenido %s", imagePath, analysis.ImagePath)
	}

	// Recuperar análisis
	retrievedAnalysis, err := repo.GetAnalysis(analysis.ID)
	if err != nil {
		t.Fatalf("Error al recuperar análisis: %v", err)
	}

	// Verificar datos recuperados
	if retrievedAnalysis.ID != analysis.ID {
		t.Fatalf("ID incorrecto: esperado %s, obtenido %s", analysis.ID, retrievedAnalysis.ID)
	}
	if retrievedAnalysis.ImagePath != imagePath {
		t.Fatalf("ImagePath incorrecto: esperado %s, obtenido %s", imagePath, retrievedAnalysis.ImagePath)
	}
}

func TestComparisonRepository(t *testing.T) {
	// Crear un archivo de base de datos temporal
	tempFile := "test_comparison_repo.db"
	defer os.Remove(tempFile)

	logger := zap.NewNop()

	// Crear nueva base de datos
	db, err := NewDatabase(tempFile, logger)
	if err != nil {
		t.Fatalf("Error al crear base de datos: %v", err)
	}
	defer db.Close()

	// Crear repositorios
	analysisRepo := NewAnalysisRepository(db, logger)
	comparisonRepo := NewComparisonRepository(db, logger)

	// Crear análisis de prueba
	features1 := map[string]float64{"feature1": 1.0}
	metadata1 := &models.AnalysisMetadata{ImageHash: "hash1", Confidence: 0.9, PythonFeaturesUsed: true}
	analysis1, err := analysisRepo.SaveAnalysis("/test/image1.jpg", features1, metadata1)
	if err != nil {
		t.Fatalf("Error al crear análisis 1: %v", err)
	}

	features2 := map[string]float64{"feature2": 2.0}
	metadata2 := &models.AnalysisMetadata{ImageHash: "hash2", Confidence: 0.8, PythonFeaturesUsed: true}
	analysis2, err := analysisRepo.SaveAnalysis("/test/image2.jpg", features2, metadata2)
	if err != nil {
		t.Fatalf("Error al crear análisis 2: %v", err)
	}

	// Datos de comparación
	similarityScore := 0.85
	confidence := 0.9
	matchResult := true
	comparisonData := map[string]interface{}{
		"method": "advanced",
		"confidence": 0.9,
	}

	// Guardar comparación
	comparison, err := comparisonRepo.SaveComparison(analysis1.ID, analysis2.ID, similarityScore, confidence, matchResult, comparisonData)
	if err != nil {
		t.Fatalf("Error al guardar comparación: %v", err)
	}

	// Verificar que se guardó correctamente
	if comparison.ID == "" {
		t.Fatal("ID de comparación vacío")
	}
	if comparison.Sample1ID != analysis1.ID {
		t.Fatalf("Sample1ID incorrecto: esperado %s, obtenido %s", analysis1.ID, comparison.Sample1ID)
	}
	if comparison.Sample2ID != analysis2.ID {
		t.Fatalf("Sample2ID incorrecto: esperado %s, obtenido %s", analysis2.ID, comparison.Sample2ID)
	}
	if comparison.Similarity != similarityScore {
		t.Fatalf("Similarity incorrecto: esperado %f, obtenido %f", similarityScore, comparison.Similarity)
	}

	// Recuperar comparación
	retrievedComparison, err := comparisonRepo.GetComparison(comparison.ID)
	if err != nil {
		t.Fatalf("Error al recuperar comparación: %v", err)
	}

	// Verificar datos recuperados
	if retrievedComparison.ID != comparison.ID {
		t.Fatalf("ID incorrecto: esperado %s, obtenido %s", comparison.ID, retrievedComparison.ID)
	}
}

func TestClassificationRepository(t *testing.T) {
	// Crear un archivo de base de datos temporal
	tempFile := "test_classification_repo.db"
	defer os.Remove(tempFile)

	logger := zap.NewNop()

	// Crear nueva base de datos
	db, err := NewDatabase(tempFile, logger)
	if err != nil {
		t.Fatalf("Error al crear base de datos: %v", err)
	}
	defer db.Close()

	// Crear repositorios
	analysisRepo := NewAnalysisRepository(db, logger)
	classificationRepo := NewClassificationRepository(db, logger)

	// Crear análisis de prueba
	features := map[string]float64{"feature": 1.0}
	metadata := &models.AnalysisMetadata{ImageHash: "test_hash", Confidence: 0.9, PythonFeaturesUsed: true}
	analysis, err := analysisRepo.SaveAnalysis("/test/image.jpg", features, metadata)
	if err != nil {
		t.Fatalf("Error al crear análisis: %v", err)
	}

	// Datos de clasificación
	weaponType := "pistol"
	caliber := "9mm"
	confidence := 0.92
	classificationData := map[string]interface{}{
		"method": "ml_classification",
		"features_used": []string{"firing_pin", "breech_face"},
	}

	// Guardar clasificación
	classification, err := classificationRepo.SaveClassification(analysis.ID, weaponType, caliber, confidence, classificationData)
	if err != nil {
		t.Fatalf("Error al guardar clasificación: %v", err)
	}

	// Verificar que se guardó correctamente
	if classification.ID == "" {
		t.Fatal("ID de clasificación vacío")
	}
	if classification.AnalysisID != analysis.ID {
		t.Fatalf("AnalysisID incorrecto: esperado %s, obtenido %s", analysis.ID, classification.AnalysisID)
	}
	if classification.WeaponType != weaponType {
		t.Fatalf("WeaponType incorrecto: esperado %s, obtenido %s", weaponType, classification.WeaponType)
	}
	if classification.Caliber != caliber {
		t.Fatalf("Caliber incorrecto: esperado %s, obtenido %s", caliber, classification.Caliber)
	}
	if classification.Confidence != confidence {
		t.Fatalf("Confidence incorrecto: esperado %f, obtenido %f", confidence, classification.Confidence)
	}

	// Recuperar clasificación
	retrievedClassification, err := classificationRepo.GetClassification(classification.ID)
	if err != nil {
		t.Fatalf("Error al recuperar clasificación: %v", err)
	}

	// Verificar datos recuperados
	if retrievedClassification.ID != classification.ID {
		t.Fatalf("ID incorrecto: esperado %s, obtenido %s", classification.ID, retrievedClassification.ID)
	}
}

func TestStorageService(t *testing.T) {
	// Crear un archivo de base de datos temporal
	tempFile := "test_storage_service.db"
	defer os.Remove(tempFile)

	logger := zap.NewNop()

	// Crear nueva base de datos
	db, err := NewDatabase(tempFile, logger)
	if err != nil {
		t.Fatalf("Error al crear base de datos: %v", err)
	}
	defer db.Close()

	// Crear servicio de almacenamiento
	storageService, err := NewStorageService(tempFile, logger)
	if err != nil {
		t.Fatalf("Error al crear servicio de almacenamiento: %v", err)
	}
	defer storageService.Close()

	// Verificar que el servicio se inicializó correctamente
	if storageService == nil {
		t.Fatal("StorageService es nil")
	}

	// Probar guardar análisis a través del servicio
	serviceFeatures := map[string]float64{"test": 1.0}
	serviceMetadata := &models.AnalysisMetadata{ImageHash: "service_hash", Confidence: 0.85, PythonFeaturesUsed: true}
	analysis, err := storageService.SaveAnalysis("/test/service_image.jpg", serviceFeatures, serviceMetadata)
	if err != nil {
		t.Fatalf("Error al guardar análisis a través del servicio: %v", err)
	}

	if analysis.ID == "" {
		t.Fatal("ID de análisis vacío")
	}

	// Recuperar análisis a través del servicio
	retrievedAnalysis, err := storageService.GetAnalysis(analysis.ID)
	if err != nil {
		t.Fatalf("Error al recuperar análisis a través del servicio: %v", err)
	}

	if retrievedAnalysis.ID != analysis.ID {
		t.Fatalf("ID incorrecto: esperado %s, obtenido %s", analysis.ID, retrievedAnalysis.ID)
	}

	// Probar estadísticas del dashboard
	stats, err := storageService.GetDashboardStats()
	if err != nil {
		t.Fatalf("Error al obtener estadísticas del dashboard: %v", err)
	}

	if stats == nil {
		t.Fatal("Estadísticas del dashboard son nil")
	}

	// Verificar que las estadísticas contienen las claves esperadas
	expectedKeys := []string{"total_analysis", "total_comparisons", "total_classifications"}
	for _, key := range expectedKeys {
		if _, exists := stats[key]; !exists {
			t.Fatalf("Clave esperada %s no encontrada en estadísticas", key)
		}
	}
}