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
	rd, err := NewNiiReader(filePath, WithRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	bJson, err := os.ReadFile("/home/tripg/workspace/gonii_test/coord.json")
	assert.NoError(err)

	annotations := []SegmentCoordinate{}

	err = json.Unmarshal(bJson, &annotations)
	assert.NoError(err)

	//fmt.Println(annotations)

	err = AnnotationJsonToNii(annotations, WithNii1Hdr(rd.GetHeader(false).(*nifti.Nii1Header)))
	assert.NoError(err)

}

func TestNewNiiWriter_MakeSegmentation_ToJson(t *testing.T) {
	assert := assert.New(t)

	filePath := "/home/tripg/workspace/gonii_test/int16_seg_single.nii.gz"

	rd, err := NewNiiReader(filePath, WithRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()
	res := []SegmentCoordinate{}

	//for _, voxel := range voxels.GetDataset() {
	//	if voxel > 0 {
	//		fmt.Println(voxel)
	//	}
	//}
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

	//fmt.Println(res)

	file, _ := json.MarshalIndent(res, "", " ")

	_ = ioutil.WriteFile("/home/tripg/workspace/gonii_test/coord.json", file, 0777)

	//for index, voxel := range voxels.GetDataset() {
	//	if voxel > 200 {
	//		voxels.GetDataset()[index] = 1
	//	} else {
	//		voxels.GetDataset()[index] = 0
	//	}
	//}
	//
	//err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	//assert.NoError(err)
	//
	//writer, err := NewNiiWriter("/home/tripg/workspace/gonii_test/int16_seg_single.nii.gz",
	//	WithNIfTIData(rd.GetNiiData()),
	//	WithCompression(true),
	//)
	//err = writer.WriteToFile()
	//assert.NoError(err)
}
