package gonii

import (
	"encoding/binary"
	"fmt"
	"github.com/okieraised/gonii/pkg/matrix"
	"github.com/okieraised/gonii/pkg/nifti"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNiiReader_Parse_SingleFile_Nii1_Int16(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	fmt.Println(rd.GetNiiData().GetNVox())

	assert.Equal(rd.GetNiiData().GetDatatype(), "INT16")
	assert.Equal(rd.GetNiiData().GetDatatype(), "INT16")
	assert.Equal(rd.GetNiiData().GetImgShape(), [4]int64{240, 240, 155, 1})
	assert.Equal(rd.GetNiiData().GetQFormCode(), "1: Scanner Anat")
	assert.Equal(rd.GetNiiData().GetAffine(), matrix.DMat44{
		M: [4][4]float64{
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{0, 0, 0, 1},
		},
	})

	assert.Equal(rd.GetNiiData().Dim, [8]int64{3, 240, 240, 155, 1, 1, 1, 1})
}

func TestNiiReader_Parse_SingleFile_Nii2_LR(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/nii2_LR.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	rd.GetHeader(false)

	assert.Equal(rd.GetNiiData().GetOrientation(), [3]string{
		nifti.OrietationToString[nifti.NIFTI_R2L],
		nifti.OrietationToString[nifti.NIFTI_P2A],
		nifti.OrietationToString[nifti.NIFTI_I2S],
	})

	assert.Equal(rd.GetBinaryOrder(), binary.LittleEndian)

	assert.Equal(rd.GetNiiData().GetAffine(), matrix.DMat44{
		M: [4][4]float64{
			{-2, 0, 0, 90},
			{0, 2, 0, -126},
			{0, 0, 2, -72},
			{0, 0, 0, 1},
		},
	})
	assert.Equal(rd.GetNiiData().GetDatatype(), "FLOAT32")
	assert.Equal(rd.GetNiiData().GetSFormCode(), "4: MNI")
	assert.Equal(rd.GetNiiData().GetQFormCode(), "0: Unknown")
}

func TestNiiReader_Parse_SingleFile_Nii2_RL(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/nii2_RL.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	rd.GetHeader(false)

	assert.Equal(rd.GetNiiData().GetOrientation(), [3]string{
		nifti.OrietationToString[nifti.NIFTI_L2R],
		nifti.OrietationToString[nifti.NIFTI_P2A],
		nifti.OrietationToString[nifti.NIFTI_I2S],
	})
	assert.Equal(rd.GetNiiData().GetAffine(), matrix.DMat44{
		M: [4][4]float64{
			{2, 0, 0, -90},
			{0, 2, 0, -126},
			{0, 0, 2, -72},
			{0, 0, 0, 1},
		},
	})
	assert.Equal(rd.GetBinaryOrder(), binary.LittleEndian)
	assert.Equal(rd.GetNiiData().GetDatatype(), "FLOAT32")
	assert.Equal(rd.GetNiiData().GetSFormCode(), "4: MNI")
	assert.Equal(rd.GetNiiData().GetQFormCode(), "0: Unknown")
}

func TestNewNiiReader_Parse_HeaderImagePair(t *testing.T) {
	assert := assert.New(t)

	imgPath := "./test_data/t1.img.gz"
	headerPath := "./test_data/t1.hdr.gz"

	rd, err := NewNiiReader(WithImageFile(imgPath), WithInMemory(true), WithRetainHeader(true), WithHeaderFile(headerPath))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	fmt.Println("datatype", rd.GetNiiData().GetDatatype())
	fmt.Println("image shape", rd.GetNiiData().GetImgShape())
	fmt.Println("affine", rd.GetNiiData().GetAffine())
	fmt.Println("orientation", rd.GetNiiData().GetOrientation())
	fmt.Println("binary order", rd.GetBinaryOrder())
	fmt.Println("slice code", rd.GetNiiData().GetSliceCode())
	fmt.Println("qform_code", rd.GetNiiData().GetQFormCode())
	fmt.Println("sform_code", rd.GetNiiData().GetSFormCode())
	fmt.Println("quatern_b", rd.GetNiiData().GetQuaternB())
	fmt.Println("quatern_c", rd.GetNiiData().GetQuaternC())
	fmt.Println("quatern_d", rd.GetNiiData().GetQuaternD())
}

func Test_MagicString(t *testing.T) {
	fmt.Println("nii1 .hdr/.img pair", []byte("ni1"))
	fmt.Println("nii1 single", []byte("n+1"))
	fmt.Println("nii2 .hdr/.img pair", []byte("ni2"))
	fmt.Println("nii2 single", []byte("n+2"))
}

func TestNewNiiWriter_Voxels(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()

	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("./test_data/int16_voxel_output.nii.gz",
		WithNIfTIData(rd.GetNiiData()),
		WithCompression(true),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}

func TestNewNiiWriter_MakeSegmentation_Single(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()

	for index, voxel := range voxels.GetDataset() {
		if voxel > 0 {
			voxels.GetDataset()[index] = 1
		} else {
			voxels.GetDataset()[index] = 0
		}
	}

	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("./test_data/int16_seg_single.nii.gz",
		WithNIfTIData(rd.GetNiiData()),
		WithCompression(true),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}

func TestNewNiiWriter_MakeSegmentation_Multi(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()

	for index, voxel := range voxels.GetDataset() {
		if voxel > 0 && voxel <= 128 {
			voxels.GetDataset()[index] = 1
		} else if voxel > 128 {
			voxels.GetDataset()[index] = 2
		} else {
			voxels.GetDataset()[index] = 0
		}
	}

	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("./test_data/int16_seg_multi.nii.gz",
		WithNIfTIData(rd.GetNiiData()),
		WithCompression(true),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}

func TestNewNiiWriter_Write_NIfTI2_Single(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()

	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("./test_data/int16_nii2.nii.gz",
		WithNIfTIData(rd.GetNiiData()),
		WithCompression(true),
		WithVersion(2),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}

func TestNewNiiWriter_Write_NIfTI1_Pair(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)
	voxels := rd.GetNiiData().GetVoxels()
	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("./test_data/int16.img",
		WithNIfTIData(rd.GetNiiData()),
		WithCompression(false),
		WithWriteHeaderFile(true),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}

func TestCommon(t *testing.T) {
	x := make([]byte, 10, 10)
	fmt.Println(x)

	y := make([]byte, 10)
	fmt.Println(y)
}

func TestNewNiiWriter_MakeSegmentation_New(t *testing.T) {
	assert := assert.New(t)

	filePath := "/home/tripg/workspace/gonii_test/int16.nii.gz"

	rd, err := NewNiiReader(WithImageFile(filePath), WithRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()

	for index, voxel := range voxels.GetDataset() {
		if voxel > 200 {
			voxels.GetDataset()[index] = 1
		} else {
			voxels.GetDataset()[index] = 0
		}
	}

	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("/home/tripg/workspace/gonii_test/int16_seg_single.nii.gz",
		WithNIfTIData(rd.GetNiiData()),
		WithCompression(true),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}
