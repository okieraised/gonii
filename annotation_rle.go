package gonii

import "github.com/okieraised/gonii/pkg/nifti"

type SegmentationRLE struct {
	encodedSeg []float64
	decodedSeg []float64
	zIndex     float64
	tIndex     float64
	pixVal     float64
}

type SegmentationRLEOption func(s *SegmentationRLE)

// WithEncodedSegmentation allows user to specify the RLE-encoded segmentation
func WithEncodedSegmentation(encodedSeg []float64) SegmentationRLEOption {
	return func(s *SegmentationRLE) {
		s.encodedSeg = encodedSeg
	}
}

// WithDecodedSegmentation allows user to specify the RLE-decoded segmentation
func WithDecodedSegmentation(decodedSeg []float64) SegmentationRLEOption {
	return func(s *SegmentationRLE) {
		s.decodedSeg = decodedSeg
	}
}

// WithZIndex allows user to specify the z-index of the RLE-encoded segmentation
func WithZIndex(zIndex float64) SegmentationRLEOption {
	return func(s *SegmentationRLE) {
		s.zIndex = zIndex
	}
}

// WithTIndex allows user to specify the z-index of the RLE-encoded segmentation
func WithTIndex(tIndex float64) SegmentationRLEOption {
	return func(s *SegmentationRLE) {
		s.tIndex = tIndex
	}
}

// WithPixVal allows user to specify the pixel value of the encoded segment
func WithPixVal(pixVal float64) SegmentationRLEOption {
	return func(s *SegmentationRLE) {
		s.pixVal = pixVal
	}
}

func NewAnnotationRLE(opts ...SegmentationRLEOption) *SegmentationRLE {
	res := &SegmentationRLE{}

	for _, opt := range opts {
		opt(res)
	}

	return res
}

func (a *SegmentationRLE) Decode() {
	var deflatedSegment []float64

	for idx, segmentLength := range a.encodedSeg {
		var s []float64
		s = make([]float64, segmentLength)
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

func (a *SegmentationRLE) Encode() error {
	encodedSegment, err := nifti.RLEEncode(a.decodedSeg)
	if err != nil {
		return err
	}
	a.encodedSeg = encodedSegment
	return nil
}
