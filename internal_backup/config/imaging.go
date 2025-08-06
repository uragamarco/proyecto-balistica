package config

type ImagingConfig struct {
	CLAHEClipLimit   float64 `yaml:"clahe_clip_limit"`
	CLAHETileSize    int     `yaml:"clahe_tile_size"`
	ORBFeatures      int     `yaml:"orb_features"`
	ORBScaleFactor   float64 `yaml:"orb_scale_factor"`
	ORBEdgeThreshold int     `yaml:"orb_edge_threshold"`
	MinKeypointScore float64 `yaml:"min_keypoint_score"`
}

func DefaultImagingConfig() ImagingConfig {
	return ImagingConfig{
		CLAHEClipLimit:   2.0,
		CLAHETileSize:    8,
		ORBFeatures:      500,
		ORBScaleFactor:   1.2,
		ORBEdgeThreshold: 31,
		MinKeypointScore: 0.01,
	}
}
