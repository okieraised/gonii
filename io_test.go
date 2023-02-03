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

	rd, err := NewNiiReader(filePath, WithInMemory(true), WithRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	assert.Equal(rd.GetDatatype(), "INT16")

	rd.GetHeader(false)

	fmt.Println("datatype", rd.GetDatatype())
	fmt.Println("image shape", rd.GetImgShape())
	fmt.Println("affine", rd.GetAffine())
	fmt.Println("orientation", rd.GetOrientation())
	fmt.Println("binary order", rd.GetBinaryOrder())
	fmt.Println("slice code", rd.GetSliceCode())
	fmt.Println("qform_code", rd.GetQFormCode())
	fmt.Println("sform_code", rd.GetSFormCode())
	fmt.Println("quatern_b", rd.GetQuaternB())
	fmt.Println("quatern_c", rd.GetQuaternC())
	fmt.Println("quatern_d", rd.GetQuaternD())
}

func TestNiiReader_Parse_SingleFile_Nii2_LR(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/nii2_LR.nii.gz"

	rd, err := NewNiiReader(filePath, WithRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	rd.GetHeader(false)

	assert.Equal(rd.GetOrientation(), [3]string{
		nifti.OrietationToString[nifti.NIFTI_R2L],
		nifti.OrietationToString[nifti.NIFTI_P2A],
		nifti.OrietationToString[nifti.NIFTI_I2S],
	})

	assert.Equal(rd.GetBinaryOrder(), binary.LittleEndian)

	assert.Equal(rd.GetAffine(), matrix.DMat44{
		M: [4][4]float64{
			{-2, 0, 0, 90},
			{0, 2, 0, -126},
			{0, 0, 2, -72},
			{0, 0, 0, 1},
		},
	})

	fmt.Println("datatype", rd.GetDatatype())
	fmt.Println("image shape", rd.GetImgShape())
	fmt.Println("affine", rd.GetAffine())
	fmt.Println("orientation", rd.GetOrientation())
	fmt.Println("binary order", rd.GetBinaryOrder())
	fmt.Println("slice code", rd.GetSliceCode())
	fmt.Println("qform_code", rd.GetQFormCode())
	fmt.Println("sform_code", rd.GetSFormCode())
	fmt.Println("quatern_b", rd.GetQuaternB())
	fmt.Println("quatern_c", rd.GetQuaternC())
	fmt.Println("quatern_d", rd.GetQuaternD())
	fmt.Println(rd.GetUnitsOfMeasurements())
}

func TestNiiReader_Parse_SingleFile_Nii2_RL(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/nii2_RL.nii.gz"

	rd, err := NewNiiReader(filePath, WithRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	rd.GetHeader(false)

	assert.Equal(rd.GetOrientation(), [3]string{
		nifti.OrietationToString[nifti.NIFTI_L2R],
		nifti.OrietationToString[nifti.NIFTI_P2A],
		nifti.OrietationToString[nifti.NIFTI_I2S],
	})

	assert.Equal(rd.GetBinaryOrder(), binary.LittleEndian)

	fmt.Println("datatype", rd.GetDatatype())
	fmt.Println("image shape", rd.GetImgShape())
	fmt.Println("affine", rd.GetAffine())
	fmt.Println("orientation", rd.GetOrientation())
	fmt.Println("binary order", rd.GetBinaryOrder())
	fmt.Println("slice code", rd.GetSliceCode())
	fmt.Println("qform_code", rd.GetQFormCode())
	fmt.Println("sform_code", rd.GetSFormCode())
	fmt.Println("quatern_b", rd.GetQuaternB())
	fmt.Println("quatern_c", rd.GetQuaternC())
	fmt.Println("quatern_d", rd.GetQuaternD())
	fmt.Println(rd.GetUnitsOfMeasurements())
}

func TestNewNiiReader_Parse_HeaderImagePair(t *testing.T) {
	assert := assert.New(t)

	imgPath := "./test_data/t1.img.gz"
	headerPath := "./test_data/t1.hdr.gz"

	rd, err := NewNiiReader(imgPath, WithInMemory(true), WithRetainHeader(true), WithHeaderFile(headerPath))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	fmt.Println("datatype", rd.GetDatatype())
	fmt.Println("image shape", rd.GetImgShape())
	fmt.Println("affine", rd.GetAffine())
	fmt.Println("orientation", rd.GetOrientation())
	fmt.Println("binary order", rd.GetBinaryOrder())
	fmt.Println("slice code", rd.GetSliceCode())
	fmt.Println("qform_code", rd.GetQFormCode())
	fmt.Println("sform_code", rd.GetSFormCode())
	fmt.Println("quatern_b", rd.GetQuaternB())
	fmt.Println("quatern_c", rd.GetQuaternC())
	fmt.Println("quatern_d", rd.GetQuaternD())
}
