package models

type BallisticImage struct {
	ID          string      `json:"id"`
	FileName    string      `json:"file_name"`
	Descriptors [][]float32 `json:"-"`
	Keypoints   []Keypoint  `json:"keypoints"`
	Timestamp   int64       `json:"timestamp"`
}

type Keypoint struct {
	X        float32 `json:"x"`
	Y        float32 `json:"y"`
	Size     float32 `json:"size"`
	Angle    float32 `json:"angle"`
	Response float32 `json:"response"`
}

type MatchResult struct {
	ImageID    string  `json:"image_id"`
	Score      float64 `json:"score"`
	Distance   float64 `json:"distance"`
	Matches    int     `json:"matches"`
	SourceFile string  `json:"source_file"`
}
