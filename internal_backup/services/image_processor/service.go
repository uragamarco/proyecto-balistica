package image_processor

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

type BallisticImageProcessor struct {
	orb   *gocv.ORB
	clahe *gocv.CLAHE
}

func New() *BallisticImageProcessor {
	return &BallisticImageProcessor{
		orb:   gocv.NewORBWithParams(500, 1.2, 8, 31, 0, 2, gocv.ORBScoreHarris, 31),
		clahe: gocv.NewCLAHEWithParams(2.0, image.Pt(8, 8)),
	}
}

func (p *BallisticImageProcessor) Process(img gocv.Mat) ([]gocv.KeyPoint, gocv.Mat, error) {
	// 1. Preprocesamiento
	enhanced := p.enhanceContrast(img)
	defer enhanced.Close()

	// 2. Extracción de características
	kp, desc := p.orb.DetectAndCompute(enhanced, gocv.NewMat())
	if len(kp) == 0 {
		return nil, gocv.Mat{}, fmt.Errorf("no features detected")
	}

	// 3. Filtrado
	filteredKP, filteredDesc := p.filterFeatures(kp, desc)

	return filteredKP, p.normalize(filteredDesc), nil
}

func (p *BallisticImageProcessor) enhanceContrast(img gocv.Mat) gocv.Mat {
	lab := gocv.NewMat()
	gocv.CvtColor(img, &lab, gocv.ColorBGRToLab)

	channels := gocv.Split(lab)
	p.clahe.Apply(channels[0], &channels[0])

	gocv.Merge(channels, &lab)
	gocv.CvtColor(lab, &img, gocv.ColorLabToBGR)
	return img
}

func (p *BallisticImageProcessor) filterFeatures(kp []gocv.KeyPoint, desc gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	var filteredKP []gocv.KeyPoint
	filteredDesc := gocv.NewMat()

	for i, point := range kp {
		if point.Response > 0.01 {
			filteredKP = append(filteredKP, point)
			filteredDesc.PushBack(desc.Row(i))
		}
	}

	return filteredKP, filteredDesc
}

func (p *BallisticImageProcessor) normalize(desc gocv.Mat) gocv.Mat {
	normalized := gocv.NewMat()
	gocv.Normalize(desc, &normalized, 1.0, 0.0, gocv.NormL2)
	return normalized
}
