package models

import (
	"image"
)

type RGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}

type ColorData struct {
	Color     RGB     `json:"color"`
	Frequency float64 `json:"frequency"`
}

type ChromaAnalysis struct {
	DominantColors []ColorData        `json:"dominant_colors"`
	ColorVariance  map[string]float64 `json:"color_variance"`
}

type BallisticAnalysis struct {
	Features       []float64         `json:"features"`
	ChromaData     ChromaAnalysis    `json:"chroma_data"`
	ProcessedImage image.Image       `json:"-"` // Will be handled separately
	Metadata       BallisticMetadata `json:"metadata"`
}

type BallisticMetadata struct {
	Caliber       string  `json:"caliber"`
	Manufacturer  string  `json:"manufacturer"`
	FirearmType   string  `json:"firearm_type"`
	DateCollected string  `json:"date_collected"`
	Confidence    float64 `json:"confidence"`
}

type CartridgeCase struct {
	ID              string    `json:"id"`
	BreechFaceMarks []float64 `json:"breech_face_marks"`
	FiringPinMarks  []float64 `json:"firing_pin_marks"`
	ChamberMarks    []float64 `json:"chamber_marks"`
	ExtractorMarks  []float64 `json:"extractor_marks"`
	EjectorMarks    []float64 `json:"ejector_marks"`
}

type Bullet struct {
	ID             string    `json:"id"`
	LandMarks      []float64 `json:"land_marks"`
	GrooveMarks    []float64 `json:"groove_marks"`
	StriaePatterns []float64 `json:"striae_patterns"`
	BaseFeatures   []float64 `json:"base_features"`
}

type ComparisonResult struct {
	Sample1ID       string   `json:"sample1_id"`
	Sample2ID       string   `json:"sample2_id"`
	Similarity      float64  `json:"similarity"`
	Match           bool     `json:"match"`
	Confidence      float64  `json:"confidence"`
	AreasOfInterest []string `json:"areas_of_interest"`
}

type DatabaseRecord struct {
	ID          string        `json:"id"`
	CaseData    CartridgeCase `json:"case_data"`
	BulletData  Bullet        `json:"bullet_data"`
	Images      []string      `json:"images"`
	CreatedAt   string        `json:"created_at"`
	LastUpdated string        `json:"last_updated"`
}
