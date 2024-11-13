package nifti

import (
	"errors"

	"github.com/okieraised/gonii/internal/utils"
)

// Voxels defines the structure of Voxel values
type Voxels struct {
	voxel                  []float64
	dimX, dimY, dimZ, dimT int64
	datatype               int32
}

// NewVoxels returns a pointer to the Voxels with specified input parameters
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

// Flip flips the image along the specified axes
func (v *Voxels) Flip(flipX, flipY, flipZ bool) *Voxels {
	if flipX {
		v.FlipX()
	}
	if flipY {
		v.FlipY()
	}
	if flipZ {
		v.FlipZ()
	}

	return v
}

// FlipSagittal flips the image along Sagittal-axis (Y-Z)
func (v *Voxels) FlipSagittal() {
	v.FlipY()
	v.FlipZ()
}

// FlipCoronal flips the image along Coronal-axis (Y-X)
func (v *Voxels) FlipCoronal() {
	v.FlipY()
	v.FlipX()
}

// FlipAxial flips the image along Axial-axis (X-Z)
func (v *Voxels) FlipAxial() {
	v.FlipX()
	v.FlipZ()
}

// FlipX flips the image along the x-axis
func (v *Voxels) FlipX() {
	for x := int64(0); x < v.dimX/2; x++ {
		k := v.dimX - 1 - x
		for y := int64(0); y < v.dimY; y++ {
			for z := int64(0); z < v.dimZ; z++ {
				for t := int64(0); t < v.dimT; t++ {
					idx := t*v.dimZ*v.dimY*v.dimX + z*v.dimY*v.dimX + y*v.dimX + x
					temp := v.voxel[idx]
					v.voxel[idx] = v.voxel[t*v.dimZ*v.dimY*v.dimX+z*v.dimY*v.dimX+y*v.dimX+k]
					v.voxel[t*v.dimZ*v.dimY*v.dimX+z*v.dimY*v.dimX+y*v.dimX+k] = temp
				}
			}
		}
	}
}

// FlipY flips the image along the y-axis
func (v *Voxels) FlipY() {
	for y := int64(0); y < v.dimY/2; y++ {
		k := v.dimY - 1 - y
		for x := int64(0); x < v.dimX; x++ {
			for z := int64(0); z < v.dimZ; z++ {
				for t := int64(0); t < v.dimT; t++ {
					idx := t*v.dimZ*v.dimY*v.dimX + z*v.dimY*v.dimX + y*v.dimX + x
					temp := v.voxel[idx]
					v.voxel[idx] = v.voxel[t*v.dimZ*v.dimY*v.dimX+z*v.dimY*v.dimX+k*v.dimX+x]
					v.voxel[t*v.dimZ*v.dimY*v.dimX+z*v.dimY*v.dimX+k*v.dimX+x] = temp
				}
			}
		}
	}
}

// FlipZ flips the image along the z-axis
func (v *Voxels) FlipZ() {
	for z := int64(0); z < v.dimZ/2; z++ {
		k := v.dimZ - 1 - z
		for x := int64(0); x < v.dimX; x++ {
			for y := int64(0); y < v.dimY; y++ {
				for t := int64(0); t < v.dimT; t++ {
					idx := t*v.dimZ*v.dimY*v.dimX + z*v.dimY*v.dimX + y*v.dimX + x
					temp := v.voxel[idx]
					v.voxel[idx] = v.voxel[t*v.dimZ*v.dimY*v.dimX+k*v.dimY*v.dimX+y*v.dimX+x]
					v.voxel[t*v.dimZ*v.dimY*v.dimX+k*v.dimY*v.dimX+y*v.dimX+x] = temp
				}
			}
		}
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

// GetDimX returns the dimX information
func (v *Voxels) GetDimX() int64 {
	return v.dimX
}

// GetDimY returns the dimY information
func (v *Voxels) GetDimY() int64 {
	return v.dimY
}

// GetDimZ returns the dimZ information
func (v *Voxels) GetDimZ() int64 {
	return v.dimZ
}

// GetDimT returns the dimT information
func (v *Voxels) GetDimT() int64 {
	return v.dimT
}

// GetSlice returns the values of voxel as a 1-D slice of float64 calculated from z, t input
func (v *Voxels) GetSlice(z, t int64) []float64 {
	res := make([]float64, v.dimX*v.dimY)
	for x := int64(0); x < v.dimX; x++ {
		for y := int64(0); y < v.dimY; y++ {
			res[y*v.dimX+x] = v.Get(x, y, z, t)
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
			for key := range valMapper {
				if key == 0 {
					continue
				}
				keyArr := make([]float64, len(sliceData))

				for idx, voxVal := range sliceData {
					if voxVal == key {
						keyArr[len(sliceData)-idx-1] = key
					}
				}
				encoded, err := RLEEncode(keyArr)
				if err != nil {
					return nil, err
				}

				// Skip zero-only segment since they do not contain any segmentation
				if len(encoded) == 1 && encoded[0] == float64(v.dimX*v.dimY) {
					continue
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
	for _, segment := range segments {
		segment.Decode()
		for x := int64(0); x < v.dimX; x++ {
			for y := int64(0); y < v.dimY; y++ {
				for z := int64(0); z < v.dimZ; z++ {
					for t := int64(0); t < v.dimT; t++ {
						if z == v.dimZ-1-int64(segment.ZIndex) {
							idx := t*v.dimZ*v.dimY*v.dimX + z*v.dimY*v.dimX + y*v.dimX + x
							if segment.DecodedSeg[y*v.dimX+x] != 0 {
								originalLength[idx] = segment.PixVal
							}
						}
					}
				}
			}
		}
	}
	v.voxel = originalLength
	return v, nil
}
