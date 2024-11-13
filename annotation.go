package gonii

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"

	"github.com/okieraised/gonii/internal/system"
	"github.com/okieraised/gonii/pkg/nifti"
)

type Segmentation struct {
	nii1Hdr       *nifti.Nii1Header
	nii2Hdr       *nifti.Nii2Header
	img           *nifti.Nii
	outFile       string
	compression   bool
	annotations   []SegmentCoordinate
	annotationRLE []nifti.SegmentRLE
}

// SegmentCoordinate defines the structure for segmentation coordinate
type SegmentCoordinate struct {
	Value any   `json:"value"`
	X     int64 `json:"x"`
	Y     int64 `json:"y"`
	Z     int64 `json:"z"`
	T     int64 `json:"t"`
}

func NewSegmentation(opts ...SegmentationOption) *Segmentation {
	seg := &Segmentation{
		compression: true,
	}
	for _, opt := range opts {
		opt(seg)
	}
	return seg
}

type SegmentationOption func(s *Segmentation)

// WithSegCompression allows user to write the segmentation as compressed NIfTI file
func WithSegCompression(compression bool) SegmentationOption {
	return func(s *Segmentation) {
		s.compression = compression
	}
}

// WithOutFile allows user to specify the location to write the segmentation data to file
func WithOutFile(outFile string) SegmentationOption {
	return func(s *Segmentation) {
		s.outFile = outFile
	}
}

// WithNii1Hdr allows users to specify NIfTI-1 header to write the annotation file to
//
// If both NIfTI-1 and NIfTI-2 header are specified. NIfTI-2 header takes precedence
func WithNii1Hdr(hdr *nifti.Nii1Header) SegmentationOption {
	return func(s *Segmentation) {
		s.nii1Hdr = hdr
	}
}

// WithNii2Hdr allows users to specify NIfTI-1 header to write the annotation file to
//
// If both NIfTI-1 and NIfTI-2 header are specified. NIfTI-2 header takes precedence
func WithNii2Hdr(hdr *nifti.Nii2Header) SegmentationOption {
	return func(s *Segmentation) {
		s.nii2Hdr = hdr
	}
}

// WithAnnotations allows users to specify the annotation coordinates to convert to NIfTI file
func WithAnnotations(annotations []SegmentCoordinate) SegmentationOption {
	return func(s *Segmentation) {
		s.annotations = annotations
	}
}

// WithSegmentRLE allows users to specify the annotation as RLE-encoded array to convert to NIfTI file
func WithSegmentRLE(segments []nifti.SegmentRLE) SegmentationOption {
	return func(s *Segmentation) {
		s.annotationRLE = segments
	}
}

// WithImage allows users to specify the NIfTI image structure to convert to (x,y,z,t) coordinate array
func WithImage(image *nifti.Nii) SegmentationOption {
	return func(s *Segmentation) {
		s.img = image
	}
}

// AnnotationNiiToJson converts the NIfTI file to a corresponding annotation coordinates (x,y,z,t) array
func (s *Segmentation) AnnotationNiiToJson() error {
	voxels := s.img.GetVoxels()

	var res []SegmentCoordinate

	for x := int64(0); x < s.img.Nx; x++ {
		for y := int64(0); y < s.img.Ny; y++ {
			for z := int64(0); z < s.img.Nz; z++ {
				for t := int64(0); t < s.img.Nt; t++ {
					val := voxels.Get(x, y, z, t)
					if val != 0 {
						coord := SegmentCoordinate{
							X:     x,
							Y:     y,
							Z:     z,
							T:     t,
							Value: int64(val),
						}
						res = append(res, coord)
					}
				}
			}
		}
	}

	if s.outFile != "" {
		file, err := os.Create(s.outFile)
		if err != nil {
			return err
		}
		defer file.Close()
		dataset, err := json.MarshalIndent(res, "", "\t")
		_, err = file.Write(dataset)
		if err != nil {
			return err
		}
	}

	return nil
}

// AnnotationJsonToNii converts the annotation coordinates (x,y,z,t) array to a corresponding NIfTI file
func (s *Segmentation) AnnotationJsonToNii() error {
	if s.annotations == nil {
		return errors.New("no input annotations is specified")
	}
	if s.nii1Hdr == nil && s.nii2Hdr == nil {
		return errors.New("at least one header structure must be specified")
	}

	if (s.nii1Hdr != nil && s.nii2Hdr != nil) || (s.nii1Hdr == nil && s.nii2Hdr != nil) {
		err := s.convertSegmentationToNii2()
		if err != nil {
			return err
		}
	} else {
		err := s.convertSegmentationToNii1()
		if err != nil {
			return err
		}
	}
	return nil
}

// convertSegmentationToNii1 converts the voxel and the header to a NIfTI-1 file
func (s *Segmentation) convertSegmentationToNii1() error {

	// Check and fix bad value in header
	for index, _ := range s.nii1Hdr.Dim {
		if index > 0 {
			if s.nii1Hdr.Dim[index] <= 0 {
				s.nii1Hdr.Dim[index] = 1
			}
		}
	}

	// If bitpix is zero then just return error
	if s.nii1Hdr.Bitpix <= 0 {
		return errors.New("bitpix value must be larger than 0")
	}

	nx, ny, nz, nt := int64(s.nii1Hdr.Dim[1]), int64(s.nii1Hdr.Dim[2]), int64(s.nii1Hdr.Dim[3]), int64(s.nii1Hdr.Dim[4])
	datatype := int32(s.nii1Hdr.Datatype)

	vox := nifti.NewVoxels(nx, ny, nz, nt, datatype)
	valMapper := map[any]float64{}

	// NIfTI can have multiple annotations on the same file,
	// so we have to assign the same pixel value for coordinates with same value
	for _, coord := range s.annotations {
		var byteCode float64 = 1
		_, ok := valMapper[coord.Value]
		if !ok {
			valMapper[coord.Value] = byteCode
			byteCode++
		} else {
			vox.Set(coord.X, coord.Y, coord.Z, coord.T, valMapper[coord.Value])
		}
	}

	// Create a zero-filled voxel slice then we convert the voxel at each index from float64 to []byte
	nByPer, _ := nifti.AssignDatatypeSize(datatype)
	rawImg := make([]byte, vox.GetRawByteSize(), vox.GetRawByteSize())
	for index, voxel := range vox.GetDataset() {
		bVal, err := nifti.ConvertVoxelToBytes(
			voxel,
			float64(s.nii1Hdr.SclSlope),
			float64(s.nii1Hdr.SclInter),
			int32(s.nii1Hdr.Datatype),
			system.NativeEndian,
			int32(nByPer),
		)
		if err != nil {
			return err
		}
		copy(rawImg[index*int(nByPer):(index+1)*int(nByPer)], bVal)
	}

	// Export segmentation to file
	hdrBuf := &bytes.Buffer{}
	err := binary.Write(hdrBuf, system.NativeEndian, s.nii1Hdr)
	if err != nil {
		return err
	}

	if s.outFile != "" {
		wr, err := NewNiiWriter(s.outFile,
			WithWriteCompression(s.compression),
			WithWriteVersion(nifti.NIIVersion1),
			WithWriteNii1Header(s.nii1Hdr),
			WithWriteNIfTIData(&nifti.Nii{Volume: rawImg}),
		)
		if err != nil {
			return err
		}
		err = wr.WriteToFile()
		if err != nil {
			return err
		}
	}

	return nil
}

// convertSegmentationToNii2 converts the voxel and the header to a NIfTI-2 file
func (s *Segmentation) convertSegmentationToNii2() error {
	// Check and fix bad value in header
	for index, _ := range s.nii2Hdr.Dim {
		if index > 0 {
			if s.nii2Hdr.Dim[index] <= 0 {
				s.nii2Hdr.Dim[index] = 1
			}
		}
	}

	// If bitpix is zero then just return error
	if s.nii2Hdr.Bitpix <= 0 {
		return errors.New("bitpix value must be larger than 0")
	}

	nx, ny, nz, nt := s.nii2Hdr.Dim[1], s.nii2Hdr.Dim[2], s.nii2Hdr.Dim[3], s.nii2Hdr.Dim[4]
	datatype := int32(s.nii1Hdr.Datatype)

	vox := nifti.NewVoxels(nx, ny, nz, nt, datatype)
	valMapper := map[any]float64{}

	// NIfTI can have multiple annotations on the same file,
	// so we have to assign the same pixel value for coordinates with same value
	for _, coord := range s.annotations {
		var byteCode float64 = 1
		_, ok := valMapper[coord.Value]
		if !ok {
			valMapper[coord.Value] = byteCode
			byteCode++
		} else {
			vox.Set(coord.X, coord.Y, coord.Z, coord.T, valMapper[coord.Value])
		}
	}

	// Create a zero-filled voxel slice then we convert the voxel at each index from float64 to []byte
	nByPer, _ := nifti.AssignDatatypeSize(datatype)
	rawImg := make([]byte, vox.GetRawByteSize(), vox.GetRawByteSize())
	for index, voxel := range vox.GetDataset() {
		bVal, err := nifti.ConvertVoxelToBytes(
			voxel,
			s.nii2Hdr.SclSlope,
			s.nii2Hdr.SclInter,
			int32(s.nii2Hdr.Datatype),
			system.NativeEndian,
			int32(nByPer),
		)
		if err != nil {
			return err
		}
		copy(rawImg[index*int(nByPer):(index+1)*int(nByPer)], bVal)
	}

	// Export segmentation to file
	hdrBuf := &bytes.Buffer{}
	err := binary.Write(hdrBuf, system.NativeEndian, s.nii2Hdr)
	if err != nil {
		return err
	}

	if s.outFile != "" {
		wr, err := NewNiiWriter(s.outFile,
			WithWriteCompression(s.compression),
			WithWriteVersion(nifti.NIIVersion2),
			WithWriteNii2Header(s.nii2Hdr),
			WithWriteNIfTIData(&nifti.Nii{Volume: rawImg}),
		)
		if err != nil {
			return err
		}
		err = wr.WriteToFile()
		if err != nil {
			return err
		}
	}

	return nil
}
