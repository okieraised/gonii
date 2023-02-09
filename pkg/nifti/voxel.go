package nifti

import "github.com/okieraised/gonii/internal/utils"

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
