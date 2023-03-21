package nifti

import (
	"errors"
	"fmt"
	"github.com/okieraised/gonii/internal/utils"
)

// Voxels defines the structure of Voxel values
type Voxels struct {
	voxel                  []float64
	dimX, dimY, dimZ, dimT int64
	datatype               int32
}

func NewVoxels(dimX, dimY, dimZ, dimT int64, datatype int32) *Voxels {
	voxel := make([]float64, dimX*dimY*dimZ*dimT)
	return &Voxels{
		voxel:    voxel,
		dimX:     dimX,
		dimY:     dimY,
		dimZ:     dimZ,
		dimT:     dimT,
		datatype: datatype,
	}
}

// Set sets the value of voxel at index calculated from x, y, z, t input
func (v *Voxels) Set(x, y, z, t int64, val float64) {
	idx := t*v.dimZ*v.dimY*v.dimX + z*v.dimY*v.dimX + y*v.dimX + x
	v.voxel[idx] = val
}

// Get returns the value of voxel at index calculated from x, y, z, t input
func (v *Voxels) Get(x, y, z, t int64) float64 {
	idx := t*v.dimZ*v.dimY*v.dimX + z*v.dimY*v.dimX + y*v.dimX + x
	return v.voxel[idx]
}

// GetSlice returns the values of voxel as a 1-D slice of float64 calculated from z, t input
func (v *Voxels) GetSlice(z, t int64) []float64 {
	res := make([]float64, 0)
	for x := int64(0); x < v.dimX; x++ {
		for y := int64(0); y < v.dimY; y++ {
			res = append(res, v.Get(x, y, z, t))
		}
	}
	return res
}

// GetVolume returns the values of voxel as a 1-D slice of float64 calculated from t input
func (v *Voxels) GetVolume(t int64) []float64 {
	res := make([]float64, 0)
	for x := int64(0); x < v.dimX; x++ {
		for y := int64(0); y < v.dimY; y++ {
			for z := int64(0); z < v.dimZ; z++ {
				res = append(res, v.Get(x, y, z, t))
			}
		}
	}
	return res
}

func (v *Voxels) Len() int {
	return len(v.voxel)
}

func (v *Voxels) GetDataset() []float64 {
	return v.voxel
}

func (v *Voxels) GetRawByteSize() int {
	nByPer, _ := AssignDatatypeSize(v.datatype)
	return int(v.dimX*v.dimY*v.dimZ*v.dimT) * int(nByPer)
}

func (v *Voxels) CountNoneZero() (pos, neg, zero int) {
	for _, vox := range v.voxel {
		if vox > 0 {
			pos++
		} else if vox < 0 {
			neg++
		} else {
			zero++
		}
	}
	return pos, neg, zero
}

// Histogram returns the histogram of the voxels based on the input bins
func (v *Voxels) Histogram(bins int) (utils.Histogram, error) {
	return utils.Hist(bins, v.voxel)
}

// RLEEncode encodes the 1-D float64 array using the RLE encoding
func (v *Voxels) RLEEncode() ([]float64, error) {
	//v.voxel = []float64{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 11, 11, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	return RLEEncode(v.voxel)
}

// MapValueOccurrence maps the occurrence of pixel value as map[float64]int
func (v *Voxels) MapValueOccurrence() map[float64]int {
	valMapper := make(map[float64]int)
	for _, val := range v.voxel {
		valMapper[val] = valMapper[val] + 1
	}
	return valMapper
}

// ImportAsRLE import the NIfTI image as an array of RLE-encoded segment
func (v *Voxels) ImportAsRLE() ([]SegmentRLE, error) {
	valMapper := v.MapValueOccurrence()
	var result []SegmentRLE

	for z := int64(0); z < v.dimZ; z++ {
		for t := int64(0); t < v.dimT; t++ {
			sliceData := v.GetSlice(z, t)
			for key, _ := range valMapper {
				if key == 0 {
					continue
				}
				keyArr := make([]float64, len(sliceData))
				for idx, voxVal := range sliceData {
					if voxVal == key {
						keyArr[idx] = key
					}
				}
				encoded, err := RLEEncode(keyArr)
				if err != nil {
					return nil, err
				}

				encodedSegment := SegmentRLE{
					EncodedSeg: encoded,
					DecodedSeg: sliceData,
					ZIndex:     float64(z),
					TIndex:     float64(t),
					PixVal:     key,
				}
				result = append(result, encodedSegment)
			}
		}
	}
	return result, nil
}

// ExportSingleFromRLE reconstruct a single NIfTI image from input RLE-encoded 1-D segments
func (v *Voxels) ExportSingleFromRLE(segments []SegmentRLE) (*Voxels, error) {

	if len(segments) == 0 {
		return v, errors.New("segments has length 0")
	}

	originalLength := make([]float64, len(v.voxel))
	initSegment := segments[0]
	fmt.Println("initSegment", initSegment.PixVal)
	initSegment.Decode()

	if len(segments) == 1 {
		v.voxel = initSegment.DecodedSeg
		fmt.Println("1", v.MapValueOccurrence())
		return v, nil
	}

	//for _, segment := range segments {
	//	segment.Decode()
	//	for x := int64(0); x < v.dimX; x++ {
	//		for y := int64(0); y < v.dimY; y++ {
	//			for z := int64(0); z < v.dimZ; z++ {
	//				for t := int64(0); t < v.dimT; t++ {
	//					if z == int64(segment.ZIndex) {
	//						idx := t*v.dimZ*v.dimY*v.dimX + z*v.dimY*v.dimX + y*v.dimX + x
	//						originalLength[idx] = segment.PixVal
	//					}
	//				}
	//			}
	//		}
	//	}
	//}

	v.voxel = originalLength

	fmt.Println("1", v.MapValueOccurrence())

	return nil, nil
}
