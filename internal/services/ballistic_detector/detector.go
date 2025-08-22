package ballistic_detector

import (
	"fmt"
	"math"

	"go.uber.org/zap"
)

// BallisticDetector servicio para detectar características balísticas específicas
type BallisticDetector struct {
	logger *zap.Logger
}

// NewBallisticDetector crea una nueva instancia del detector balístico
func NewBallisticDetector(logger *zap.Logger) *BallisticDetector {
	return &BallisticDetector{
		logger: logger,
	}
}

// WeaponClassification resultado de clasificación de arma
type WeaponClassification struct {
	WeaponType  string             `json:"weapon_type"`
	Caliber     string             `json:"caliber"`
	Confidence  float64            `json:"confidence"`
	Indicators  map[string]float64 `json:"indicators"`
	Evidence    []string           `json:"evidence"`
}

// BallisticCharacteristics características balísticas específicas detectadas
type BallisticCharacteristics struct {
	FiringPinType      string  `json:"firing_pin_type"`
	BreechFacePattern  string  `json:"breech_face_pattern"`
	StriationPattern   string  `json:"striation_pattern"`
	RiflingType        string  `json:"rifling_type"`
	TwistDirection     string  `json:"twist_direction"`
	LandGrooveRatio    float64 `json:"land_groove_ratio"`
	StriationDensity   float64 `json:"striation_density"`
	FiringPinDepth     float64 `json:"firing_pin_depth"`
}

// DetectWeaponType clasifica el tipo de arma basándose en las características
func (bd *BallisticDetector) DetectWeaponType(features map[string]float64) *WeaponClassification {
	indicators := make(map[string]float64)
	evidence := []string{}

	// Analizar características de marcas de percutor
	firingPinScore := bd.analyzeFiringPinCharacteristics(features, &evidence)
	indicators["firing_pin_score"] = firingPinScore

	// Analizar patrones de estriado
	striationScore := bd.analyzeStriationCharacteristics(features, &evidence)
	indicators["striation_score"] = striationScore

	// Analizar características de cara de cierre
	breechFaceScore := bd.analyzeBreechFaceCharacteristics(features, &evidence)
	indicators["breech_face_score"] = breechFaceScore

	// Determinar tipo de arma basándose en los indicadores
	weaponType, confidence := bd.classifyWeaponType(indicators, features)

	bd.logger.Info("Clasificación de arma completada",
		zap.String("weapon_type", weaponType),
		zap.Float64("confidence", confidence),
		zap.Strings("evidence", evidence))

	return &WeaponClassification{
		WeaponType: weaponType,
		Confidence: confidence,
		Indicators: indicators,
		Evidence:   evidence,
	}
}

// DetectCaliber estima el calibre basándose en las características
func (bd *BallisticDetector) DetectCaliber(features map[string]float64) *WeaponClassification {
	indicators := make(map[string]float64)
	evidence := []string{}

	// Analizar dimensiones del proyectil
	sizeScore := bd.analyzeProjectileSize(features, &evidence)
	indicators["size_score"] = sizeScore

	// Analizar patrones de rifling
	riflingScore := bd.analyzeRiflingPattern(features, &evidence)
	indicators["rifling_score"] = riflingScore

	// Analizar densidad de estriado
	striationDensity := bd.analyzeStriationDensity(features, &evidence)
	indicators["striation_density"] = striationDensity

	// Determinar calibre basándose en los indicadores
	caliber, confidence := bd.classifyCaliber(indicators, features)

	bd.logger.Info("Clasificación de calibre completada",
		zap.String("caliber", caliber),
		zap.Float64("confidence", confidence),
		zap.Strings("evidence", evidence))

	return &WeaponClassification{
		Caliber:    caliber,
		Confidence: confidence,
		Indicators: indicators,
		Evidence:   evidence,
	}
}

// DetectBallisticCharacteristics detecta características balísticas específicas
func (bd *BallisticDetector) DetectBallisticCharacteristics(features map[string]float64) *BallisticCharacteristics {
	characteristics := &BallisticCharacteristics{}

	// Detectar tipo de percutor
	characteristics.FiringPinType = bd.detectFiringPinType(features)

	// Detectar patrón de cara de cierre
	characteristics.BreechFacePattern = bd.detectBreechFacePattern(features)

	// Detectar patrón de estriado
	characteristics.StriationPattern = bd.detectStriationPattern(features)

	// Detectar tipo de rifling
	characteristics.RiflingType = bd.detectRiflingType(features)

	// Detectar dirección de twist
	characteristics.TwistDirection = bd.detectTwistDirection(features)

	// Calcular métricas cuantitativas
	characteristics.LandGrooveRatio = bd.calculateLandGrooveRatio(features)
	characteristics.StriationDensity = bd.calculateStriationDensity(features)
	characteristics.FiringPinDepth = bd.calculateFiringPinDepth(features)

	bd.logger.Info("Características balísticas detectadas",
		zap.String("firing_pin_type", characteristics.FiringPinType),
		zap.String("rifling_type", characteristics.RiflingType),
		zap.Float64("striation_density", characteristics.StriationDensity))

	return characteristics
}

// analyzeFiringPinCharacteristics analiza características del percutor
func (bd *BallisticDetector) analyzeFiringPinCharacteristics(features map[string]float64, evidence *[]string) float64 {
	score := 0.0

	// Analizar número de marcas de percutor
	if firingPinCount, exists := features["firing_pin_count"]; exists {
		if firingPinCount > 0 {
			score += 0.3
			*evidence = append(*evidence, fmt.Sprintf("Marcas de percutor detectadas: %.0f", firingPinCount))
		}
	}

	// Analizar radio promedio de marcas de percutor
	if avgRadius, exists := features["firing_pin_avg_radius"]; exists {
		if avgRadius > 0 {
			score += 0.4
			*evidence = append(*evidence, fmt.Sprintf("Radio promedio de percutor: %.2f", avgRadius))
			
			// Clasificar por tamaño de percutor
			if avgRadius < 2.0 {
				*evidence = append(*evidence, "Percutor pequeño (posible pistola)")
				score += 0.2
			} else if avgRadius > 4.0 {
				*evidence = append(*evidence, "Percutor grande (posible rifle)")
				score += 0.3
			}
		}
	}

	return math.Min(score, 1.0)
}

// analyzeStriationCharacteristics analiza características de estriado
func (bd *BallisticDetector) analyzeStriationCharacteristics(features map[string]float64, evidence *[]string) float64 {
	score := 0.0

	// Analizar número de patrones de estriado
	if striationCount, exists := features["striation_count"]; exists {
		if striationCount > 0 {
			score += 0.3
			*evidence = append(*evidence, fmt.Sprintf("Patrones de estriado detectados: %.0f", striationCount))
		}
	}

	// Analizar ángulo promedio de estriado
	if avgAngle, exists := features["striation_avg_angle"]; exists {
		if avgAngle > 0 {
			score += 0.4
			*evidence = append(*evidence, fmt.Sprintf("Ángulo promedio de estriado: %.2f°", avgAngle))
		}
	}

	// Analizar fuerza promedio de estriado
	if avgStrength, exists := features["striation_avg_strength"]; exists {
		if avgStrength > 0.5 {
			score += 0.3
			*evidence = append(*evidence, "Estriado fuerte detectado")
		}
	}

	return math.Min(score, 1.0)
}

// analyzeBreechFaceCharacteristics analiza características de cara de cierre
func (bd *BallisticDetector) analyzeBreechFaceCharacteristics(features map[string]float64, evidence *[]string) float64 {
	score := 0.0

	// Analizar uniformidad LBP (textura de superficie)
	if lbpUniformity, exists := features["lbp_uniformity"]; exists {
		if lbpUniformity > 0.05 {
			score += 0.4
			*evidence = append(*evidence, "Textura de cara de cierre detectada")
		}
	}

	// Analizar características de contorno
	if contourArea, exists := features["contour_area"]; exists {
		if contourArea > 1000 {
			score += 0.3
			*evidence = append(*evidence, "Área de contacto significativa")
		}
	}

	return math.Min(score, 1.0)
}

// classifyWeaponType clasifica el tipo de arma basándose en los indicadores
func (bd *BallisticDetector) classifyWeaponType(indicators map[string]float64, features map[string]float64) (string, float64) {
	// Calcular puntuación total
	totalScore := 0.0
	for _, score := range indicators {
		totalScore += score
	}
	avgScore := totalScore / float64(len(indicators))

	// Extraer características clave
	firingPinCount := features["firing_pin_count"]
	firingPinRadius := features["firing_pin_avg_radius"]
	striationCount := features["striation_count"]
	striationAngle := features["striation_avg_angle"]
	contourArea := features["contour_area"]
	lbpUniformity := features["lbp_uniformity"]

	// Sistema de puntuación por características
	scores := make(map[string]float64)

	// Análisis para Pistola
	pistolaScore := 0.0
	if firingPinRadius >= 1.0 && firingPinRadius <= 2.5 {
		pistolaScore += 0.3
	}
	if striationCount >= 5 && striationCount <= 15 {
		pistolaScore += 0.25
	}
	if contourArea >= 200 && contourArea <= 800 {
		pistolaScore += 0.2
	}
	if striationAngle >= 10 && striationAngle <= 30 {
		pistolaScore += 0.15
	}
	if lbpUniformity >= 0.03 && lbpUniformity <= 0.08 {
		pistolaScore += 0.1
	}
	scores["Pistola"] = pistolaScore

	// Análisis para Rifle
	rifleScore := 0.0
	if firingPinRadius >= 1.5 && firingPinRadius <= 3.5 {
		rifleScore += 0.25
	}
	if striationCount >= 15 && striationCount <= 35 {
		rifleScore += 0.3
	}
	if contourArea >= 400 && contourArea <= 1500 {
		rifleScore += 0.2
	}
	if striationAngle >= 15 && striationAngle <= 45 {
		rifleScore += 0.15
	}
	if lbpUniformity >= 0.05 && lbpUniformity <= 0.12 {
		rifleScore += 0.1
	}
	scores["Rifle"] = rifleScore

	// Análisis para Revólver
	revolverScore := 0.0
	if firingPinCount > 1 {
		revolverScore += 0.4 // Característica distintiva
	}
	if firingPinRadius >= 1.2 && firingPinRadius <= 2.8 {
		revolverScore += 0.2
	}
	if striationCount >= 8 && striationCount <= 20 {
		revolverScore += 0.2
	}
	if contourArea >= 300 && contourArea <= 1000 {
		revolverScore += 0.15
	}
	if lbpUniformity >= 0.04 && lbpUniformity <= 0.09 {
		revolverScore += 0.05
	}
	scores["Revólver"] = revolverScore

	// Análisis para Escopeta
	escopetaScore := 0.0
	if firingPinRadius >= 2.0 && firingPinRadius <= 4.0 {
		escopetaScore += 0.25
	}
	if striationCount >= 20 || striationCount <= 3 { // Muchas o muy pocas estrías
		escopetaScore += 0.3
	}
	if contourArea >= 800 {
		escopetaScore += 0.2
	}
	if striationAngle <= 15 || striationAngle >= 60 {
		escopetaScore += 0.15
	}
	if lbpUniformity >= 0.08 {
		escopetaScore += 0.1
	}
	scores["Escopeta"] = escopetaScore

	// Análisis para Subfusil
	subfusilScore := 0.0
	if firingPinRadius >= 1.8 && firingPinRadius <= 3.0 {
		subfusilScore += 0.25
	}
	if striationCount >= 12 && striationCount <= 25 {
		subfusilScore += 0.3
	}
	if contourArea >= 250 && contourArea <= 600 {
		subfusilScore += 0.2
	}
	if striationAngle >= 20 && striationAngle <= 40 {
		subfusilScore += 0.15
	}
	if lbpUniformity >= 0.06 && lbpUniformity <= 0.11 {
		subfusilScore += 0.1
	}
	scores["Subfusil"] = subfusilScore

	// Encontrar la clasificación con mayor puntuación
	maxScore := 0.0
	bestType := "Indeterminado"
	for weaponType, score := range scores {
		if score > maxScore {
			maxScore = score
			bestType = weaponType
		}
	}

	// Calcular confianza final
	confidence := maxScore * avgScore

	// Aplicar umbral mínimo de confianza
	if confidence < 0.4 {
		bestType = "Indeterminado"
		confidence *= 0.7
	}

	return bestType, math.Min(confidence, 1.0)
}

// analyzeProjectileSize analiza el tamaño del proyectil
func (bd *BallisticDetector) analyzeProjectileSize(features map[string]float64, evidence *[]string) float64 {
	score := 0.0

	if contourArea, exists := features["contour_area"]; exists {
		*evidence = append(*evidence, fmt.Sprintf("Área del proyectil: %.0f", contourArea))
		
		// Clasificar por tamaño
		if contourArea < 500 {
			*evidence = append(*evidence, "Proyectil pequeño")
			score = 0.8
		} else if contourArea > 2000 {
			*evidence = append(*evidence, "Proyectil grande")
			score = 0.9
		} else {
			*evidence = append(*evidence, "Proyectil mediano")
			score = 0.7
		}
	}

	return score
}

// analyzeRiflingPattern analiza el patrón de rifling
func (bd *BallisticDetector) analyzeRiflingPattern(features map[string]float64, evidence *[]string) float64 {
	score := 0.0

	if striationAngle, exists := features["striation_avg_angle"]; exists {
		*evidence = append(*evidence, fmt.Sprintf("Ángulo de rifling: %.2f°", striationAngle))
		
		// Analizar ángulo de rifling
		if striationAngle > 10 && striationAngle < 30 {
			*evidence = append(*evidence, "Rifling estándar")
			score = 0.8
		} else if striationAngle > 30 {
			*evidence = append(*evidence, "Rifling agresivo")
			score = 0.9
		}
	}

	return score
}

// analyzeStriationDensity analiza la densidad de estriado
func (bd *BallisticDetector) analyzeStriationDensity(features map[string]float64, evidence *[]string) float64 {
	if striationCount, exists := features["striation_count"]; exists {
		*evidence = append(*evidence, fmt.Sprintf("Densidad de estriado: %.0f", striationCount))
		return math.Min(striationCount/50.0, 1.0) // Normalizar a 0-1
	}
	return 0.0
}

// classifyCaliber clasifica el calibre basándose en los indicadores
func (bd *BallisticDetector) classifyCaliber(indicators map[string]float64, features map[string]float64) (string, float64) {
	// Extraer características clave
	contourArea := features["contour_area"]
	striationCount := features["striation_count"]
	firingPinRadius := features["firing_pin_avg_radius"]
	striationAngle := features["striation_avg_angle"]
	lbpUniformity := features["lbp_uniformity"]

	// Calcular confianza promedio
	totalScore := 0.0
	for _, score := range indicators {
		totalScore += score
	}
	avgScore := totalScore / float64(len(indicators))

	// Sistema de puntuación por calibre
	scores := make(map[string]float64)

	// Análisis para .22 LR
	twentyTwoScore := 0.0
	if contourArea >= 150 && contourArea <= 350 {
		twentyTwoScore += 0.35
	}
	if firingPinRadius >= 0.8 && firingPinRadius <= 1.6 {
		twentyTwoScore += 0.3
	}
	if striationCount >= 3 && striationCount <= 8 {
		twentyTwoScore += 0.2
	}
	if striationAngle >= 5 && striationAngle <= 20 {
		twentyTwoScore += 0.1
	}
	if lbpUniformity <= 0.05 {
		twentyTwoScore += 0.05
	}
	scores[".22 LR"] = twentyTwoScore

	// Análisis para 9mm
	nineMMScore := 0.0
	if contourArea >= 400 && contourArea <= 900 {
		nineMMScore += 0.35
	}
	if firingPinRadius >= 1.2 && firingPinRadius <= 2.2 {
		nineMMScore += 0.3
	}
	if striationCount >= 8 && striationCount <= 18 {
		nineMMScore += 0.2
	}
	if striationAngle >= 12 && striationAngle <= 28 {
		nineMMScore += 0.1
	}
	if lbpUniformity >= 0.03 && lbpUniformity <= 0.08 {
		nineMMScore += 0.05
	}
	scores["9mm"] = nineMMScore

	// Análisis para .40 S&W
	fortyScore := 0.0
	if contourArea >= 600 && contourArea <= 1200 {
		fortyScore += 0.35
	}
	if firingPinRadius >= 1.4 && firingPinRadius <= 2.4 {
		fortyScore += 0.3
	}
	if striationCount >= 12 && striationCount <= 22 {
		fortyScore += 0.2
	}
	if striationAngle >= 15 && striationAngle <= 32 {
		fortyScore += 0.1
	}
	if lbpUniformity >= 0.04 && lbpUniformity <= 0.09 {
		fortyScore += 0.05
	}
	scores[".40 S&W"] = fortyScore

	// Análisis para .45 ACP
	fortyFiveScore := 0.0
	if contourArea >= 800 && contourArea <= 1500 {
		fortyFiveScore += 0.35
	}
	if firingPinRadius >= 1.6 && firingPinRadius <= 2.8 {
		fortyFiveScore += 0.3
	}
	if striationCount >= 10 && striationCount <= 20 {
		fortyFiveScore += 0.2
	}
	if striationAngle >= 10 && striationAngle <= 25 {
		fortyFiveScore += 0.1
	}
	if lbpUniformity >= 0.05 && lbpUniformity <= 0.1 {
		fortyFiveScore += 0.05
	}
	scores[".45 ACP"] = fortyFiveScore

	// Análisis para .308 Winchester
	threeOhEightScore := 0.0
	if contourArea >= 1000 && contourArea <= 2000 {
		threeOhEightScore += 0.35
	}
	if firingPinRadius >= 1.8 && firingPinRadius <= 3.2 {
		threeOhEightScore += 0.3
	}
	if striationCount >= 20 && striationCount <= 35 {
		threeOhEightScore += 0.2
	}
	if striationAngle >= 20 && striationAngle <= 45 {
		threeOhEightScore += 0.1
	}
	if lbpUniformity >= 0.06 && lbpUniformity <= 0.12 {
		threeOhEightScore += 0.05
	}
	scores[".308 Winchester"] = threeOhEightScore

	// Análisis para .30-06
	thirtyOhSixScore := 0.0
	if contourArea >= 1500 && contourArea <= 2500 {
		thirtyOhSixScore += 0.35
	}
	if firingPinRadius >= 2.0 && firingPinRadius <= 3.5 {
		thirtyOhSixScore += 0.3
	}
	if striationCount >= 25 && striationCount <= 40 {
		thirtyOhSixScore += 0.2
	}
	if striationAngle >= 25 && striationAngle <= 50 {
		thirtyOhSixScore += 0.1
	}
	if lbpUniformity >= 0.07 && lbpUniformity <= 0.13 {
		thirtyOhSixScore += 0.05
	}
	scores[".30-06"] = thirtyOhSixScore

	// Análisis para .38 Special
	thirtyEightScore := 0.0
	if contourArea >= 500 && contourArea <= 1000 {
		thirtyEightScore += 0.35
	}
	if firingPinRadius >= 1.3 && firingPinRadius <= 2.3 {
		thirtyEightScore += 0.3
	}
	if striationCount >= 10 && striationCount <= 20 {
		thirtyEightScore += 0.2
	}
	if striationAngle >= 12 && striationAngle <= 30 {
		thirtyEightScore += 0.1
	}
	if lbpUniformity >= 0.04 && lbpUniformity <= 0.09 {
		thirtyEightScore += 0.05
	}
	scores[".38 Special"] = thirtyEightScore

	// Análisis para .357 Magnum
	threeFiftySeven := 0.0
	if contourArea >= 600 && contourArea <= 1200 {
		threeFiftySeven += 0.35
	}
	if firingPinRadius >= 1.4 && firingPinRadius <= 2.5 {
		threeFiftySeven += 0.3
	}
	if striationCount >= 12 && striationCount <= 25 {
		threeFiftySeven += 0.2
	}
	if striationAngle >= 15 && striationAngle <= 35 {
		threeFiftySeven += 0.1
	}
	if lbpUniformity >= 0.05 && lbpUniformity <= 0.1 {
		threeFiftySeven += 0.05
	}
	scores[".357 Magnum"] = threeFiftySeven

	// Encontrar la clasificación con mayor puntuación
	maxScore := 0.0
	bestCaliber := "Indeterminado"
	for caliber, score := range scores {
		if score > maxScore {
			maxScore = score
			bestCaliber = caliber
		}
	}

	// Calcular confianza final
	confidence := maxScore * avgScore

	// Aplicar umbral mínimo de confianza
	if confidence < 0.35 {
		bestCaliber = "Indeterminado"
		confidence *= 0.6
	}

	return bestCaliber, math.Min(confidence, 1.0)
}

// Métodos auxiliares para detectar características específicas

func (bd *BallisticDetector) detectFiringPinType(features map[string]float64) string {
	if radius, exists := features["firing_pin_avg_radius"]; exists {
		if radius < 1.5 {
			return "Circular pequeño"
		} else if radius > 3.0 {
			return "Circular grande"
		}
		return "Circular mediano"
	}
	return "Indeterminado"
}

func (bd *BallisticDetector) detectBreechFacePattern(features map[string]float64) string {
	if lbp, exists := features["lbp_uniformity"]; exists {
		if lbp > 0.1 {
			return "Textura rugosa"
		} else if lbp > 0.05 {
			return "Textura media"
		}
		return "Textura lisa"
	}
	return "Indeterminado"
}

func (bd *BallisticDetector) detectStriationPattern(features map[string]float64) string {
	if count, exists := features["striation_count"]; exists {
		if count > 30 {
			return "Estriado denso"
		} else if count > 15 {
			return "Estriado moderado"
		} else if count > 5 {
			return "Estriado ligero"
		}
		return "Estriado mínimo"
	}
	return "Indeterminado"
}

func (bd *BallisticDetector) detectRiflingType(features map[string]float64) string {
	if angle, exists := features["striation_avg_angle"]; exists {
		if angle > 45 {
			return "Rifling pronunciado"
		} else if angle > 20 {
			return "Rifling estándar"
		} else if angle > 5 {
			return "Rifling suave"
		}
		return "Rifling mínimo"
	}
	return "Indeterminado"
}

func (bd *BallisticDetector) detectTwistDirection(features map[string]float64) string {
	// Simplificado: basado en ángulo promedio
	if angle, exists := features["striation_avg_angle"]; exists {
		if angle > 90 {
			return "Izquierda"
		}
		return "Derecha"
	}
	return "Indeterminado"
}

func (bd *BallisticDetector) calculateLandGrooveRatio(features map[string]float64) float64 {
	// Estimación basada en características de estriado
	if striationCount, exists := features["striation_count"]; exists {
		if striationLength, exists := features["striation_avg_length"]; exists {
			return striationLength / (striationCount + 1)
		}
	}
	return 0.0
}

func (bd *BallisticDetector) calculateStriationDensity(features map[string]float64) float64 {
	if count, exists := features["striation_count"]; exists {
		if area, exists := features["contour_area"]; exists && area > 0 {
			return count / area * 1000 // Densidad por unidad de área
		}
	}
	return 0.0
}

func (bd *BallisticDetector) calculateFiringPinDepth(features map[string]float64) float64 {
	// Estimación basada en el radio del percutor
	if radius, exists := features["firing_pin_avg_radius"]; exists {
		return radius * 0.3 // Estimación simplificada
	}
	return 0.0
}