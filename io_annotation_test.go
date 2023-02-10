package gonii

import (
	"encoding/json"
	"github.com/okieraised/gonii/pkg/nifti"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewNiiWriter_MakeSegmentation_ToJson(t *testing.T) {
	assert := assert.New(t)

	filePath := "/home/tripg/workspace/gonii_test/int16_seg_single.nii.gz"
	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()
	var res []SegmentCoordinate

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

func TestNewSegmentation_NiiToJson(t *testing.T) {
	assert := assert.New(t)
	filePath := "/home/tripg/workspace/gonii_test/int16_seg_9223_2.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	s := NewSegmentation(
		WithImage(rd.GetNiiData()),
		WithOutFile("/home/tripg/workspace/gonii_test/seg_out.json"),
	)

	err = s.AnnotationNiiToJson()
	assert.NoError(err)

}

func TestNewSegmentation_JsonToNii(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"
	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	bJson, err := os.ReadFile("/home/tripg/workspace/gonii_test/coord.json")
	assert.NoError(err)

	annotations := []SegmentCoordinate{}

	err = json.Unmarshal(bJson, &annotations)
	assert.NoError(err)

	s := NewSegmentation(
		WithAnnotations(annotations),
		WithNii1Hdr(rd.GetHeader(false).(*nifti.Nii1Header)),
		WithSegCompression(true),
		WithOutFile("/home/tripg/workspace/gonii_test/int16_seg_10223_2.nii.gz"),
	)

	err = s.AnnotationJsonToNii()
	assert.NoError(err)
}
