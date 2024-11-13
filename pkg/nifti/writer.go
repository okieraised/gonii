package nifti

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	gzip "github.com/klauspost/pgzip"
	"github.com/okieraised/gonii/internal/system"
)

type Writer interface {
	// WriteToFile write the header and image to either a single NIfTI file or a pair of .hdr/.img file
	WriteToFile() error
	// GetNiiData returns the current NIfTI image data
	GetNiiData() *Nii
	// GetHeader returns the current NIfTI header
	GetHeader() interface{}
	// WriteToBytes the NIfTI dataset as byte slice
	WriteToBytes() ([]byte, error)
}

// NiiWriter define the NIfTI writer structure.
//
// Parameters:
//   - `filePath`         : Export file path to write NIfTI image
//   - `writeHeaderFile`  : Whether to write NIfTI file pair (hdr + img file)
//   - `compression`      : Whether the NIfTI volume will be compressed. If writeHeaderFile is set to True, both the .hdr and .img files will be compressed
//   - `niiData`          : Input NIfTI data to write to file
//   - `header`           : Input NIfTI header to write to file. If nil, the default header will be constructed
//   - `version`          : Specify the version (NIfTI-1 or NIfTI-2) to export
type NiiWriter struct {
	filePath        string      // Export file path to write NIfTI image
	writeHeaderFile bool        // Whether to write NIfTI file pair (hdr + img file)
	compression     bool        // Whether the NIfTI file will be compressed
	niiData         *Nii        // Input NIfTI data to write to file
	header          interface{} // Input NIfTI header to write to file. If nil, the default header will be constructed
	version         int         //Specify the version (NIfTI-1 or NIfTI-2) to export
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

func (w *NiiWriter) SetHeader(hdr interface{}) {
	w.header = hdr
}

func (w *NiiWriter) SetVersion(version int) {
	w.version = version
}

func (w *NiiWriter) WriteToBytes() ([]byte, error) {
	// Convert image to header
	switch w.version {
	case NIIVersion1:
		err := w.convertImageToNii1Header()
		if err != nil {
			return nil, err
		}
	case NIIVersion2:
		err := w.convertImageToNii2Header()
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown NIfTI version %d", w.version)
	}

	return w.reconstructDataset()
}

// WriteToFile write the header and image to either a single NIfTI file or a pair of .hdr/.img file
func (w *NiiWriter) WriteToFile() error {
	// Convert image to header
	switch w.version {
	case NIIVersion1:
		err := w.convertImageToNii1Header()
		if err != nil {
			return err
		}
	case NIIVersion2:
		err := w.convertImageToNii2Header()
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown NIfTI version %d", w.version)
	}

	// convert image structure to file
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

func (w *NiiWriter) reconstructDataset() ([]byte, error) {
	var offset []byte
	var offsetFromHeaderToVoxel int

	// Need to get the number of bytes between the end of header structure and the start of the image data
	switch hdr := w.header.(type) {
	case *Nii1Header:
		offsetFromHeaderToVoxel = int(hdr.VoxOffset) - int(hdr.SizeofHdr)
	case *Nii2Header:
		offsetFromHeaderToVoxel = int(hdr.VoxOffset) - int(hdr.SizeofHdr)
	default:
		return nil, fmt.Errorf("unknown header type")
	}

	if offsetFromHeaderToVoxel > 0 {
		offset = make([]byte, offsetFromHeaderToVoxel)
	} else {
		offset = make([]byte, DefaultHeaderPadding)
	}

	// Make a buffer and write the header to it with default system endian
	hdrBuf := &bytes.Buffer{}
	err := binary.Write(hdrBuf, system.NativeEndian, w.header)
	if err != nil {
		return nil, err
	}

	bHeader := hdrBuf.Bytes()
	bData := w.niiData.Volume

	var dataset []byte
	dataset = append(dataset, bHeader...)
	dataset = append(dataset, offset...)
	dataset = append(dataset, bData...)

	return dataset, nil
}

// writePairNii writes the header and NIfTI image Nii as 2 separate files
func (w *NiiWriter) writePairNii() error {
	var headerFilePath string

	// Check if the user-specified filePath suffix is ending with '.nii' or '.gz'.
	// If not, we append '.nii' to the end to signify the file is NIfTI format
	if !strings.HasSuffix(w.filePath, NIFTI_EXT) && !strings.HasSuffix(w.filePath, NIFTI_COMPRESSED_EXT) {
		// If user specifies the file with '.gz' extension then default to compression
		if strings.HasSuffix(w.filePath, NIFTI_COMPRESSED_EXT) {
			w.compression = true
		}
		w.filePath = w.filePath + NIFTI_EXT
	}

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
		if err != nil {
			return err
		}

		// Write compressed data to file
		gzipWriter = gzip.NewWriter(fData)
		_, err = gzipWriter.Write(bData)
		if err != nil {
			return err
		}
		err = gzipWriter.Close()
		if err != nil {
			return err
		}

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
	//var offset []byte
	//var offsetFromHeaderToVoxel int
	//
	//// Need to get the number of bytes between the end of header structure and the start of the image data
	//switch hdr := w.header.(type) {
	//case *Nii1Header:
	//	offsetFromHeaderToVoxel = int(hdr.VoxOffset) - int(hdr.SizeofHdr)
	//case *Nii2Header:
	//	offsetFromHeaderToVoxel = int(hdr.VoxOffset) - int(hdr.SizeofHdr)
	//default:
	//	return fmt.Errorf("unknown header type")
	//}
	//
	//if offsetFromHeaderToVoxel > 0 {
	//	offset = make([]byte, offsetFromHeaderToVoxel, offsetFromHeaderToVoxel)
	//} else {
	//	offset = make([]byte, DefaultHeaderPadding, DefaultHeaderPadding)
	//}
	//
	//// Make a buffer and write the header to it with default system endian
	//hdrBuf := &bytes.Buffer{}
	//err := binary.Write(hdrBuf, system.NativeEndian, w.header)
	//if err != nil {
	//	return err
	//}
	//
	//bHeader := hdrBuf.Bytes()
	//bData := w.niiData.Volume
	//
	//var dataset []byte
	//dataset = append(dataset, bHeader...)
	//dataset = append(dataset, offset...)
	//dataset = append(dataset, bData...)

	dataset, err := w.reconstructDataset()
	if err != nil {
		return err
	}

	// Check if the user-specified filePath suffix is ending with '.nii' or '.gz'.
	// If not, we append '.nii' to the end to signify the file is NIfTI format
	if !strings.HasSuffix(w.filePath, NIFTI_EXT) && !strings.HasSuffix(w.filePath, NIFTI_COMPRESSED_EXT) {
		// If user specifies the file with '.gz' extension then default to compression
		if strings.HasSuffix(w.filePath, NIFTI_COMPRESSED_EXT) {
			w.compression = true
		}
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
	err = WriteToFile(w.filePath, w.compression, dataset)
	if err != nil {
		return err
	}
	return nil
}

// convertImageToNii1Header returns the header from a NIfTI image structure
func (w *NiiWriter) convertImageToNii1Header() error {
	if w.header != nil {
		return nil
	}

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

	// Load NIFTI specific stuff into the header
	if w.writeHeaderFile {
		header.Magic = NIFTI_1_MAGIC_PAIR // ni1
		header.VoxOffset = 0
	} else {
		header.Magic = NIFTI_1_MAGIC_SINGLE // n+1
		// This is for a case where we read the image as .hdr/.img pair but then want to write to a single file.
		// We have to update the VoxOffset value
		if int(header.VoxOffset)-int(header.SizeofHdr) <= 0 {
			header.VoxOffset = float32(header.SizeofHdr + DefaultHeaderPadding)
		}
	}

	w.header = header

	return nil
}

// convertImageToNii1Header returns the header from a NIfTI image structure
func (w *NiiWriter) convertImageToNii2Header() error {
	if w.header != nil {
		return nil
	}

	if w.niiData == nil {
		return errors.New("image data structure is nil")
	}

	header := new(Nii2Header)
	header.SizeofHdr = NII2HeaderSize

	header.Dim[0] = w.niiData.NDim
	header.Dim[1], header.Dim[2], header.Dim[3] = w.niiData.Nx, w.niiData.Ny, w.niiData.Nz
	header.Dim[4], header.Dim[5], header.Dim[6] = w.niiData.Nt, w.niiData.Nu, w.niiData.Nv
	header.Dim[7] = w.niiData.Nw

	header.Pixdim[0] = 0.0
	header.Pixdim[1], header.Pixdim[2], header.Pixdim[3] = math.Abs(w.niiData.Dx), math.Abs(w.niiData.Dy), math.Abs(w.niiData.Dz)
	header.Pixdim[4], header.Pixdim[5], header.Pixdim[6] = math.Abs(w.niiData.Dt), math.Abs(w.niiData.Du), math.Abs(w.niiData.Dv)
	header.Pixdim[7] = w.niiData.Dw

	header.Datatype = int16(w.niiData.Datatype)
	header.Bitpix = int16(8 * w.niiData.NByPer)

	if w.niiData.CalMax > w.niiData.CalMin {
		header.CalMin = w.niiData.CalMin
		header.CalMax = w.niiData.CalMax
	}

	if w.niiData.SclSlope != 0.0 {
		header.SclSlope = w.niiData.SclSlope
		header.SclInter = w.niiData.SclInter
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

	header.IntentCode = w.niiData.IntentCode
	header.IntentP1 = w.niiData.IntentP1
	header.IntentP2 = w.niiData.IntentP2
	header.IntentP3 = w.niiData.IntentP3
	if w.niiData.IntentName[0] != 0x0 {
		for i := 0; i < 15; i++ {
			header.IntentName[i] = w.niiData.IntentName[i]
		}
		header.IntentName[15] = 0x0
	}

	header.VoxOffset = int64(w.niiData.VoxOffset)
	header.XyztUnits = int32(convertSpaceTimeToXYZT(w.niiData.XYZUnits, w.niiData.TimeUnits))
	header.Toffset = w.niiData.TOffset

	if w.niiData.QformCode > 0 {
		header.QformCode = w.niiData.QformCode
		header.QuaternB = w.niiData.QuaternB
		header.QuaternC = w.niiData.QuaternC
		header.QuaternD = w.niiData.QuaternD

		header.QoffsetX = w.niiData.QoffsetX
		header.QoffsetY = w.niiData.QoffsetY
		header.QoffsetZ = w.niiData.QoffsetZ

		if w.niiData.QFac >= 0 {
			header.Pixdim[0] = 1.0
		} else {
			header.Pixdim[0] = -1.0
		}
	}

	if w.niiData.SformCode > 0 {
		header.SformCode = w.niiData.SformCode
		header.SrowX[0] = w.niiData.StoXYZ.M[0][0]
		header.SrowX[1] = w.niiData.StoXYZ.M[0][1]
		header.SrowX[2] = w.niiData.StoXYZ.M[0][2]
		header.SrowX[3] = w.niiData.StoXYZ.M[0][3]

		header.SrowY[0] = w.niiData.StoXYZ.M[1][0]
		header.SrowY[1] = w.niiData.StoXYZ.M[1][1]
		header.SrowY[2] = w.niiData.StoXYZ.M[1][2]
		header.SrowY[3] = w.niiData.StoXYZ.M[1][3]

		header.SrowZ[0] = w.niiData.StoXYZ.M[2][0]
		header.SrowZ[1] = w.niiData.StoXYZ.M[2][1]
		header.SrowZ[2] = w.niiData.StoXYZ.M[2][2]
		header.SrowZ[3] = w.niiData.StoXYZ.M[2][3]
	}

	header.DimInfo = convertFPSIntoDimInfo(w.niiData.FreqDim, w.niiData.PhaseDim, w.niiData.SliceDim)

	header.SliceCode = w.niiData.SliceCode
	header.SliceStart = w.niiData.SliceStart
	header.SliceEnd = w.niiData.SliceEnd
	header.SliceDuration = w.niiData.SliceDuration

	// Load NIFTI specific stuff into the header
	if w.writeHeaderFile {
		header.Magic = NIFTI_2_MAGIC_PAIR // ni2
		header.VoxOffset = 0
	} else {
		header.Magic = NIFTI_2_MAGIC_SINGLE // n+2
		// This is for a case where we read the image as .hdr/.img pair but then want to write to a single file.
		// We have to update the VoxOffset value
		header.VoxOffset = int64(header.SizeofHdr + DefaultHeaderPadding)
	}

	w.header = header

	return nil
}

// GetNiiData returns the current NIfTI image data
func (w *NiiWriter) GetNiiData() *Nii {
	return w.niiData
}

// GetHeader returns the current NIfTI header
func (w *NiiWriter) GetHeader() interface{} {
	return w.header
}
