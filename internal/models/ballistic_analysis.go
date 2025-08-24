package models

import (
	"time"
)

type RGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

func (c RGB) RGBA() (r, g, b, a uint32) {
	return uint32(c.R)<<8 | uint32(c.R)<<0,
		uint32(c.G)<<8 | uint32(c.G)<<0,
		uint32(c.B)<<8 | uint32(c.B)<<0,
		0xFFFF
}

type ColorData struct {
	Color     RGB     `json:"color"`
	Frequency float64 `json:"frequency"`
}

type ChromaAnalysis struct {
	DominantColors []ColorData `json:"dominant_colors"`
	ColorVariance  float64     `json:"color_variance"` // Cambiado de map[string]float64 a float64
}

type BallisticAnalysis struct {
	Features       map[string]float64      `json:"features"` // Cambiado de []float64 a map
	ChromaData     ChromaAnalysis          `json:"chroma_data"`
	Classification *ClassificationResult   `json:"classification,omitempty"` // Nuevo: clasificación automática
	Metadata       AnalysisMetadata        `json:"metadata"` // Estructura mejorada
}

// ClassificationResult resultado de clasificación automática
type ClassificationResult struct {
	WeaponType   string             `json:"weapon_type"`
	Caliber      string             `json:"caliber"`
	Confidence   float64            `json:"confidence"`
	Indicators   map[string]float64 `json:"indicators"`
	Evidence     []string           `json:"evidence"`
	OverallScore float64            `json:"overall_score"`
}

// AnalysisMetadata reemplaza BallisticMetadata con campos más relevantes
type AnalysisMetadata struct {
	Timestamp          string  `json:"timestamp"`              // ISO 8601 format
	ImageHash          string  `json:"image_hash"`             // SHA-256 del archivo
	ProcessorVersion   string  `json:"processor_version"`      // Versión del algoritmo
	PythonFeaturesUsed bool    `json:"python_features"`        // Indica si se usó Python
	Confidence         float64 `json:"confidence"`             // Confianza del análisis (0-1)
	Caliber            string  `json:"caliber,omitempty"`      // Opcional
	Manufacturer       string  `json:"manufacturer,omitempty"` // Opcional
	Filename           string  `json:"filename,omitempty"`     // Nombre del archivo original
	ContentType        string  `json:"content_type,omitempty"` // Tipo MIME del archivo
	FileSize           int64   `json:"file_size,omitempty"`    // Tamaño del archivo en bytes
}

type CartridgeCase struct {
	ID              string             `json:"id"`
	BreechFaceMarks map[string]float64 `json:"breech_face_marks"` // Cambiado a map
	FiringPinMarks  map[string]float64 `json:"firing_pin_marks"`  // Cambiado a map
	ChamberMarks    map[string]float64 `json:"chamber_marks"`     // Cambiado a map
	ExtractorMarks  map[string]float64 `json:"extractor_marks"`   // Cambiado a map
	EjectorMarks    map[string]float64 `json:"ejector_marks"`     // Cambiado a map
	Features        map[string]float64 `json:"features"`          // Nuevo: características consolidadas
}

type Bullet struct {
	ID             string             `json:"id"`
	LandMarks      map[string]float64 `json:"land_marks"`      // Cambiado a map
	GrooveMarks    map[string]float64 `json:"groove_marks"`    // Cambiado a map
	StriaePatterns map[string]float64 `json:"striae_patterns"` // Cambiado a map
	BaseFeatures   map[string]float64 `json:"base_features"`   // Cambiado a map
}

type ComparisonResult struct {
	Sample1ID       string             `json:"sample1_id"`
	Sample2ID       string             `json:"sample2_id"`
	Similarity      float64            `json:"similarity"`
	Match           bool               `json:"match"`
	Confidence      float64            `json:"confidence"`
	FeatureWeights  map[string]float64 `json:"feature_weights"` // Nuevo: pesos por característica
	AreasOfInterest []string           `json:"areas_of_interest"`
	DiffPerFeature  map[string]float64 `json:"diff_per_feature"` // Nuevo: diferencias por feature
}

type DatabaseRecord struct {
	ID          string            `json:"id"`
	CaseData    CartridgeCase     `json:"case_data"`
	BulletData  Bullet            `json:"bullet_data"`
	Images      []string          `json:"images"`
	CreatedAt   time.Time         `json:"created_at"`   // Cambiado a time.Time
	LastUpdated time.Time         `json:"last_updated"` // Cambiado a time.Time
	Analysis    BallisticAnalysis `json:"analysis"`     // Nuevo: resultados completos
}

// Helper functions
func NewBallisticAnalysis() BallisticAnalysis {
	return BallisticAnalysis{
		Features: make(map[string]float64),
		Metadata: AnalysisMetadata{
			Timestamp:        time.Now().UTC().Format(time.RFC3339),
			ProcessorVersion: "1.3.0",
		},
	}
}

func NewCartridgeCase() CartridgeCase {
	return CartridgeCase{
		BreechFaceMarks: make(map[string]float64),
		FiringPinMarks:  make(map[string]float64),
		ChamberMarks:    make(map[string]float64),
		ExtractorMarks:  make(map[string]float64),
		EjectorMarks:    make(map[string]float64),
		Features:        make(map[string]float64),
	}
}
