package gonii

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/okieraised/gonii/pkg/nifti"
	"github.com/stretchr/testify/assert"
)

func TestNewSegmentation_Json_1(t *testing.T) {
	assert := assert.New(t)

	filePath := "/home/tripg/workspace/gonii_test/int16_seg_single.nii.gz"
	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(false))
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
	_ = os.WriteFile("/home/tripg/workspace/gonii_test/coord.json", file, 0777)
}

func TestNewSegmentation_Json_2(t *testing.T) {
	assert := assert.New(t)
	filePath := "/home/tripg/workspace/gonii_test/int16_seg_9223_2.nii.gz"

	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(false))
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

func TestNewSegmentation_Nii_1(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"
	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(true))
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

func TestNewSegmentation_Nii_2(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"
	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	seg1 := []float64{10657, 7, 215, 7, 9, 11, 211, 11, 6, 13, 209, 13, 5, 13, 2, 7, 200, 13, 4, 25, 197, 15, 3, 26, 196,
		15, 3, 26, 179, 7, 10, 15, 3, 27, 176, 11, 8, 15, 3, 27, 175, 13, 7, 15, 3, 31, 171, 13, 7, 15, 3, 33, 168, 15,
		6, 15, 4, 33, 167, 15, 7, 13, 5, 33, 167, 15, 7, 13, 6, 33, 166, 15, 8, 11, 9, 7, 2, 22, 166, 15, 10, 7, 20, 22,
		166, 15, 38, 21, 166, 15, 40, 19, 8, 7, 152, 13, 45, 15, 6, 11, 150, 13, 45, 15, 5, 13, 150, 11, 47, 13, 6, 13,
		24, 8, 120, 7, 49, 33, 21, 12, 99, 7, 69, 32, 20, 14, 96, 11, 69, 30, 20, 14, 95, 13, 74, 24, 19, 16, 94, 13,
		73, 25, 19, 16, 93, 15, 72, 25, 19, 16, 93, 15, 72, 25, 18, 17, 93, 15, 72, 25, 18, 17, 93, 15, 72, 25, 18, 17,
		93, 15, 50, 7, 15, 25, 18, 17, 93, 15, 48, 11, 13, 25, 18, 16, 94, 15, 47, 13, 13, 24, 18, 16, 95, 13, 48, 13,
		13, 25, 17, 16, 95, 13, 47, 15, 13, 24, 18, 15, 96, 11, 47, 16, 15, 22, 18, 15, 98, 7, 48, 18, 6, 7, 4, 19, 17,
		16, 153, 33, 1, 21, 16, 16, 152, 56, 16, 15, 153, 56, 16, 15, 152, 58, 15, 15, 89, 7, 56, 58, 15, 15, 87, 11, 1,
		7, 45, 60, 13, 16, 86, 22, 42, 61, 13, 15, 87, 23, 41, 61, 12, 16, 86, 24, 40, 62, 12, 15, 87, 25, 39, 63, 11,
		15, 87, 25, 39, 63, 11, 15, 87, 25, 39, 64, 10, 15, 87, 25, 37, 66, 10, 15, 87, 25, 36, 67, 9, 16, 87, 25, 35,
		68, 9, 15, 89, 24, 34, 69, 9, 15, 89, 23, 34, 70, 9, 15, 90, 22, 34, 70, 9, 15, 92, 7, 1, 11, 34, 71, 9, 15,
		102, 7, 36, 72, 8, 15, 145, 72, 8, 14, 146, 73, 6, 15, 142, 77, 6, 14, 141, 79, 5, 15, 140, 80, 5, 15, 140, 80,
		5, 15, 139, 81, 5, 15, 139, 81, 5, 15, 139, 80, 6, 15, 139, 81, 5, 15, 139, 20, 1, 61, 5, 13, 140, 20, 1, 61, 4,
		14, 134, 89, 2, 14, 95, 7, 31, 91, 2, 13, 94, 11, 28, 92, 1, 15, 92, 13, 27, 92, 1, 15, 92, 13, 26, 93, 1, 15,
		91, 15, 25, 93, 1, 15, 91, 15, 25, 93, 1, 15, 91, 15, 25, 92, 2, 15, 91, 15, 25, 92, 2, 15, 91, 15, 25, 91, 4,
		13, 92, 15, 25, 108, 92, 15, 26, 106, 94, 13, 27, 104, 96, 13, 28, 104, 96, 11, 30, 103, 98, 7, 32, 103, 138,
		102, 140, 100, 141, 99, 105, 7, 29, 99, 103, 11, 28, 97, 103, 13, 27, 97, 103, 13, 28, 101, 97, 15, 27, 109, 2,
		19, 68, 15, 27, 140, 58, 15, 27, 144, 54, 15, 27, 146, 52, 15, 27, 147, 51, 15, 27, 147, 51, 15, 27, 82, 1, 65,
		51, 13, 29, 80, 9, 58, 51, 13, 29, 80, 8, 59, 52, 11, 30, 79, 9, 59, 54, 7, 31, 80, 9, 59, 63, 7, 22, 80, 9, 59,
		61, 11, 19, 82, 8, 59, 60, 13, 18, 83, 7, 58, 61, 13, 18, 83, 7, 58, 60, 18, 14, 84, 6, 15, 2, 7, 12, 21, 61, 20,
		12, 84, 6, 15, 30, 10, 63, 21, 11, 84, 6, 15, 103, 21, 11, 84, 6, 15, 103, 23, 10, 83, 6, 15, 103, 25, 5, 86, 6,
		15, 103, 26, 2, 88, 6, 15, 104, 25, 1, 88, 7, 15, 104, 114, 7, 15, 105, 113, 7, 15, 105, 113, 7, 15, 105, 113, 8,
		13, 106, 113, 8, 13, 106, 113, 7, 13, 40, 8, 60, 112, 7, 13, 38, 12, 58, 111, 7, 15, 34, 16, 58, 22, 1, 87, 7, 15,
		32, 18, 60, 7, 1, 11, 2, 86, 8, 15, 28, 23, 69, 7, 4, 73, 4, 7, 10, 15, 24, 27, 81, 72, 21, 15, 12, 9, 1, 29, 81,
		72, 21, 15, 10, 41, 81, 72, 21, 15, 1, 50, 81, 77, 17, 65, 81, 79, 15, 65, 81, 80, 14, 65, 82, 79, 14, 64, 83, 80,
		13, 64, 84, 79, 13, 63, 87, 59, 3, 15, 13, 61, 95, 52, 4, 15, 13, 57, 99, 50, 4, 17, 13, 54, 103, 41, 10, 19, 13,
		52, 105, 42, 8, 20, 13, 48, 108, 50, 1, 19, 14, 44, 112, 70, 14, 42, 113, 70, 15, 30, 125, 68, 17, 19, 136, 20,
		1, 7, 2, 36, 19, 15, 140, 19, 11, 36, 19, 15, 140, 17, 13, 36, 19, 15, 140, 15, 15, 36, 19, 15, 140, 15, 16, 34,
		20, 15, 141, 13, 17, 34, 20, 15, 141, 13, 18, 32, 20, 16, 142, 13, 19, 20, 1, 7, 22, 15, 143, 13, 26, 13, 30, 15,
		142, 15, 26, 11, 31, 15, 142, 15, 28, 7, 33, 15, 142, 15, 68, 15, 142, 15, 68, 15, 142, 15, 68, 15, 142, 15, 68,
		15, 142, 15, 68, 16, 142, 13, 69, 16, 142, 13, 69, 16, 142, 13, 69, 16, 141, 15, 68, 16, 141, 15, 68, 16, 141, 15,
		68, 16, 141, 16, 67, 15, 142, 16, 66, 16, 142, 17, 65, 16, 142, 17, 65, 16, 143, 16, 65, 16, 143, 16, 65, 16, 143,
		16, 65, 16, 143, 16, 65, 16, 142, 17, 66, 15, 142, 16, 67, 14, 143, 16, 68, 13, 143, 16, 69, 11, 144, 16, 71, 7,
		146, 16, 224, 16, 224, 16, 224, 15, 225, 15, 226, 13, 227, 13, 228, 11, 231, 7, 2071}
	seg2 := []float64{12933, 8, 226, 16, 222, 19, 220, 22, 216, 25, 214, 28, 211, 30, 209, 32, 207, 34, 205, 36, 204,
		37, 202, 39, 201, 40, 199, 43, 197, 44, 195, 46, 67, 17, 110, 46, 63, 24, 106, 47, 61, 28, 104, 47, 60, 30, 103,
		47, 58, 33, 102, 47, 57, 36, 99, 48, 56, 39, 97, 48, 55, 41, 96, 48, 54, 43, 95, 48, 53, 43, 96, 48, 52, 43, 97,
		48, 52, 42, 98, 15, 2, 31, 51, 43, 98, 16, 1, 31, 51, 42, 99, 17, 1, 30, 51, 42, 99, 48, 50, 42, 100, 48, 50, 42,
		100, 48, 50, 41, 101, 48, 50, 40, 103, 47, 49, 41, 103, 47, 49, 40, 105, 46, 50, 39, 105, 46, 50, 39, 106, 45, 50,
		39, 107, 44, 50, 37, 110, 43, 50, 36, 112, 42, 50, 35, 115, 40, 50, 34, 117, 38, 50, 34, 119, 36, 51, 34, 120, 34,
		52, 33, 123, 30, 54, 33, 124, 27, 56, 33, 126, 23, 58, 33, 128, 19, 60, 29, 134, 14, 63, 27, 213, 26, 215, 25, 215,
		24, 216, 24, 217, 23, 217, 23, 217, 23, 217, 23, 217, 17, 223, 15, 225, 14, 226, 14, 227, 12, 228, 12, 229, 11,
		230, 10, 231, 9, 232, 8, 234, 6, 28016}
	zIndex := 32

	encoded1 := nifti.SegmentRLE{
		EncodedSeg: seg1,
		ZIndex:     float64(zIndex),
		TIndex:     0,
		PixVal:     1,
	}

	encoded2 := nifti.SegmentRLE{
		EncodedSeg: seg2,
		ZIndex:     float64(zIndex),
		TIndex:     0,
		PixVal:     2,
	}

	encoded := []nifti.SegmentRLE{encoded1, encoded2}

	voxels, err := rd.GetNiiData().GetVoxels().ExportSingleFromRLE(encoded)
	assert.NoError(err)

	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("/home/tripg/workspace/int16_seg.nii.gz",
		WithWriteNIfTIData(rd.GetNiiData()),
		WithWriteCompression(true),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}

func TestNewSegmentation_Nii_3(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"
	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)
	fmt.Println(rd.GetNiiData().GetVoxels().CountNoneZero())

	seg1 := []float64{10657, 7, 215, 7, 9, 11, 211, 11, 6, 13, 209, 13, 5, 13, 2, 7, 200, 13, 4, 25, 197, 15, 3, 26, 196,
		15, 3, 26, 179, 7, 10, 15, 3, 27, 176, 11, 8, 15, 3, 27, 175, 13, 7, 15, 3, 31, 171, 13, 7, 15, 3, 33, 168, 15,
		6, 15, 4, 33, 167, 15, 7, 13, 5, 33, 167, 15, 7, 13, 6, 33, 166, 15, 8, 11, 9, 7, 2, 22, 166, 15, 10, 7, 20, 22,
		166, 15, 38, 21, 166, 15, 40, 19, 8, 7, 152, 13, 45, 15, 6, 11, 150, 13, 45, 15, 5, 13, 150, 11, 47, 13, 6, 13,
		24, 8, 120, 7, 49, 33, 21, 12, 99, 7, 69, 32, 20, 14, 96, 11, 69, 30, 20, 14, 95, 13, 74, 24, 19, 16, 94, 13,
		73, 25, 19, 16, 93, 15, 72, 25, 19, 16, 93, 15, 72, 25, 18, 17, 93, 15, 72, 25, 18, 17, 93, 15, 72, 25, 18, 17,
		93, 15, 50, 7, 15, 25, 18, 17, 93, 15, 48, 11, 13, 25, 18, 16, 94, 15, 47, 13, 13, 24, 18, 16, 95, 13, 48, 13,
		13, 25, 17, 16, 95, 13, 47, 15, 13, 24, 18, 15, 96, 11, 47, 16, 15, 22, 18, 15, 98, 7, 48, 18, 6, 7, 4, 19, 17,
		16, 153, 33, 1, 21, 16, 16, 152, 56, 16, 15, 153, 56, 16, 15, 152, 58, 15, 15, 89, 7, 56, 58, 15, 15, 87, 11, 1,
		7, 45, 60, 13, 16, 86, 22, 42, 61, 13, 15, 87, 23, 41, 61, 12, 16, 86, 24, 40, 62, 12, 15, 87, 25, 39, 63, 11,
		15, 87, 25, 39, 63, 11, 15, 87, 25, 39, 64, 10, 15, 87, 25, 37, 66, 10, 15, 87, 25, 36, 67, 9, 16, 87, 25, 35,
		68, 9, 15, 89, 24, 34, 69, 9, 15, 89, 23, 34, 70, 9, 15, 90, 22, 34, 70, 9, 15, 92, 7, 1, 11, 34, 71, 9, 15,
		102, 7, 36, 72, 8, 15, 145, 72, 8, 14, 146, 73, 6, 15, 142, 77, 6, 14, 141, 79, 5, 15, 140, 80, 5, 15, 140, 80,
		5, 15, 139, 81, 5, 15, 139, 81, 5, 15, 139, 80, 6, 15, 139, 81, 5, 15, 139, 20, 1, 61, 5, 13, 140, 20, 1, 61, 4,
		14, 134, 89, 2, 14, 95, 7, 31, 91, 2, 13, 94, 11, 28, 92, 1, 15, 92, 13, 27, 92, 1, 15, 92, 13, 26, 93, 1, 15,
		91, 15, 25, 93, 1, 15, 91, 15, 25, 93, 1, 15, 91, 15, 25, 92, 2, 15, 91, 15, 25, 92, 2, 15, 91, 15, 25, 91, 4,
		13, 92, 15, 25, 108, 92, 15, 26, 106, 94, 13, 27, 104, 96, 13, 28, 104, 96, 11, 30, 103, 98, 7, 32, 103, 138,
		102, 140, 100, 141, 99, 105, 7, 29, 99, 103, 11, 28, 97, 103, 13, 27, 97, 103, 13, 28, 101, 97, 15, 27, 109, 2,
		19, 68, 15, 27, 140, 58, 15, 27, 144, 54, 15, 27, 146, 52, 15, 27, 147, 51, 15, 27, 147, 51, 15, 27, 82, 1, 65,
		51, 13, 29, 80, 9, 58, 51, 13, 29, 80, 8, 59, 52, 11, 30, 79, 9, 59, 54, 7, 31, 80, 9, 59, 63, 7, 22, 80, 9, 59,
		61, 11, 19, 82, 8, 59, 60, 13, 18, 83, 7, 58, 61, 13, 18, 83, 7, 58, 60, 18, 14, 84, 6, 15, 2, 7, 12, 21, 61, 20,
		12, 84, 6, 15, 30, 10, 63, 21, 11, 84, 6, 15, 103, 21, 11, 84, 6, 15, 103, 23, 10, 83, 6, 15, 103, 25, 5, 86, 6,
		15, 103, 26, 2, 88, 6, 15, 104, 25, 1, 88, 7, 15, 104, 114, 7, 15, 105, 113, 7, 15, 105, 113, 7, 15, 105, 113, 8,
		13, 106, 113, 8, 13, 106, 113, 7, 13, 40, 8, 60, 112, 7, 13, 38, 12, 58, 111, 7, 15, 34, 16, 58, 22, 1, 87, 7,
		15, 32, 18, 60, 7, 1, 11, 2, 86, 8, 15, 28, 23, 69, 7, 4, 73, 4, 7, 10, 15, 24, 27, 81, 72, 21, 15, 12, 9, 1, 29,
		81, 72, 21, 15, 10, 41, 81, 72, 21, 15, 1, 50, 81, 77, 17, 65, 81, 79, 15, 65, 81, 80, 14, 65, 82, 79, 14, 64,
		83, 80, 13, 64, 84, 79, 13, 63, 87, 59, 3, 15, 13, 61, 95, 52, 4, 15, 13, 57, 99, 50, 4, 17, 13, 54, 103, 41,
		10, 19, 13, 52, 105, 42, 8, 20, 13, 48, 108, 50, 1, 19, 14, 44, 112, 70, 14, 42, 113, 70, 15, 30, 125, 68, 17,
		19, 136, 20, 1, 7, 2, 36, 19, 15, 140, 19, 11, 36, 19, 15, 140, 17, 13, 36, 19, 15, 140, 15, 15, 36, 19, 15, 140,
		15, 16, 34, 20, 15, 141, 13, 17, 34, 20, 15, 141, 13, 18, 32, 20, 16, 142, 13, 19, 20, 1, 7, 22, 15, 143, 13, 26,
		13, 30, 15, 142, 15, 26, 11, 31, 15, 142, 15, 28, 7, 33, 15, 142, 15, 68, 15, 142, 15, 68, 15, 142, 15, 68, 15,
		142, 15, 68, 15, 142, 15, 68, 16, 142, 13, 69, 16, 142, 13, 69, 16, 142, 13, 69, 16, 141, 15, 68, 16, 141, 15,
		68, 16, 141, 15, 68, 16, 141, 16, 67, 15, 142, 16, 66, 16, 142, 17, 65, 16, 142, 17, 65, 16, 143, 16, 65, 16,
		143, 16, 65, 16, 143, 16, 65, 16, 143, 16, 65, 16, 142, 17, 66, 15, 142, 16, 67, 14, 143, 16, 68, 13, 143, 16,
		69, 11, 144, 16, 71, 7, 146, 16, 224, 16, 224, 16, 224, 15, 225, 15, 226, 13, 227, 13, 228, 11, 231, 7, 2071}
	seg2 := []float64{12933, 8, 226, 16, 222, 19, 220, 22, 216, 25, 214, 28, 211, 30, 209, 32, 207, 34, 205, 36, 204, 37,
		202, 39, 201, 40, 199, 43, 197, 44, 195, 46, 67, 17, 110, 46, 63, 24, 106, 47, 61, 28, 104, 47, 60, 30, 103, 47,
		58, 33, 102, 47, 57, 36, 99, 48, 56, 39, 97, 48, 55, 41, 96, 48, 54, 43, 95, 48, 53, 43, 96, 48, 52, 43, 97, 48,
		52, 42, 98, 15, 2, 31, 51, 43, 98, 16, 1, 31, 51, 42, 99, 17, 1, 30, 51, 42, 99, 48, 50, 42, 100, 48, 50, 42,
		100, 48, 50, 41, 101, 48, 50, 40, 103, 47, 49, 41, 103, 47, 49, 40, 105, 46, 50, 39, 105, 46, 50, 39, 106, 45,
		50, 39, 107, 44, 50, 37, 110, 43, 50, 36, 112, 42, 50, 35, 115, 40, 50, 34, 117, 38, 50, 34, 119, 36, 51, 34,
		120, 34, 52, 33, 123, 30, 54, 33, 124, 27, 56, 33, 126, 23, 58, 33, 128, 19, 60, 29, 134, 14, 63, 27, 213, 26,
		215, 25, 215, 24, 216, 24, 217, 23, 217, 23, 217, 23, 217, 23, 217, 17, 223, 15, 225, 14, 226, 14, 227, 12, 228,
		12, 229, 11, 230, 10, 231, 9, 232, 8, 234, 6, 28016}
	zIndex := 32

	encoded1 := nifti.SegmentRLE{
		EncodedSeg: seg1,
		ZIndex:     float64(zIndex),
		TIndex:     0,
		PixVal:     1,
	}

	encoded2 := nifti.SegmentRLE{
		EncodedSeg: seg2,
		ZIndex:     float64(zIndex),
		TIndex:     0,
		PixVal:     2,
	}

	encoded := []nifti.SegmentRLE{encoded1, encoded2}
	//encoded = []nifti.SegmentRLE{encoded1}

	voxels, err := rd.GetNiiData().GetVoxels().ExportSingleFromRLE(encoded)
	assert.NoError(err)

	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("/home/tripg/workspace/int16_seg.nii.gz",
		WithWriteNIfTIData(rd.GetNiiData()),
		WithWriteCompression(true),
	)
	b, err := writer.WriteToBytes()
	assert.NoError(err)

	fmt.Println(len(b))
}

func TestSegmentation_Annotation3(t *testing.T) {

	y := []float64{1, 3, 4, 6, 8, 2, 532, 6234, 5, 75, 75, 656}
	var res string
	for _, e := range y {
		res += fmt.Sprintf("%d ", int(e))
	}

	fmt.Println("res", strings.TrimSpace(res))
}

func TestSegmentation_Annotation4(t *testing.T) {
	assert := assert.New(t)

	filePath := "/home/tripg/Downloads/annot_import/CT_Abdo.1680080052.TSK-7.seg.nii.gz"
	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()
	voxels.FlipX()
	voxels.FlipY()
	voxels.FlipZ()
	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	segments, err := rd.GetNiiData().GetVoxels().ImportAsRLE()
	assert.NoError(err)

	voxels, err = rd.GetNiiData().GetVoxels().ExportSingleFromRLE(segments)
	assert.NoError(err)

	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("/home/tripg/Downloads/reexport.nii.gz",
		WithWriteNIfTIData(rd.GetNiiData()),
		WithWriteCompression(true),
	)
	err = writer.WriteToFile()
	assert.NoError(err)

}

func TestSegmentation_Annotation5(t *testing.T) {
	assert := assert.New(t)

	filePath := "/home/tripg/Downloads/annot_import/CT_Abdo.1680080052.TSK-7.seg.nii.gz"
	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()
	voxels.FlipX()
	voxels.FlipY()
	voxels.FlipZ()
	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	segments, err := rd.GetNiiData().GetVoxels().ImportAsRLE()
	assert.NoError(err)

	for _, segment := range segments {
		fmt.Println(segment.EncodedSeg)
	}
}
