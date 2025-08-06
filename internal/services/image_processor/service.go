package chroma

import (
	"image"
	"image/color"
)

type Service struct {
	config *Config
}

type Config struct {
	ColorThreshold float64
	SampleSize     int
}

type ChromaAnalysis struct {
	DominantColors []ColorData
	ColorVariance  map[string]float64
}

type ColorData struct {
	Color     RGB
	Frequency float64
}

type RGB struct {
	R uint8
	G uint8
	B uint8
}

func NewService(cfg *Config) *Service {
	return &Service{
		config: cfg,
	}
}

func (s *Service) Analyze(img image.Image) (*ChromaAnalysis, error) {
	bounds := img.Bounds()
	samplePoints := s.generateSamplePoints(bounds)
	colorValues := make([]color.Color, 0, len(samplePoints))

	for _, point := range samplePoints {
		c := img.At(point.X, point.Y)
		colorValues = append(colorValues, c)
	}

	dominantColors := s.calculateDominantColors(colorValues)
	colorVariance := s.calculateColorVariance(colorValues)

	return &ChromaAnalysis{
		DominantColors: dominantColors,
		ColorVariance:  colorVariance,
	}, nil
}

func (s *Service) generateSamplePoints(bounds image.Rectangle) []image.Point {
	points := make([]image.Point, 0)
	stepX := bounds.Dx() / s.config.SampleSize
	stepY := bounds.Dy() / s.config.SampleSize

	if stepX < 1 {
		stepX = 1
	}
	if stepY < 1 {
		stepY = 1
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y += stepY {
		for x := bounds.Min.X; x < bounds.Max.X; x += stepX {
			points = append(points, image.Point{X: x, Y: y})
		}
	}

	return points
}

func (s *Service) calculateDominantColors(colors []color.Color) []ColorData {
	colorCount := make(map[RGB]int)

	for _, c := range colors {
		r, g, b, _ := c.RGBA()
		rgb := RGB{
			R: uint8(r >> 8),
			G: uint8(g >> 8),
			B: uint8(b >> 8),
		}
		colorCount[rgb]++
	}

	var dominant []ColorData
	for clr, count := range colorCount {
		freq := float64(count) / float64(len(colors))
		if freq > s.config.ColorThreshold {
			dominant = append(dominant, ColorData{
				Color:     clr,
				Frequency: freq,
			})
		}
	}

	return dominant
}

func (s *Service) calculateColorVariance(colors []color.Color) map[string]float64 {
	var sumR, sumG, sumB float64
	var sumRSq, sumGSq, sumBSq float64

	for _, c := range colors {
		r, g, b, _ := c.RGBA()
		rf := float64(r >> 8)
		gf := float64(g >> 8)
		bf := float64(b >> 8)

		sumR += rf
		sumG += gf
		sumB += bf

		sumRSq += rf * rf
		sumGSq += gf * gf
		sumBSq += bf * bf
	}

	n := float64(len(colors))
	variance := make(map[string]float64)

	variance["red"] = (sumRSq - (sumR*sumR)/n) / n
	variance["green"] = (sumGSq - (sumG*sumG)/n) / n
	variance["blue"] = (sumBSq - (sumB*sumB)/n) / n

	return variance
}
