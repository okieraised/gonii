package gonii

import (
	"encoding/json"
	"fmt"
	"github.com/okieraised/gonii/pkg/nifti"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestAnnotationJsonToNii(t *testing.T) {
	assert := assert.New(t)

	filePath := "/home/tripg/workspace/gonii_test/int16.nii.gz"
	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	bJson, err := os.ReadFile("/home/tripg/workspace/gonii_test/coord.json")
	assert.NoError(err)

	annotations := []SegmentCoordinate{}

	err = json.Unmarshal(bJson, &annotations)
	assert.NoError(err)

	err = AnnotationJsonToNii(annotations,
		WithNii1Hdr(rd.GetHeader(false).(*nifti.Nii1Header)),
		WithSegCompression(true),
		WithOutFile("/home/tripg/workspace/gonii_test/int16_seg_9223.nii.gz"))
	assert.NoError(err)

}

func TestNewNiiWriter_MakeSegmentation_ToJson(t *testing.T) {
	assert := assert.New(t)

	filePath := "/home/tripg/workspace/gonii_test/int16_seg_single.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()
	var res []SegmentCoordinate

	fmt.Println(voxels.CountNoneZero())

	fmt.Println(rd.GetNiiData().Nx, rd.GetNiiData().Ny, rd.GetNiiData().Nz, rd.GetNiiData().Nt)

	for x := int64(0); x < rd.GetNiiData().Nx; x++ {
		for y := int64(0); y < rd.GetNiiData().Ny; y++ {
			for z := int64(0); z < rd.GetNiiData().Nz; z++ {
				for tt := int64(0); tt < rd.GetNiiData().Nt; tt++ {
					val := voxels.Get(x, y, z, tt)
					if val > 0 {
						coord := SegmentCoordinate{
							X:     x,
							Y:     y,
							Z:     z,
							T:     tt,
							Value: int64(val),
						}
						res = append(res, coord)
					}
				}
			}
		}
	}

	file, _ := json.MarshalIndent(res, "", " ")

	_ = ioutil.WriteFile("/home/tripg/workspace/gonii_test/coord.json", file, 0777)
}

func TestNewNiiWriter_MultipleSegmentations(t *testing.T) {
	assert := assert.New(t)

	filePath := "/home/tripg/workspace/gonii_test/corie_seg.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()

	voxelValMapper := map[float64]bool{}
	for _, voxel := range voxels.GetDataset() {
		_, ok := voxelValMapper[voxel]
		if !ok {
			voxelValMapper[voxel] = true
		}
	}
	fmt.Println(voxelValMapper)
}
