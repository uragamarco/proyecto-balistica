package storage

import (
	"fmt"
	"time"

	"github.com/uragamarco/proyecto-balistica/internal/models"
	"go.uber.org/zap"
)

// StorageService proporciona una interfaz unificada para todas las operaciones de almacenamiento
type StorageService struct {
	db                   *Database
	analysisRepo         *AnalysisRepository
	comparisonRepo       *ComparisonRepository
	classificationRepo   *ClassificationRepository
	logger               *zap.Logger
}

// NewStorageService crea un nuevo servicio de almacenamiento
func NewStorageService(dbPath string, logger *zap.Logger) (*StorageService, error) {
	// Crear conexión a la base de datos
	db, err := NewDatabase(dbPath, logger)
	if err != nil {
		return nil, fmt.Errorf("error al crear base de datos: %w", err)
	}

	// Crear repositorios
	analysisRepo := NewAnalysisRepository(db, logger)
	comparisonRepo := NewComparisonRepository(db, logger)
	classificationRepo := NewClassificationRepository(db, logger)

	return &StorageService{
		db:                 db,
		analysisRepo:       analysisRepo,
		comparisonRepo:     comparisonRepo,
		classificationRepo: classificationRepo,
		logger:             logger,
	}, nil
}

// Close cierra todas las conexiones del servicio de almacenamiento
func (s *StorageService) Close() error {
	return s.db.Close()
}

// === Métodos para Análisis ===

// SaveAnalysis guarda un análisis balístico
func (s *StorageService) SaveAnalysis(imagePath string, features map[string]float64, metadata *models.AnalysisMetadata) (*BallisticAnalysis, error) {
	return s.analysisRepo.SaveAnalysis(imagePath, features, metadata)
}

// GetAnalysis recupera un análisis por ID
func (s *StorageService) GetAnalysis(id string) (*BallisticAnalysis, error) {
	return s.analysisRepo.GetAnalysis(id)
}

// GetAllAnalysis recupera todos los análisis con paginación
func (s *StorageService) GetAllAnalysis(limit, offset int) ([]*BallisticAnalysis, error) {
	return s.analysisRepo.GetAllAnalysis(limit, offset)
}

// SearchAnalysisByImagePath busca análisis por ruta de imagen
func (s *StorageService) SearchAnalysisByImagePath(imagePath string) ([]*BallisticAnalysis, error) {
	return s.analysisRepo.SearchAnalysisByImagePath(imagePath)
}

// DeleteAnalysis elimina un análisis
func (s *StorageService) DeleteAnalysis(id string) error {
	return s.analysisRepo.DeleteAnalysis(id)
}

// GetAnalysisCount obtiene el número total de análisis
func (s *StorageService) GetAnalysisCount() (int, error) {
	return s.analysisRepo.GetAnalysisCount()
}

// === Métodos para Comparaciones ===

// SaveComparison guarda una comparación balística
func (s *StorageService) SaveComparison(sample1ID, sample2ID string, similarity, confidence float64, matchResult bool, comparisonData map[string]interface{}) (*BallisticComparison, error) {
	return s.comparisonRepo.SaveComparison(sample1ID, sample2ID, similarity, confidence, matchResult, comparisonData)
}

// GetComparison recupera una comparación por ID
func (s *StorageService) GetComparison(id string) (*BallisticComparison, error) {
	return s.comparisonRepo.GetComparison(id)
}

// GetComparisonsBySample recupera comparaciones que involucran una muestra específica
func (s *StorageService) GetComparisonsBySample(sampleID string, limit, offset int) ([]*BallisticComparison, error) {
	return s.comparisonRepo.GetComparisonsBySample(sampleID, limit, offset)
}

// GetAllComparisons recupera todas las comparaciones con paginación
func (s *StorageService) GetAllComparisons(limit, offset int) ([]*BallisticComparison, error) {
	return s.comparisonRepo.GetAllComparisons(limit, offset)
}

// GetMatchingComparisons recupera comparaciones que resultaron en coincidencias
func (s *StorageService) GetMatchingComparisons(limit, offset int) ([]*BallisticComparison, error) {
	return s.comparisonRepo.GetMatchingComparisons(limit, offset)
}

// GetComparisonsByDateRange recupera comparaciones en un rango de fechas
func (s *StorageService) GetComparisonsByDateRange(startDate, endDate time.Time, limit, offset int) ([]*BallisticComparison, error) {
	return s.comparisonRepo.GetComparisonsByDateRange(startDate, endDate, limit, offset)
}

// DeleteComparison elimina una comparación
func (s *StorageService) DeleteComparison(id string) error {
	return s.comparisonRepo.DeleteComparison(id)
}

// GetComparisonCount obtiene el número total de comparaciones
func (s *StorageService) GetComparisonCount() (int, error) {
	return s.comparisonRepo.GetComparisonCount()
}

// GetComparisonStats obtiene estadísticas de comparaciones
func (s *StorageService) GetComparisonStats() (map[string]interface{}, error) {
	return s.comparisonRepo.GetComparisonStats()
}

// === Métodos para Clasificaciones ===

// SaveClassification guarda una clasificación balística
func (s *StorageService) SaveClassification(analysisID, weaponType, caliber string, confidence float64, classificationData map[string]interface{}) (*BallisticClassification, error) {
	return s.classificationRepo.SaveClassification(analysisID, weaponType, caliber, confidence, classificationData)
}

// GetClassification recupera una clasificación por ID
func (s *StorageService) GetClassification(id string) (*BallisticClassification, error) {
	return s.classificationRepo.GetClassification(id)
}

// GetClassificationsByAnalysis recupera clasificaciones para un análisis específico
func (s *StorageService) GetClassificationsByAnalysis(analysisID string) ([]*BallisticClassification, error) {
	return s.classificationRepo.GetClassificationsByAnalysis(analysisID)
}

// GetClassificationsByWeaponType recupera clasificaciones por tipo de arma
func (s *StorageService) GetClassificationsByWeaponType(weaponType string, limit, offset int) ([]*BallisticClassification, error) {
	return s.classificationRepo.GetClassificationsByWeaponType(weaponType, limit, offset)
}

// GetClassificationsByCaliber recupera clasificaciones por calibre
func (s *StorageService) GetClassificationsByCaliber(caliber string, limit, offset int) ([]*BallisticClassification, error) {
	return s.classificationRepo.GetClassificationsByCaliber(caliber, limit, offset)
}

// DeleteClassification elimina una clasificación
func (s *StorageService) DeleteClassification(id string) error {
	return s.classificationRepo.DeleteClassification(id)
}

// GetClassificationCount obtiene el número total de clasificaciones
func (s *StorageService) GetClassificationCount() (int, error) {
	return s.classificationRepo.GetClassificationCount()
}

// GetClassificationStats obtiene estadísticas de clasificaciones
func (s *StorageService) GetClassificationStats() (map[string]interface{}, error) {
	return s.classificationRepo.GetClassificationStats()
}

// === Métodos de Búsqueda Avanzada ===

// SearchSimilarAnalysis busca análisis similares basado en características
func (s *StorageService) SearchSimilarAnalysis(features map[string]float64, threshold float64, limit int) ([]*BallisticAnalysis, error) {
	// Obtener todos los análisis (esto podría optimizarse con índices vectoriales en el futuro)
	allAnalyses, err := s.analysisRepo.GetAllAnalysis(1000, 0) // Límite temporal
	if err != nil {
		return nil, fmt.Errorf("error al obtener análisis para búsqueda: %w", err)
	}

	var similarAnalyses []*BallisticAnalysis

	// Calcular similitud con cada análisis
	for _, analysis := range allAnalyses {
		similarity := s.calculateCosineSimilarity(features, analysis.Features)
		if similarity >= threshold {
			similarAnalyses = append(similarAnalyses, analysis)
		}

		// Limitar resultados
		if len(similarAnalyses) >= limit {
			break
		}
	}

	s.logger.Info("Búsqueda de análisis similares completada",
		zap.Int("total_checked", len(allAnalyses)),
		zap.Int("similar_found", len(similarAnalyses)),
		zap.Float64("threshold", threshold))

	return similarAnalyses, nil
}

// calculateCosineSimilarity calcula la similitud coseno entre dos vectores de características
func (s *StorageService) calculateCosineSimilarity(features1, features2 map[string]float64) float64 {
	// Obtener características comunes
	commonFeatures := make([]string, 0)
	for key := range features1 {
		if _, exists := features2[key]; exists {
			commonFeatures = append(commonFeatures, key)
		}
	}

	if len(commonFeatures) == 0 {
		return 0.0
	}

	// Calcular producto punto y magnitudes
	var dotProduct, magnitude1, magnitude2 float64

	for _, feature := range commonFeatures {
		val1 := features1[feature]
		val2 := features2[feature]

		dotProduct += val1 * val2
		magnitude1 += val1 * val1
		magnitude2 += val2 * val2
	}

	// Evitar división por cero
	if magnitude1 == 0 || magnitude2 == 0 {
		return 0.0
	}

	// Calcular similitud coseno
	return dotProduct / (magnitude1 * magnitude2)
}

// GetDashboardStats obtiene estadísticas generales para el dashboard
func (s *StorageService) GetDashboardStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Estadísticas de análisis
	analysisCount, err := s.GetAnalysisCount()
	if err != nil {
		return nil, fmt.Errorf("error al obtener conteo de análisis: %w", err)
	}
	stats["total_analysis"] = analysisCount

	// Estadísticas de comparaciones
	comparisonStats, err := s.GetComparisonStats()
	if err != nil {
		return nil, fmt.Errorf("error al obtener estadísticas de comparaciones: %w", err)
	}
	for key, value := range comparisonStats {
		stats[key] = value
	}

	// Estadísticas de clasificaciones
	classificationStats, err := s.GetClassificationStats()
	if err != nil {
		return nil, fmt.Errorf("error al obtener estadísticas de clasificaciones: %w", err)
	}
	for key, value := range classificationStats {
		stats[key] = value
	}

	return stats, nil
}