package gonii

import (
	"testing"

	"github.com/okieraised/gonii/internal/utils"
	"github.com/stretchr/testify/assert"
)

func Test_ReaderProfiling_90MBCompressed(t *testing.T) {
	assert := assert.New(t)

	fn := func() {
		ReadNifti90MBCompressed()
	}
	err := utils.CPUProfilingFunc(fn, "./profiling_90Compressed.pprof")
	assert.NoError(err)
}

func Test_ReaderProfiling_90MB(t *testing.T) {
	assert := assert.New(t)

	fn := func() {
		ReadNifti90MB()
	}
	err := utils.CPUProfilingFunc(fn, "./profiling_90.pprof")
	assert.NoError(err)
}

func BenchmarkNewNiiReader_90MBCompressed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ReadNifti90MBCompressed()
	}
}

func BenchmarkNewNiiReader_90MB(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ReadNifti90MB()
	}
}

func BenchmarkNewNiiReader_2_2MB(b *testing.B) {
	b.N = 100
	for i := 0; i < b.N; i++ {
		ReadNifti2_2MB()
	}
}

func ReadNifti90MBCompressed() {
	filePath := "/home/tripg/workspace/anim3.nii.gz"
	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(false))
	if err != nil {
		return
	}
	err = rd.Parse()
	if err != nil {
		return
	}
}

func ReadNifti90MB() {
	filePath := "/home/tripg/workspace/anim3.nii"
	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(false))
	if err != nil {
		return
	}
	err = rd.Parse()
	if err != nil {
		return
	}
}

func ReadNifti2_2MB() {
	filePath := "/home/tripg/workspace/int16.nii.gz"
	rd, err := NewNiiReader(WithReadImageFile(filePath), WithReadRetainHeader(false))
	if err != nil {
		return
	}
	err = rd.Parse()
	if err != nil {
		return
	}
}
