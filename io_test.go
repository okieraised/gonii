package gonii

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/okieraised/gonii/pkg/matrix"
	"github.com/okieraised/gonii/pkg/nifti"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewNiiReader_Nii1_Single_Int16(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

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

func TestNewNiiReader_Nii2_Single_LR(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/nii2_LR.nii.gz"

	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(true))
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

func TestNewNiiReader_Nii2_Single_RL(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/nii2_RL.nii.gz"

	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(true))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	rd.GetHeader(false)

	assert.Equal(rd.GetNiiData().GetOrientation(), [3]string{
		nifti.OrietationToString[nifti.NIFTI_L2R],
		nifti.OrietationToString[nifti.NIFTI_P2A],
		nifti.OrietationToString[nifti.NIFTI_I2S],
	})

	assert.Equal(rd.GetBinaryOrder(), binary.LittleEndian)

	assert.Equal(rd.GetNiiData().GetAffine(), matrix.DMat44{
		M: [4][4]float64{
			{2, 0, 0, -90},
			{0, 2, 0, -126},
			{0, 0, 2, -72},
			{0, 0, 0, 1},
		},
	})
	assert.Equal(rd.GetNiiData().GetDatatype(), "FLOAT32")
	assert.Equal(rd.GetNiiData().GetSFormCode(), "4: MNI")
	assert.Equal(rd.GetNiiData().GetQFormCode(), "0: Unknown")
}

func TestNewNiiReader_Nii1_Pair(t *testing.T) {
	assert := assert.New(t)

	imgPath := "./test_data/t1.img.gz"
	headerPath := "./test_data/t1.hdr.gz"

	rd, err := NewNiiReader(
		WithReadImageFile(imgPath),
		WithReadInMemory(true),
		WithReadRetainHeader(true),
		WithReadHeaderFile(headerPath),
	)
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

	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()

	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("./test_data/int16_voxel_output.nii.gz",
		WithWriteNIfTIData(rd.GetNiiData()),
		WithWriteCompression(true),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}

func TestNewNiiWriter_Nii2_Single(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()

	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("./test_data/int16_nii2.nii.gz",
		WithWriteNIfTIData(rd.GetNiiData()),
		WithWriteCompression(true),
		WithWriteVersion(2),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}

func TestNewNiiWriter_Nii1_Pair(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)
	voxels := rd.GetNiiData().GetVoxels()
	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("./test_data/int16.img",
		WithWriteNIfTIData(rd.GetNiiData()),
		WithWriteCompression(false),
		WithWriteHeaderFile(true),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}

func TestNewNiiWriter_Segmentation_Single(t *testing.T) {
	assert := assert.New(t)

	filePath := "/home/tripg/workspace/gonii_test/int16.nii.gz"

	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(false))
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
		WithWriteNIfTIData(rd.GetNiiData()),
		WithWriteCompression(true),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}

func TestNewNiiWriter_Segmentation_Multi(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(false))
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
		WithWriteNIfTIData(rd.GetNiiData()),
		WithWriteCompression(true),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}

func TestNewNiiWriter_90MB(t *testing.T) {
	assert := assert.New(t)

	filePath := "/home/tripg/workspace/anim3.nii.gz"

	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(false))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)
}

func TestNewNiiWriter_Nii2_BytesReader(t *testing.T) {
	assert := assert.New(t)

	filePath := "./test_data/int16.nii.gz"

	bContent, err := os.ReadFile(filePath)
	assert.NoError(err)

	rd, err := NewNiiReader(WithReadImageReader(bytes.NewReader(bContent)))
	assert.NoError(err)
	err = rd.Parse()
	assert.NoError(err)

	voxels := rd.GetNiiData().GetVoxels()

	err = rd.GetNiiData().SetVoxelToRawVolume(voxels)
	assert.NoError(err)

	writer, err := NewNiiWriter("./test_data/int16_nii2.nii.gz",
		WithWriteNIfTIData(rd.GetNiiData()),
		WithWriteCompression(true),
		WithWriteVersion(2),
	)
	err = writer.WriteToFile()
	assert.NoError(err)
}
