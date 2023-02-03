package gonii

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNiiWriter_Segmentation(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	rd, err := NewNiiReader(filePath, WithInMemory(true), WithRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	rd.GetHeader(false)

	//voxels := rd.GetVoxels()
	//
	//for index, voxel := range voxels.GetDataset() {
	//	if voxel > -200 {
	//		voxels.GetDataset()[index] = 1
	//	} else {
	//		voxels.GetDataset()[index] = 0
	//	}
	//}
	//
	//err = rd.SetVoxelToRawVolume(voxels)
	//assert.NoError(err)
	//
	//writer, err := nii_io.NewNiiWriter("/home/tripg/workspace/anim3_out.nii.gz",
	//	nii_io.WithNIfTIData(rd.GetNiiData()),
	//	nii_io.WithCompression(true),
	//)
	//
	//err = writer.WriteToFile()
	//assert.NoError(err)
}
