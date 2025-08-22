package comparison

import (
	"math"
	"sort"

	"go.uber.org/zap"
)

// AdvancedComparison implementa algoritmos avanzados de comparación balística
type AdvancedComparison struct {
	logger *zap.Logger
}

// NewAdvancedComparison crea una nueva instancia del servicio de comparación avanzado
func NewAdvancedComparison(logger *zap.Logger) *AdvancedComparison {
	return &AdvancedComparison{
		logger: logger,
	}
}

// ComparisonResult contiene el resultado detallado de una comparación
type ComparisonResult struct {
	Similarity          float64                `json:"similarity"`
	Confidence          float64                `json:"confidence"`
	Match               bool                   `json:"match"`
	CriticalFeatures    []string               `json:"critical_features"`
	FeatureScores       map[string]float64     `json:"feature_scores"`
	StatisticalMetrics  StatisticalMetrics     `json:"statistical_metrics"`
	BallisticIndicators BallisticIndicators    `json:"ballistic_indicators"`
}

// StatisticalMetrics contiene métricas estadísticas de la comparación
type StatisticalMetrics struct {
	CorrelationCoeff    float64 `json:"correlation_coefficient"`
	EuclideanDistance   float64 `json:"euclidean_distance"`
	ManhattanDistance   float64 `json:"manhattan_distance"`
	CosineSimilarity    float64 `json:"cosine_similarity"`
	JaccardIndex        float64 `json:"jaccard_index"`
}

// BallisticIndicators contiene indicadores específicos de análisis balístico
type BallisticIndicators struct {
	StriationMatch      float64 `json:"striation_match"`
	FiringPinMatch      float64 `json:"firing_pin_match"`
	BreechFaceMatch     float64 `json:"breech_face_match"`
	EjectorMatch        float64 `json:"ejector_match"`
	ExtractorMatch      float64 `json:"extractor_match"`
	OverallBallisticScore float64 `json:"overall_ballistic_score"`
}

// FeatureWeights define los pesos para diferentes tipos de características
type FeatureWeights struct {
	// Características balísticas críticas
	StriationFeatures   float64 `json:"striation_features"`
	FiringPinFeatures   float64 `json:"firing_pin_features"`
	BreechFaceFeatures  float64 `json:"breech_face_features"`
	EjectorFeatures     float64 `json:"ejector_features"`
	ExtractorFeatures   float64 `json:"extractor_features"`
	
	// Características geométricas
	GeometricFeatures   float64 `json:"geometric_features"`
	TextureFeatures     float64 `json:"texture_features"`
	ColorFeatures       float64 `json:"color_features"`
	
	// Características de forma
	ShapeFeatures       float64 `json:"shape_features"`
	ContourFeatures     float64 `json:"contour_features"`
}

// GetDefaultWeights retorna los pesos por defecto para características balísticas
func GetDefaultWeights() FeatureWeights {
	return FeatureWeights{
		// Características balísticas tienen mayor peso
		StriationFeatures:  3.0,
		FiringPinFeatures:  2.8,
		BreechFaceFeatures: 2.5,
		EjectorFeatures:    2.2,
		ExtractorFeatures:  2.0,
		
		// Características geométricas
		GeometricFeatures:  1.5,
		TextureFeatures:    1.3,
		ColorFeatures:      1.0,
		
		// Características de forma
		ShapeFeatures:      1.8,
		ContourFeatures:    1.6,
	}
}

// CompareAdvanced realiza una comparación avanzada entre dos conjuntos de características
func (ac *AdvancedComparison) CompareAdvanced(features1, features2 map[string]float64, weights FeatureWeights) *ComparisonResult {
	ac.logger.Info("Iniciando comparación avanzada", 
		zap.Int("features1_count", len(features1)),
		zap.Int("features2_count", len(features2)))

	result := &ComparisonResult{
		FeatureScores: make(map[string]float64),
	}

	// Calcular métricas estadísticas
	result.StatisticalMetrics = ac.calculateStatisticalMetrics(features1, features2)

	// Calcular indicadores balísticos específicos
	result.BallisticIndicators = ac.calculateBallisticIndicators(features1, features2, weights)

	// Calcular similitud ponderada avanzada
	result.Similarity = ac.calculateWeightedSimilarity(features1, features2, weights)

	// Calcular puntuaciones por característica
	result.FeatureScores = ac.calculateFeatureScores(features1, features2)

	// Identificar características críticas
	result.CriticalFeatures = ac.identifyCriticalFeatures(features1, features2)

	// Calcular confianza basada en múltiples factores
	result.Confidence = ac.calculateAdvancedConfidence(result)

	// Determinar si es una coincidencia
	result.Match = ac.determineMatch(result)

	ac.logger.Info("Comparación avanzada completada", 
		zap.Float64("similarity", result.Similarity),
		zap.Float64("confidence", result.Confidence),
		zap.Bool("match", result.Match))

	return result
}

// calculateStatisticalMetrics calcula métricas estadísticas entre dos conjuntos de características
func (ac *AdvancedComparison) calculateStatisticalMetrics(f1, f2 map[string]float64) StatisticalMetrics {
	var metrics StatisticalMetrics

	// Obtener características comunes
	commonFeatures := ac.getCommonFeatures(f1, f2)
	if len(commonFeatures) == 0 {
		return metrics
	}

	// Preparar vectores para cálculos
	vec1, vec2 := ac.prepareVectors(f1, f2, commonFeatures)

	// Coeficiente de correlación de Pearson
	metrics.CorrelationCoeff = ac.calculateCorrelation(vec1, vec2)

	// Distancia euclidiana
	metrics.EuclideanDistance = ac.calculateEuclideanDistance(vec1, vec2)

	// Distancia de Manhattan
	metrics.ManhattanDistance = ac.calculateManhattanDistance(vec1, vec2)

	// Similitud del coseno
	metrics.CosineSimilarity = ac.calculateCosineSimilarity(vec1, vec2)

	// Índice de Jaccard (adaptado para características continuas)
	metrics.JaccardIndex = ac.calculateJaccardIndex(f1, f2)

	return metrics
}

// calculateBallisticIndicators calcula indicadores específicos de análisis balístico
func (ac *AdvancedComparison) calculateBallisticIndicators(f1, f2 map[string]float64, weights FeatureWeights) BallisticIndicators {
	var indicators BallisticIndicators

	// Analizar características de estriado
	indicators.StriationMatch = ac.analyzeStriationFeatures(f1, f2)

	// Analizar características del percutor
	indicators.FiringPinMatch = ac.analyzeFiringPinFeatures(f1, f2)

	// Analizar características de la cara de cierre
	indicators.BreechFaceMatch = ac.analyzeBreechFaceFeatures(f1, f2)

	// Analizar características del eyector
	indicators.EjectorMatch = ac.analyzeEjectorFeatures(f1, f2)

	// Analizar características del extractor
	indicators.ExtractorMatch = ac.analyzeExtractorFeatures(f1, f2)

	// Calcular puntuación balística general
	indicators.OverallBallisticScore = ac.calculateOverallBallisticScore(indicators, weights)

	return indicators
}

// getCommonFeatures obtiene las características comunes entre dos mapas
func (ac *AdvancedComparison) getCommonFeatures(f1, f2 map[string]float64) []string {
	var common []string
	for key := range f1 {
		if _, exists := f2[key]; exists {
			common = append(common, key)
		}
	}
	sort.Strings(common) // Para consistencia
	return common
}

// prepareVectors prepara vectores ordenados para cálculos estadísticos
func (ac *AdvancedComparison) prepareVectors(f1, f2 map[string]float64, features []string) ([]float64, []float64) {
	vec1 := make([]float64, len(features))
	vec2 := make([]float64, len(features))

	for i, feature := range features {
		vec1[i] = f1[feature]
		vec2[i] = f2[feature]
	}

	return vec1, vec2
}

// calculateCorrelation calcula el coeficiente de correlación de Pearson
func (ac *AdvancedComparison) calculateCorrelation(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) || len(vec1) == 0 {
		return 0.0
	}

	n := float64(len(vec1))
	sum1, sum2, sum1Sq, sum2Sq, sumProduct := 0.0, 0.0, 0.0, 0.0, 0.0

	for i := 0; i < len(vec1); i++ {
		sum1 += vec1[i]
		sum2 += vec2[i]
		sum1Sq += vec1[i] * vec1[i]
		sum2Sq += vec2[i] * vec2[i]
		sumProduct += vec1[i] * vec2[i]
	}

	numerator := sumProduct - (sum1*sum2)/n
	denominator := math.Sqrt((sum1Sq - sum1*sum1/n) * (sum2Sq - sum2*sum2/n))

	if denominator == 0 {
		return 0.0
	}

	return numerator / denominator
}

// calculateEuclideanDistance calcula la distancia euclidiana
func (ac *AdvancedComparison) calculateEuclideanDistance(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return math.Inf(1)
	}

	sum := 0.0
	for i := 0; i < len(vec1); i++ {
		diff := vec1[i] - vec2[i]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// calculateManhattanDistance calcula la distancia de Manhattan
func (ac *AdvancedComparison) calculateManhattanDistance(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return math.Inf(1)
	}

	sum := 0.0
	for i := 0; i < len(vec1); i++ {
		sum += math.Abs(vec1[i] - vec2[i])
	}

	return sum
}

// calculateCosineSimilarity calcula la similitud del coseno
func (ac *AdvancedComparison) calculateCosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) || len(vec1) == 0 {
		return 0.0
	}

	dotProduct, norm1, norm2 := 0.0, 0.0, 0.0

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// calculateJaccardIndex calcula el índice de Jaccard adaptado para características continuas
func (ac *AdvancedComparison) calculateJaccardIndex(f1, f2 map[string]float64) float64 {
	allFeatures := make(map[string]bool)
	for k := range f1 {
		allFeatures[k] = true
	}
	for k := range f2 {
		allFeatures[k] = true
	}

	intersection := 0.0
	union := float64(len(allFeatures))

	for feature := range allFeatures {
		v1, exists1 := f1[feature]
		v2, exists2 := f2[feature]

		if exists1 && exists2 {
			// Similitud basada en la diferencia normalizada
			maxVal := math.Max(math.Abs(v1), math.Abs(v2))
			if maxVal < 1e-9 {
				maxVal = 1
			}
			similarity := 1 - math.Abs(v1-v2)/maxVal
			intersection += similarity
		}
	}

	if union == 0 {
		return 0.0
	}

	return intersection / union
}

// calculateWeightedSimilarity calcula similitud ponderada avanzada
func (ac *AdvancedComparison) calculateWeightedSimilarity(f1, f2 map[string]float64, weights FeatureWeights) float64 {
	commonFeatures := ac.getCommonFeatures(f1, f2)
	if len(commonFeatures) == 0 {
		return 0.0
	}

	var totalWeight, weightedSum float64

	for _, feature := range commonFeatures {
		v1, v2 := f1[feature], f2[feature]
		weight := ac.getFeatureWeight(feature, weights)
		
		// Calcular similitud normalizada
		maxVal := math.Max(math.Abs(v1), math.Abs(v2))
		if maxVal < 1e-9 {
			maxVal = 1
		}
		
		diff := math.Abs(v1 - v2)
		normalizedSimilarity := 1 - (diff / maxVal)
		
		weightedSum += weight * normalizedSimilarity
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return weightedSum / totalWeight
}

// calculateFeatureScores calcula puntuaciones individuales por característica
func (ac *AdvancedComparison) calculateFeatureScores(f1, f2 map[string]float64) map[string]float64 {
	scores := make(map[string]float64)
	
	for feature, v1 := range f1 {
		if v2, exists := f2[feature]; exists {
			maxVal := math.Max(math.Abs(v1), math.Abs(v2))
			if maxVal < 1e-9 {
				maxVal = 1
			}
			
			diff := math.Abs(v1 - v2)
			scores[feature] = 1 - (diff / maxVal)
		}
	}
	
	return scores
}

// identifyCriticalFeatures identifica características con diferencias significativas
func (ac *AdvancedComparison) identifyCriticalFeatures(f1, f2 map[string]float64) []string {
	var critical []string
	const threshold = 0.3 // Umbral para diferencias críticas
	
	for feature, v1 := range f1 {
		if v2, exists := f2[feature]; exists {
			maxVal := math.Max(math.Abs(v1), math.Abs(v2))
			if maxVal < 1e-9 {
				maxVal = 1
			}
			
			diff := math.Abs(v1 - v2) / maxVal
			if diff > threshold {
				critical = append(critical, feature)
			}
		}
	}
	
	sort.Strings(critical)
	return critical
}

// calculateAdvancedConfidence calcula confianza basada en múltiples factores
func (ac *AdvancedComparison) calculateAdvancedConfidence(result *ComparisonResult) float64 {
	// Factores que contribuyen a la confianza
	correlationFactor := math.Abs(result.StatisticalMetrics.CorrelationCoeff) * 0.3
	cosineFactor := result.StatisticalMetrics.CosineSimilarity * 0.2
	jaccardFactor := result.StatisticalMetrics.JaccardIndex * 0.2
	ballisticFactor := result.BallisticIndicators.OverallBallisticScore * 0.3
	
	// Penalizar por características críticas
	criticalPenalty := float64(len(result.CriticalFeatures)) * 0.05
	
	confidence := correlationFactor + cosineFactor + jaccardFactor + ballisticFactor - criticalPenalty
	
	// Normalizar entre 0 y 1
	if confidence < 0 {
		confidence = 0
	} else if confidence > 1 {
		confidence = 1
	}
	
	return confidence
}

// determineMatch determina si hay coincidencia basada en múltiples criterios
func (ac *AdvancedComparison) determineMatch(result *ComparisonResult) bool {
	// Umbrales para diferentes métricas
	const (
		similarityThreshold = 0.85
		confidenceThreshold = 0.75
		ballisticThreshold = 0.80
		maxCriticalFeatures = 3
	)
	
	// Criterios para determinar coincidencia
	highSimilarity := result.Similarity >= similarityThreshold
	highConfidence := result.Confidence >= confidenceThreshold
	goodBallisticScore := result.BallisticIndicators.OverallBallisticScore >= ballisticThreshold
	fewCriticalFeatures := len(result.CriticalFeatures) <= maxCriticalFeatures
	
	// Requiere al menos 3 de 4 criterios
	criteriaMet := 0
	if highSimilarity { criteriaMet++ }
	if highConfidence { criteriaMet++ }
	if goodBallisticScore { criteriaMet++ }
	if fewCriticalFeatures { criteriaMet++ }
	
	return criteriaMet >= 3
}

// getFeatureWeight obtiene el peso de una característica específica
func (ac *AdvancedComparison) getFeatureWeight(feature string, weights FeatureWeights) float64 {
	// Mapear características a sus pesos correspondientes
	switch {
	case ac.isStriationFeature(feature):
		return weights.StriationFeatures
	case ac.isFiringPinFeature(feature):
		return weights.FiringPinFeatures
	case ac.isBreechFaceFeature(feature):
		return weights.BreechFaceFeatures
	case ac.isEjectorFeature(feature):
		return weights.EjectorFeatures
	case ac.isExtractorFeature(feature):
		return weights.ExtractorFeatures
	case ac.isGeometricFeature(feature):
		return weights.GeometricFeatures
	case ac.isTextureFeature(feature):
		return weights.TextureFeatures
	case ac.isColorFeature(feature):
		return weights.ColorFeatures
	case ac.isShapeFeature(feature):
		return weights.ShapeFeatures
	case ac.isContourFeature(feature):
		return weights.ContourFeatures
	default:
		return 1.0 // Peso por defecto
	}
}

// Métodos para identificar tipos de características
func (ac *AdvancedComparison) isStriationFeature(feature string) bool {
	return feature == "striation_density" || feature == "striation_angle" || feature == "striation_depth"
}

func (ac *AdvancedComparison) isFiringPinFeature(feature string) bool {
	return feature == "firing_pin_impression" || feature == "firing_pin_shape" || feature == "firing_pin_depth"
}

func (ac *AdvancedComparison) isBreechFaceFeature(feature string) bool {
	return feature == "breech_face_marks" || feature == "breech_face_texture" || feature == "breech_face_pattern"
}

func (ac *AdvancedComparison) isEjectorFeature(feature string) bool {
	return feature == "ejector_marks" || feature == "ejector_position" || feature == "ejector_shape"
}

func (ac *AdvancedComparison) isExtractorFeature(feature string) bool {
	return feature == "extractor_marks" || feature == "extractor_groove" || feature == "extractor_depth"
}

func (ac *AdvancedComparison) isGeometricFeature(feature string) bool {
	return feature == "area" || feature == "perimeter" || feature == "aspect_ratio" || feature == "extent"
}

func (ac *AdvancedComparison) isTextureFeature(feature string) bool {
	return feature == "contrast" || feature == "dissimilarity" || feature == "homogeneity" || feature == "energy"
}

func (ac *AdvancedComparison) isColorFeature(feature string) bool {
	return feature == "mean_hue" || feature == "mean_saturation" || feature == "mean_value" || feature == "color_variance"
}

func (ac *AdvancedComparison) isShapeFeature(feature string) bool {
	return feature == "circularity" || feature == "solidity" || feature == "convexity" || feature == "eccentricity"
}

func (ac *AdvancedComparison) isContourFeature(feature string) bool {
	return feature == "contour_length" || feature == "contour_smoothness" || feature == "contour_complexity"
}

// Métodos para análisis de características balísticas específicas
func (ac *AdvancedComparison) analyzeStriationFeatures(f1, f2 map[string]float64) float64 {
	striationFeatures := []string{"striation_density", "striation_angle", "striation_depth"}
	return ac.analyzeFeatureGroup(f1, f2, striationFeatures)
}

func (ac *AdvancedComparison) analyzeFiringPinFeatures(f1, f2 map[string]float64) float64 {
	firingPinFeatures := []string{"firing_pin_impression", "firing_pin_shape", "firing_pin_depth"}
	return ac.analyzeFeatureGroup(f1, f2, firingPinFeatures)
}

func (ac *AdvancedComparison) analyzeBreechFaceFeatures(f1, f2 map[string]float64) float64 {
	breechFaceFeatures := []string{"breech_face_marks", "breech_face_texture", "breech_face_pattern"}
	return ac.analyzeFeatureGroup(f1, f2, breechFaceFeatures)
}

func (ac *AdvancedComparison) analyzeEjectorFeatures(f1, f2 map[string]float64) float64 {
	ejectorFeatures := []string{"ejector_marks", "ejector_position", "ejector_shape"}
	return ac.analyzeFeatureGroup(f1, f2, ejectorFeatures)
}

func (ac *AdvancedComparison) analyzeExtractorFeatures(f1, f2 map[string]float64) float64 {
	extractorFeatures := []string{"extractor_marks", "extractor_groove", "extractor_depth"}
	return ac.analyzeFeatureGroup(f1, f2, extractorFeatures)
}

// analyzeFeatureGroup analiza un grupo específico de características
func (ac *AdvancedComparison) analyzeFeatureGroup(f1, f2 map[string]float64, features []string) float64 {
	var totalSimilarity float64
	validFeatures := 0
	
	for _, feature := range features {
		v1, exists1 := f1[feature]
		v2, exists2 := f2[feature]
		
		if exists1 && exists2 {
			maxVal := math.Max(math.Abs(v1), math.Abs(v2))
			if maxVal < 1e-9 {
				maxVal = 1
			}
			
			diff := math.Abs(v1 - v2)
			similarity := 1 - (diff / maxVal)
			totalSimilarity += similarity
			validFeatures++
		}
	}
	
	if validFeatures == 0 {
		return 0.0
	}
	
	return totalSimilarity / float64(validFeatures)
}

// calculateOverallBallisticScore calcula la puntuación balística general
func (ac *AdvancedComparison) calculateOverallBallisticScore(indicators BallisticIndicators, weights FeatureWeights) float64 {
	// Pesos relativos para cada indicador balístico
	totalWeight := weights.StriationFeatures + weights.FiringPinFeatures + weights.BreechFaceFeatures + 
				  weights.EjectorFeatures + weights.ExtractorFeatures
	
	if totalWeight == 0 {
		return 0.0
	}
	
	weightedScore := (indicators.StriationMatch * weights.StriationFeatures +
					 indicators.FiringPinMatch * weights.FiringPinFeatures +
					 indicators.BreechFaceMatch * weights.BreechFaceFeatures +
					 indicators.EjectorMatch * weights.EjectorFeatures +
					 indicators.ExtractorMatch * weights.ExtractorFeatures) / totalWeight
	
	return weightedScore
}