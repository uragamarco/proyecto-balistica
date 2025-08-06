package image_processor

import (
	"image"
	
	"gocv.io/x/gocv"
)

type Processor struct {
	orb   *gocv.ORB
	clahe *gocv.CLAHE
}

func New() *Processor {
	return &Processor{
		orb:   gocv.NewORBWithParams(500, 1.2, 8, 31, 0, 2, gocv.ORBScoreHarris, 31),
		clahe: gocv.NewCLAHEWithParams(2.0, image.Pt(8, 8)),
	}
}

func (p *Processor) Process(img gocv.Mat) ([]gocv.KeyPoint, gocv.Mat, error) {
	// 1. Mejorar contraste
	enhanced := p.enhanceContrast(img)
	defer enhanced.Close()

	// 2. Extraer características
	kp, desc := p.orb.DetectAndCompute(enhanced, gocv.NewMat())
	if len(kp) == 0 {
		return nil, gocv.Mat{}, fmt.Errorf("no se detectaron características")
	}

	// 3. Normalizar descriptores
	return kp, p.normalize(desc), nil
}

func (p *Processor) enhanceContrast(img gocv.Mat) gocv.Mat {
	lab := gocv.NewMat()
	defer lab.Close()
	
	gocv.CvtColor(img, &lab, gocv.ColorBGRToLab)
	channels := gocv.Split(lab)
	p.clahe.Apply(channels[0], &channels[0])
	gocv.Merge(channels, &lab)
	gocv.CvtColor(lab, &img, gocv.ColorLabToBGR)
	
	return img
}

func (p *Processor) normalize(desc gocv.Mat) gocv.Mat {
	normalized := gocv.NewMat()
	gocv.Normalize(desc, &normalized, 1.0, 0.0, gocv.NormL2)
	return normalized
}
