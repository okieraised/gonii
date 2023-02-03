package nifti

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"github.com/okieraised/gonii/internal/system"
	"github.com/okieraised/gonii/pkg/matrix"
	"math"
	"os"
	"strings"
)

type Writer interface {
	// WriteToFile write the header and image to either a single NIfTI file or a pair of .hdr/.img file
	WriteToFile() error
	// GetNiiData returns the current NIfTI image data
	GetNiiData() *Nii
	// GetVoxels returns the 1-D slices of voxel values as float64 type
	GetVoxels() *Voxels
	// SetAt sets the new value at (x, y, z, t) location
	SetAt(newVal float64, x, y, z, t int64) error
	// SetSliceCode sets the new slice code of the NIFTI image
	SetSliceCode(sliceCode int32) error
	// SetQFormCode sets the new QForm code
	SetQFormCode(qFormCode int32) error
	// SetSFormCode sets the new SForm code
	SetSFormCode(sFormCode int32) error
	// SetDatatype sets the new NIfTI datatype
	SetDatatype(datatype int32) error
	// SetAffine sets the new 4x4 affine matrix
	SetAffine(mat matrix.DMat44)
	// SetDescrip sets the new description
	SetDescrip(descrip string) error
	// SetIntentName sets the new intent name
	SetIntentName(intentName string) error
	// SetSliceDuration sets the new slice duration info
	SetSliceDuration(sliceDuration float64)
	// SetSliceStart sets the new slice start info
	SetSliceStart(sliceStart int64)
	// SetSliceEnd sets the new slice end info
	SetSliceEnd(sliceEnd int64)
	// SetXYZUnits sets the new spatial unit of measurements
	SetXYZUnits(xyzUnit int32)
	// SetTimeUnits sets the new temporal unit of measurements
	SetTimeUnits(timeUnit int32)
	// SetVolume set the NIfTI image bytes to a user-defined byte slice
	SetVolume(vol []byte) error
	// SetVoxelToRawVolume converts the float64 slice of voxel back to its corresponding byte slice
	SetVoxelToRawVolume(vox *Voxels) error
}

// NiiWriter define the NIfTI writer structure.
//
// Parameters:
//     - `filePath`         : Export file path to write NIfTI image
//     - `writeHeaderFile`  : Whether to write NIfTI file pair (hdr + img file)
//     - `compression`      : Whether the NIfTI volume will be compressed. If writeHeaderFile is set to True, both the .hdr and .img files will be compressed
//     - `niiData`          : Input NIfTI data to write to file
//     - `header`           : Input NIfTI header to write to file. If nil, the default header will be constructed
type NiiWriter struct {
	filePath        string      // Export file path to write NIfTI image
	writeHeaderFile bool        // Whether to write NIfTI file pair (hdr + img file)
	compression     bool        // Whether the NIfTI file will be compressed
	niiData         *Nii        // Input NIfTI data to write to file
	header          *Nii1Header // Input NIfTI header to write to file. If nil, the default header will be constructed
}

func (w *NiiWriter) SetFilePath(filePath string) {
	w.filePath = filePath
}

func (w *NiiWriter) SetWriteHeaderFile(writeHeaderFile bool) {
	w.writeHeaderFile = writeHeaderFile
}

func (w *NiiWriter) SetCompression(compression bool) {
	w.compression = compression
}

func (w *NiiWriter) SetNiiData(nii *Nii) {
	w.niiData = nii
}

func (w *NiiWriter) SetHeader(hdr *Nii1Header) {
	w.header = hdr
}

// NewNiiWriter returns a new write for export
func NewNiiWriter(filePath string, options ...func(*NiiWriter)) (*NiiWriter, error) {
	writer := new(NiiWriter)

	writer.filePath = filePath
	writer.writeHeaderFile = false // Default to false. Write to a single file only
	writer.compression = false     // Default to false. No compression

	// Other options
	for _, opt := range options {
		opt(writer)
	}
	return writer, nil
}

// WithWriteHeaderFile sets the option to write NIfTI image to a header/image (.hdr/.img) file pair
//
// If true, output will be two files for the header and the image. Default is false.
func WithWriteHeaderFile(writeHeaderFile bool) func(*NiiWriter) {
	return func(w *NiiWriter) {
		w.writeHeaderFile = writeHeaderFile
	}
}

// WithCompression sets the option to write compressed NIfTI image to a single file (.nii.gz)
//
// If true, the whole file will be compressed. Default is false.
func WithCompression(withCompression bool) func(writer *NiiWriter) {
	return func(w *NiiWriter) {
		w.compression = withCompression
	}
}

// WithHeader sets the option to allow user to provide predefined NIfTI-1 header structure.
//
// If no header provided, the header will be converted from the NIfTI image structure
func WithHeader(header *Nii1Header) func(*NiiWriter) {
	return func(w *NiiWriter) {
		w.header = header
	}
}

// WithNIfTIData sets the option to allow user to provide predefined NIfTI-1 data structure.
func WithNIfTIData(data *Nii) func(writer *NiiWriter) {
	return func(w *NiiWriter) {
		w.niiData = data
	}
}

// WriteToFile write the header and image to either a single NIfTI file or a pair of .hdr/.img file
func (w *NiiWriter) WriteToFile() error {
	// convert image structure to file
	err := w.convertImageToHeader()
	if err != nil {
		return err
	}

	// If user decides to write to a separate hdr/img file pair
	if w.writeHeaderFile {
		err := w.writePairNii()
		if err != nil {
			return err
		}
	} else { // Just one file for both header and the image data
		err := w.writeSingleNii()
		if err != nil {
			return err
		}
	}
	return nil
}

// writePairNii writes the header and NIfTI image Nii as 2 separate files
func (w *NiiWriter) writePairNii() error {
	var headerFilePath string

	// Check if the user-specified filePath suffix is ending with '.nii'.
	// If not, we append '.nii' to the end to signify the file is NIfTI format
	if !strings.HasSuffix(w.filePath, NIFTI_EXT) {
		w.filePath = w.filePath + NIFTI_EXT
	}

	headerFilePath = w.filePath
	// Now replace the suffix to identify the header and img file
	headerFilePath = strings.ReplaceAll(w.filePath, NIFTI_EXT, "_nifti.hdr")
	w.filePath = strings.ReplaceAll(w.filePath, NIFTI_EXT, "_nifti.img")

	// Check if the user-specified filePath suffix is ending with '.gz'.
	// If not, we append '.gz' to the end to signify the file is compressed
	if w.compression {
		if !strings.HasSuffix(w.filePath, NIFTI_COMPRESSED_EXT) {
			w.filePath = w.filePath + NIFTI_COMPRESSED_EXT
			headerFilePath = headerFilePath + NIFTI_COMPRESSED_EXT
		}
	}

	// Set the magic string to ni1
	w.header.Magic = [4]uint8{110, 105, 49, 0}
	// Set the VoxOffset to 0 since we write to separate header/img file
	w.header.VoxOffset = 0

	// Write header structure as bytes
	hdrBuf := &bytes.Buffer{}
	err := binary.Write(hdrBuf, system.NativeEndian, w.header)
	if err != nil {
		return err
	}
	bHeader := hdrBuf.Bytes()

	// Image data
	bData := w.niiData.Volume

	// Create header file object
	fHeader, err := os.Create(headerFilePath)
	if err != nil {
		return err
	}
	defer fHeader.Close()

	// Create data file object
	fData, err := os.Create(w.filePath)
	if err != nil {
		return err
	}
	defer fData.Close()

	// If compression option is set to true, write both the header and image data as compressed files
	if w.compression {
		// Write compressed header to file
		gzipWriter := gzip.NewWriter(fHeader)
		_, err = gzipWriter.Write(bHeader)
		if err != nil {
			return err
		}
		err = gzipWriter.Close()

		// Write compressed data to file
		gzipWriter = gzip.NewWriter(fData)
		_, err = gzipWriter.Write(bData)
		if err != nil {
			return err
		}
		err = gzipWriter.Close()

	} else { // Write both the header and image data normally
		_, err = fHeader.Write(bHeader)
		if err != nil {
			return err
		}

		_, err = fData.Write(bData)
		if err != nil {
			return err
		}
	}

	return nil
}

// writeSingleNii writes the header and NIfTI image Nii to a single NIfTI file
func (w *NiiWriter) writeSingleNii() error {
	// Set the magic string to n+1
	w.header.Magic = [4]uint8{110, 43, 49, 0}

	// Need to get the number of bytes between the end of header structure and the start of the image data
	offsetFromHeaderToVoxel := int(w.header.VoxOffset) - int(w.header.SizeofHdr)
	var offset []byte

	// If the header size is equals to 348 (default), then we need to add 4 bytes or the offset we calculated to make it divisible by 6
	if int(w.header.SizeofHdr) == NII1HeaderSize {
		defaultPadding := 4
		if offsetFromHeaderToVoxel > 0 {
			offset = make([]byte, offsetFromHeaderToVoxel, offsetFromHeaderToVoxel)
		} else {
			// This is for a case where we read the image as .hdr/.img pair but then want to write to a single file.
			// We have to update the VoxOffset value
			w.header.VoxOffset = float32(int(w.header.SizeofHdr) + defaultPadding)
			offset = make([]byte, defaultPadding, defaultPadding)
		}
	}

	// Make a buffer and write the header to it with default system endian
	hdrBuf := &bytes.Buffer{}
	err := binary.Write(hdrBuf, system.NativeEndian, w.header)
	if err != nil {
		return err
	}

	bHeader := hdrBuf.Bytes()
	bData := w.niiData.Volume

	var dataset []byte
	dataset = append(dataset, bHeader...)
	dataset = append(dataset, offset...)
	dataset = append(dataset, bData...)

	// Check if the user-specified filePath suffix is ending with '.nii'.
	// If not, we append '.nii' to the end to signify the file is NIfTI format
	if !strings.HasSuffix(w.filePath, NIFTI_EXT) {
		w.filePath = w.filePath + NIFTI_EXT
	}

	// Check if the user-specified filePath suffix is ending with '.gz'.
	// If not, we append '.gz' to the end to signify the file is compressed
	if w.compression {
		if !strings.HasSuffix(w.filePath, NIFTI_COMPRESSED_EXT) {
			w.filePath = w.filePath + NIFTI_COMPRESSED_EXT
		}
	}

	// Create a file object from the specified filePath
	file, err := os.Create(w.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if w.compression { // If the compression is set to true, then write a compressed file
		gzipWriter := gzip.NewWriter(file)
		_, err = gzipWriter.Write(dataset)
		if err != nil {
			return err
		}
		err = gzipWriter.Close()
		if err != nil {
			return err
		}
	} else { // Otherwise, just write normal file
		_, err = file.Write(dataset)
		if err != nil {
			return err
		}
	}
	return nil
}

// convertImageToHeader returns the header from a NIfTI image structure
func (w *NiiWriter) convertImageToHeader() error {
	if w.niiData == nil {
		return errors.New("image data structure is nil")
	}

	header := new(Nii1Header)
	header.SizeofHdr = NII1HeaderSize
	header.Regular = 'r'

	header.Dim[0] = int16(w.niiData.NDim)
	header.Dim[1], header.Dim[2], header.Dim[3] = int16(w.niiData.Nx), int16(w.niiData.Ny), int16(w.niiData.Nz)
	header.Dim[4], header.Dim[5], header.Dim[6] = int16(w.niiData.Nt), int16(w.niiData.Nu), int16(w.niiData.Nv)
	header.Dim[7] = int16(w.niiData.Nw)

	header.Pixdim[0] = 0.0
	header.Pixdim[1], header.Pixdim[2], header.Pixdim[3] = float32(math.Abs(w.niiData.Dx)), float32(math.Abs(w.niiData.Dy)), float32(math.Abs(w.niiData.Dz))
	header.Pixdim[4], header.Pixdim[5], header.Pixdim[6] = float32(math.Abs(w.niiData.Dt)), float32(math.Abs(w.niiData.Du)), float32(math.Abs(w.niiData.Dv))
	header.Pixdim[7] = float32(w.niiData.Dw)

	header.Datatype = int16(w.niiData.Datatype)
	header.Bitpix = int16(8 * w.niiData.NByPer)

	if w.niiData.CalMax > w.niiData.CalMin {
		header.CalMin = float32(w.niiData.CalMin)
		header.CalMax = float32(w.niiData.CalMax)
	}

	if w.niiData.SclSlope != 0.0 {
		header.SclSlope = float32(w.niiData.SclSlope)
		header.SclInter = float32(w.niiData.SclInter)
	}

	if w.niiData.Descrip[0] != 0x0 {
		for i := 0; i < 79; i++ {
			header.Descrip[i] = w.niiData.Descrip[i]
		}
		header.Descrip[79] = 0x0
	}

	if w.niiData.AuxFile[0] != 0x0 {
		for i := 0; i < 23; i++ {
			header.AuxFile[i] = w.niiData.AuxFile[i]
		}
		header.AuxFile[23] = 0x0
	}

	// Load NIFTI specific stuff into the header
	if w.writeHeaderFile {
		header.Magic = [4]byte{110, 43, 49, 0} // n+1
	} else {
		header.Magic = [4]byte{110, 105, 49, 0} // ni1
	}

	header.IntentCode = int16(w.niiData.IntentCode)
	header.IntentP1 = float32(w.niiData.IntentP1)
	header.IntentP2 = float32(w.niiData.IntentP2)
	header.IntentP3 = float32(w.niiData.IntentP3)
	if w.niiData.IntentName[0] != 0x0 {
		for i := 0; i < 15; i++ {
			header.IntentName[i] = w.niiData.IntentName[i]
		}
		header.IntentName[15] = 0x0
	}

	header.VoxOffset = float32(w.niiData.VoxOffset)
	header.XyztUnits = convertSpaceTimeToXYZT(w.niiData.XYZUnits, w.niiData.TimeUnits)
	header.Toffset = float32(w.niiData.TOffset)

	if w.niiData.QformCode > 0 {
		header.QformCode = int16(w.niiData.QformCode)
		header.QuaternB = float32(w.niiData.QuaternB)
		header.QuaternC = float32(w.niiData.QuaternC)
		header.QuaternD = float32(w.niiData.QuaternD)

		header.QoffsetX = float32(w.niiData.QoffsetX)
		header.QoffsetY = float32(w.niiData.QoffsetY)
		header.QoffsetZ = float32(w.niiData.QoffsetZ)

		if w.niiData.QFac >= 0 {
			header.Pixdim[0] = 1.0
		} else {
			header.Pixdim[0] = -1.0
		}
	}

	if w.niiData.SformCode > 0 {
		header.SformCode = int16(w.niiData.SformCode)
		header.SrowX[0] = float32(w.niiData.StoXYZ.M[0][0])
		header.SrowX[1] = float32(w.niiData.StoXYZ.M[0][1])
		header.SrowX[2] = float32(w.niiData.StoXYZ.M[0][2])
		header.SrowX[3] = float32(w.niiData.StoXYZ.M[0][3])

		header.SrowY[0] = float32(w.niiData.StoXYZ.M[1][0])
		header.SrowY[1] = float32(w.niiData.StoXYZ.M[1][1])
		header.SrowY[2] = float32(w.niiData.StoXYZ.M[1][2])
		header.SrowY[3] = float32(w.niiData.StoXYZ.M[1][3])

		header.SrowZ[0] = float32(w.niiData.StoXYZ.M[2][0])
		header.SrowZ[1] = float32(w.niiData.StoXYZ.M[2][1])
		header.SrowZ[2] = float32(w.niiData.StoXYZ.M[2][2])
		header.SrowZ[3] = float32(w.niiData.StoXYZ.M[2][3])
	}

	header.DimInfo = convertFPSIntoDimInfo(w.niiData.FreqDim, w.niiData.PhaseDim, w.niiData.SliceDim)

	header.SliceCode = uint8(w.niiData.SliceCode)
	header.SliceStart = int16(w.niiData.SliceStart)
	header.SliceEnd = int16(w.niiData.SliceEnd)
	header.SliceDuration = float32(w.niiData.SliceDuration)

	w.header = header

	return nil
}

// GetNiiData returns the current NIfTI image data
func (w *NiiWriter) GetNiiData() *Nii {
	return w.niiData
}

// SetAt sets the new value at (x, y, z, t) location
func (w *NiiWriter) SetAt(newVal float64, x, y, z, t int64) error {
	return w.niiData.setAt(newVal, x, y, z, t)
}

// SetSliceCode sets the new slice code of the NIFTI image
func (w *NiiWriter) SetSliceCode(sliceCode int32) error {
	return w.niiData.setSliceCode(sliceCode)
}

// SetQFormCode sets the new QForm code
func (w *NiiWriter) SetQFormCode(qFormCode int32) error {
	return w.niiData.setQFormCode(qFormCode)
}

// SetSFormCode sets the new SForm code
func (w *NiiWriter) SetSFormCode(sFormCode int32) error {
	return w.niiData.setSFormCode(sFormCode)
}

// SetDatatype sets the new NIfTI datatype
func (w *NiiWriter) SetDatatype(datatype int32) error {
	return w.niiData.setDatatype(datatype)
}

// SetAffine sets the new 4x4 affine matrix
func (w *NiiWriter) SetAffine(mat matrix.DMat44) {
	w.niiData.setAffine(mat)
}

// SetDescrip returns the description with trailing null bytes removed
func (w *NiiWriter) SetDescrip(descrip string) error {
	return w.niiData.setDescrip(descrip)
}

// SetIntentName sets the new intent name
func (w *NiiWriter) SetIntentName(intentName string) error {
	return w.niiData.setIntentName(intentName)
}

// SetSliceDuration sets the new slice duration info
func (w *NiiWriter) SetSliceDuration(sliceDuration float64) {
	w.niiData.setSliceDuration(sliceDuration)
}

// SetSliceStart sets the new slice start info
func (w *NiiWriter) SetSliceStart(sliceStart int64) {
	w.niiData.setSliceStart(sliceStart)
}

// SetSliceEnd sets the new slice end info
func (w *NiiWriter) SetSliceEnd(sliceEnd int64) {
	w.niiData.setSliceEnd(sliceEnd)
}

// SetXYZUnits sets the new spatial unit of measurements
func (w *NiiWriter) SetXYZUnits(xyzUnit int32) {
	w.niiData.setXYZUnits(xyzUnit)
}

// SetTimeUnits sets the new temporal unit of measurements
func (w *NiiWriter) SetTimeUnits(timeUnit int32) {
	w.niiData.setTimeUnits(timeUnit)
}

func (w *NiiWriter) SetVolume(vol []byte) error {
	return w.niiData.setVolume(vol)
}

// GetVoxels returns the 1-D slices of voxel value as float64 type
func (w *NiiWriter) GetVoxels() *Voxels {
	return w.niiData.getVoxel()
}

// SetVoxelToRawVolume converts the float64 slice of voxel back to its corresponding byte slice
func (w *NiiWriter) SetVoxelToRawVolume(vox *Voxels) error {
	return w.niiData.setVoxelToRawVolume(vox)
}
