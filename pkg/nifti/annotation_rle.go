package nifti

type SegmentRLE struct {
	encodedSeg []float64
	decodedSeg []float64
	zIndex     float64
	tIndex     float64
	pixVal     float64
}

type SegmentationRLEOption func(s *SegmentRLE)

// WithEncodedSegmentation allows user to specify the RLE-encoded segmentation
func WithEncodedSegmentation(encodedSeg []float64) SegmentationRLEOption {
	return func(s *SegmentRLE) {
		s.encodedSeg = encodedSeg
	}
}

// WithDecodedSegmentation allows user to specify the RLE-decoded segmentation
func WithDecodedSegmentation(decodedSeg []float64) SegmentationRLEOption {
	return func(s *SegmentRLE) {
		s.decodedSeg = decodedSeg
	}
}

// WithZIndex allows user to specify the z-index of the RLE-encoded segmentation
func WithZIndex(zIndex float64) SegmentationRLEOption {
	return func(s *SegmentRLE) {
		s.zIndex = zIndex
	}
}

// WithTIndex allows user to specify the z-index of the RLE-encoded segmentation
func WithTIndex(tIndex float64) SegmentationRLEOption {
	return func(s *SegmentRLE) {
		s.tIndex = tIndex
	}
}

// WithPixVal allows user to specify the pixel value of the encoded segment
func WithPixVal(pixVal float64) SegmentationRLEOption {
	return func(s *SegmentRLE) {
		s.pixVal = pixVal
	}
}

func NewAnnotationRLE(opts ...SegmentationRLEOption) *SegmentRLE {
	res := &SegmentRLE{}

	for _, opt := range opts {
		opt(res)
	}

	return res
}

func (a *SegmentRLE) Decode() {
	var deflatedSegment []float64

	for idx, segmentLength := range a.encodedSeg {
		var s []float64
		s = make([]float64, int(segmentLength))
		if idx%2 == 0 {
			for i := range s {
				s[i] = 0
			}
		} else {
			for i := range s {
				s[i] = a.pixVal
			}
		}
		deflatedSegment = append(deflatedSegment, s...)
	}
	a.decodedSeg = deflatedSegment
}

func (a *SegmentRLE) Encode() error {
	encodedSegment, err := RLEEncode(a.decodedSeg)
	if err != nil {
		return err
	}
	a.encodedSeg = encodedSegment
	return nil
}
