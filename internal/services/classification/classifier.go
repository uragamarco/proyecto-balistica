package classification

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/uragamarco/proyecto-balistica/internal/services/ballistic_detector"
	"github.com/uragamarco/proyecto-balistica/internal/storage"
)

// ClassificationService servicio para clasificación de armas y calibres
type ClassificationService struct {
	detector *ballistic_detector.BallisticDetector
	storage  *storage.StorageService
	logger   *zap.Logger
}

// NewClassificationService crea una nueva instancia del servicio de clasificación
func NewClassificationService(
	detector *ballistic_detector.BallisticDetector,
	storageService *storage.StorageService,
	logger *zap.Logger,
) *ClassificationService {
	return &ClassificationService{
		detector: detector,
		storage:  storageService,
		logger:   logger,
	}
}

// ClassificationResult resultado completo de clasificación
type ClassificationResult struct {
	AnalysisID      string                                           `json:"analysis_id"`
	WeaponType      *ballistic_detector.WeaponClassification        `json:"weapon_type"`
	Caliber         *ballistic_detector.WeaponClassification        `json:"caliber"`
	Characteristics *ballistic_detector.BallisticCharacteristics   `json:"characteristics"`
	OverallScore    float64                                          `json:"overall_score"`
	Recommendations []string                                         `json:"recommendations"`
}

// ClassifyBallistic realiza clasificación completa de una muestra balística
func (cs *ClassificationService) ClassifyBallistic(ctx context.Context, analysisID string, features map[string]float64) (*ClassificationResult, error) {
	cs.logger.Info("Iniciando clasificación balística",
		zap.String("analysis_id", analysisID),
		zap.Int("features_count", len(features)))

	// Detectar tipo de arma
	weaponType := cs.detector.DetectWeaponType(features)

	// Detectar calibre
	caliber := cs.detector.DetectCaliber(features)

	// Detectar características específicas
	characteristics := cs.detector.DetectBallisticCharacteristics(features)

	// Calcular puntuación general
	overallScore := cs.calculateOverallScore(weaponType, caliber)

	// Generar recomendaciones
	recommendations := cs.generateRecommendations(weaponType, caliber, characteristics)

	// Crear resultado
	result := &ClassificationResult{
		AnalysisID:      analysisID,
		WeaponType:      weaponType,
		Caliber:         caliber,
		Characteristics: characteristics,
		OverallScore:    overallScore,
		Recommendations: recommendations,
	}

	// Guardar clasificación en la base de datos
	err := cs.saveClassification(ctx, result)
	if err != nil {
		cs.logger.Error("Error guardando clasificación",
			zap.String("analysis_id", analysisID),
			zap.Error(err))
		return nil, fmt.Errorf("error guardando clasificación: %w", err)
	}

	cs.logger.Info("Clasificación balística completada",
		zap.String("analysis_id", analysisID),
		zap.String("weapon_type", weaponType.WeaponType),
		zap.String("caliber", caliber.Caliber),
		zap.Float64("overall_score", overallScore))

	return result, nil
}

// GetClassificationHistory obtiene el historial de clasificaciones
func (cs *ClassificationService) GetClassificationHistory(ctx context.Context, limit int) ([]*storage.BallisticClassification, error) {
	// Por ahora retornamos una lista vacía hasta implementar el método en StorageService
	cs.logger.Info("Obteniendo historial de clasificaciones")
	return []*storage.BallisticClassification{}, nil
}

// GetClassificationByAnalysisID obtiene clasificación por ID de análisis
func (cs *ClassificationService) GetClassificationByAnalysisID(ctx context.Context, analysisID string) ([]*storage.BallisticClassification, error) {
	classifications, err := cs.storage.GetClassificationsByAnalysis(analysisID)
	if err != nil {
		cs.logger.Error("Error obteniendo clasificación",
			zap.String("analysis_id", analysisID),
			zap.Error(err))
		return nil, fmt.Errorf("error obteniendo clasificación: %w", err)
	}

	return classifications, nil
}

// SearchByWeaponType busca clasificaciones por tipo de arma
func (cs *ClassificationService) SearchByWeaponType(ctx context.Context, weaponType string) ([]*storage.BallisticClassification, error) {
	classifications, err := cs.storage.GetClassificationsByWeaponType(weaponType, 100, 0)
	if err != nil {
		cs.logger.Error("Error buscando por tipo de arma",
			zap.String("weapon_type", weaponType),
			zap.Error(err))
		return nil, fmt.Errorf("error buscando por tipo de arma: %w", err)
	}

	cs.logger.Info("Búsqueda por tipo de arma completada",
		zap.String("weapon_type", weaponType),
		zap.Int("results", len(classifications)))

	return classifications, nil
}

// SearchByCaliber busca clasificaciones por calibre
func (cs *ClassificationService) SearchByCaliber(ctx context.Context, caliber string) ([]*storage.BallisticClassification, error) {
	classifications, err := cs.storage.GetClassificationsByCaliber(caliber, 100, 0)
	if err != nil {
		cs.logger.Error("Error buscando por calibre",
			zap.String("caliber", caliber),
			zap.Error(err))
		return nil, fmt.Errorf("error buscando por calibre: %w", err)
	}

	cs.logger.Info("Búsqueda por calibre completada",
		zap.String("caliber", caliber),
		zap.Int("results", len(classifications)))

	return classifications, nil
}

// calculateOverallScore calcula la puntuación general de la clasificación
func (cs *ClassificationService) calculateOverallScore(weaponType, caliber *ballistic_detector.WeaponClassification) float64 {
	// Ponderar las confianzas de tipo de arma y calibre
	weaponWeight := 0.6
	caliberWeight := 0.4

	overallScore := (weaponType.Confidence * weaponWeight) + (caliber.Confidence * caliberWeight)

	// Aplicar penalización si alguna clasificación es "Indeterminado"
	if weaponType.WeaponType == "Indeterminado" {
		overallScore *= 0.7
	}
	if caliber.Caliber == "Indeterminado" {
		overallScore *= 0.8
	}

	return overallScore
}

// generateRecommendations genera recomendaciones basadas en la clasificación
func (cs *ClassificationService) generateRecommendations(weaponType, caliber *ballistic_detector.WeaponClassification, characteristics *ballistic_detector.BallisticCharacteristics) []string {
	recommendations := []string{}

	// Recomendaciones basadas en confianza
	if weaponType.Confidence < 0.7 {
		recommendations = append(recommendations, "Se recomienda análisis adicional para confirmar el tipo de arma")
	}
	if caliber.Confidence < 0.7 {
		recommendations = append(recommendations, "Se recomienda análisis adicional para confirmar el calibre")
	}

	// Recomendaciones basadas en características
	if characteristics.StriationDensity < 0.1 {
		recommendations = append(recommendations, "Baja densidad de estriado detectada - verificar calidad de la muestra")
	}
	if characteristics.FiringPinDepth < 0.5 {
		recommendations = append(recommendations, "Marca de percutor poco profunda - posible desgaste del arma")
	}

	// Recomendaciones para clasificaciones indeterminadas
	if weaponType.WeaponType == "Indeterminado" {
		recommendations = append(recommendations, "Considerar análisis comparativo con base de datos de referencia")
	}
	if caliber.Caliber == "Indeterminado" {
		recommendations = append(recommendations, "Realizar mediciones físicas adicionales para determinar calibre")
	}

	// Recomendaciones específicas por tipo de arma
	switch weaponType.WeaponType {
	case "Pistola":
		recommendations = append(recommendations, "Verificar características típicas de pistola semiautomática")
	case "Rifle":
		recommendations = append(recommendations, "Analizar patrones de rifling característicos de rifle")
	case "Revólver":
		recommendations = append(recommendations, "Examinar múltiples marcas de percutor típicas de revólver")
	case "Escopeta":
		recommendations = append(recommendations, "Considerar análisis de patrones de perdigones si aplica")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Clasificación completada con alta confianza")
	}

	return recommendations
}

// saveClassification guarda la clasificación en la base de datos
func (cs *ClassificationService) saveClassification(_ context.Context, result *ClassificationResult) error {
	// Crear datos de clasificación simplificados
	classificationData := map[string]interface{}{
		"weapon_indicators":   result.WeaponType.Indicators,
		"caliber_indicators":  result.Caliber.Indicators,
		"weapon_evidence":     result.WeaponType.Evidence,
		"caliber_evidence":    result.Caliber.Evidence,
		"characteristics": map[string]interface{}{
			"firing_pin_type":     result.Characteristics.FiringPinType,
			"breech_face_pattern": result.Characteristics.BreechFacePattern,
			"striation_pattern":   result.Characteristics.StriationPattern,
			"rifling_type":        result.Characteristics.RiflingType,
			"twist_direction":     result.Characteristics.TwistDirection,
			"land_groove_ratio":   result.Characteristics.LandGrooveRatio,
			"striation_density":   result.Characteristics.StriationDensity,
			"firing_pin_depth":    result.Characteristics.FiringPinDepth,
		},
		"recommendations":     result.Recommendations,
	}

	_, err := cs.storage.SaveClassification(result.AnalysisID, result.WeaponType.WeaponType, result.Caliber.Caliber, result.OverallScore, classificationData)
	return err
}

// combineIndicators combina indicadores de tipo de arma y calibre
func (cs *ClassificationService) combineIndicators(weaponIndicators, caliberIndicators map[string]float64) map[string]float64 {
	combined := make(map[string]float64)

	// Copiar indicadores de arma con prefijo
	for key, value := range weaponIndicators {
		combined["weapon_"+key] = value
	}

	// Copiar indicadores de calibre con prefijo
	for key, value := range caliberIndicators {
		combined["caliber_"+key] = value
	}

	return combined
}

// combineEvidence combina evidencia de tipo de arma y calibre
func (cs *ClassificationService) combineEvidence(weaponEvidence, caliberEvidence []string) []string {
	combined := []string{}
	combined = append(combined, weaponEvidence...)
	combined = append(combined, caliberEvidence...)
	return combined
}