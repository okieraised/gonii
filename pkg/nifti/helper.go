package nifti

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"

	gzip "github.com/klauspost/pgzip"
	"github.com/okieraised/gonii/internal/system"
)

// IsValidDatatype checks whether the datatype is valid for NIFTI format
func IsValidDatatype(datatype int32) bool {
	if ValidDatatype[datatype] {
		return true
	}
	return false
}

// SwapNIFTI1Header swaps all NIFTI fields
func SwapNIFTI1Header(header *Nii1Header) (*Nii1Header, error) {
	newHeader := new(Nii1Header)
	var err error

	newHeader.SizeofHdr = swapInt32(header.SizeofHdr)
	newHeader.Extents = swapInt32(header.Extents)
	newHeader.SessionError = swapInt16(header.SessionError)
	for i := 0; i < 8; i++ {
		newHeader.Dim[i] = swapInt16(header.Dim[i])
	}

	newHeader.IntentP1, err = swapFloat32(header.IntentP1)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap IntentP1: %v", err)
	}
	newHeader.IntentP2, err = swapFloat32(header.IntentP2)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap IntentP2: %v", err)
	}
	newHeader.IntentP3, err = swapFloat32(header.IntentP3)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap IntentP3: %v", err)
	}

	newHeader.IntentCode = swapInt16(header.IntentCode)
	newHeader.Datatype = swapInt16(header.Datatype)
	newHeader.Bitpix = swapInt16(header.Bitpix)
	newHeader.SliceStart = swapInt16(header.SliceStart)

	for i := 0; i < 8; i++ {
		newHeader.Pixdim[i], err = swapFloat32(header.Pixdim[i])
		if err != nil {
			return nil, fmt.Errorf("failed to byte swap Pixdim[%d]: %v", i, err)
		}
	}

	newHeader.VoxOffset, err = swapFloat32(header.VoxOffset)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap VoxOffset: %v", err)
	}

	newHeader.SclSlope, err = swapFloat32(header.SclSlope)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap SclSlope: %v", err)
	}

	newHeader.SclInter, err = swapFloat32(header.SclInter)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap SclInter: %v", err)
	}

	newHeader.SliceEnd = swapInt16(header.SliceEnd)

	newHeader.CalMin, err = swapFloat32(header.CalMin)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap CalMin: %v", err)
	}

	newHeader.CalMax, err = swapFloat32(header.CalMax)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap CalMax: %v", err)
	}

	newHeader.SliceDuration, err = swapFloat32(header.SliceDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap SliceDuration: %v", err)
	}

	newHeader.Glmin = swapInt32(header.Glmin)
	newHeader.Glmax = swapInt32(header.Glmax)

	newHeader.QformCode = swapInt16(header.QformCode)
	newHeader.SformCode = swapInt16(header.SformCode)

	newHeader.QuaternB, err = swapFloat32(header.QuaternB)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap QuaternB: %v", err)
	}

	newHeader.QuaternC, err = swapFloat32(header.QuaternC)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap QuaternC: %v", err)
	}

	newHeader.QuaternD, err = swapFloat32(header.QuaternD)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap QuaternD: %v", err)
	}

	newHeader.QoffsetX, err = swapFloat32(header.QoffsetX)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap QoffsetX: %v", err)
	}
	newHeader.QoffsetY, err = swapFloat32(header.QoffsetY)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap QoffsetY: %v", err)
	}
	newHeader.QoffsetZ, err = swapFloat32(header.QoffsetZ)
	if err != nil {
		return nil, fmt.Errorf("failed to byte swap QoffsetZ: %v", err)
	}

	for i := 0; i < 4; i++ {
		newHeader.SrowX[i], err = swapFloat32(header.SrowX[i])
		if err != nil {
			return nil, fmt.Errorf("failed to byte swap SrowX[%d]: %v", i, err)
		}
		newHeader.SrowY[i], err = swapFloat32(header.SrowY[i])
		if err != nil {
			return nil, fmt.Errorf("failed to byte swap SrowY[%d]: %v", i, err)
		}
		newHeader.SrowZ[i], err = swapFloat32(header.SrowZ[i])
		if err != nil {
			return nil, fmt.Errorf("failed to byte swap SrowZ[%d]: %v", i, err)
		}
	}
	return newHeader, nil
}

// getDatatype returns the appropriate datatype of the NIFTI image
func getDatatype(datatype int32) string {
	switch datatype {
	case DT_UNKNOWN:
		return "UNKNOWN"
	case DT_BINARY:
		return "BINARY"
	case DT_INT8:
		return "INT8"
	case DT_UINT8:
		return "UINT8"
	case DT_INT16:
		return "INT16"
	case DT_UINT16:
		return "UINT16"
	case DT_INT32:
		return "INT32"
	case DT_UINT32:
		return "UINT32"
	case DT_INT64:
		return "INT64"
	case DT_UINT64:
		return "UINT64"
	case DT_FLOAT32:
		return "FLOAT32"
	case DT_FLOAT64:
		return "FLOAT64"
	case DT_FLOAT128:
		return "FLOAT128"
	case DT_COMPLEX64:
		return "COMPLEX64"
	case DT_COMPLEX128:
		return "COMPLEX128"
	case DT_COMPLEX256:
		return "COMPLEX256"
	case DT_RGB24:
		return "RGB24"
	case DT_RGBA32:
		return "RGBA32"
	}
	return ILLEGAL
}

// getSliceCode returns the name of the slice code
func getSliceCode(sliceCode int32) string {
	switch sliceCode {
	case NIFTI_SLICE_UNKNOWN:
		return NiiSliceAcquistionInfo[NIFTI_SLICE_UNKNOWN]
	case NIFTI_SLICE_SEQ_INC:
		return NiiSliceAcquistionInfo[NIFTI_SLICE_SEQ_INC]
	case NIFTI_SLICE_SEQ_DEC:
		return NiiSliceAcquistionInfo[NIFTI_SLICE_SEQ_DEC]
	case NIFTI_SLICE_ALT_INC:
		return NiiSliceAcquistionInfo[NIFTI_SLICE_ALT_INC]
	case NIFTI_SLICE_ALT_DEC:
		return NiiSliceAcquistionInfo[NIFTI_SLICE_ALT_DEC]
	case NIFTI_SLICE_ALT_INC2:
		return NiiSliceAcquistionInfo[NIFTI_SLICE_ALT_INC2]
	case NIFTI_SLICE_ALT_DEC2:
		return NiiSliceAcquistionInfo[NIFTI_SLICE_ALT_DEC2]
	}

	return "UNKNOWN"
}

// AssignDatatypeSize sets the number of bytes per voxel and the swapsize based on a datatype code
// returns nByper and swapSize
func AssignDatatypeSize(datatype int32) (int16, int16) {
	var nByper, swapSize int16
	switch datatype {
	case DT_INT8, DT_UINT8:
		nByper = 1
		swapSize = 0
	case DT_INT16, DT_UINT16:
		nByper = 2
		swapSize = 2
	case DT_RGB24:
		nByper = 3
		swapSize = 0
	case DT_RGBA32:
		nByper = 4
		swapSize = 0
	case DT_INT32, DT_UINT32, DT_FLOAT32:
		nByper = 4
		swapSize = 4
	case DT_COMPLEX64:
		nByper = 8
		swapSize = 4
	case DT_FLOAT64, DT_INT64, DT_UINT64:
		nByper = 8
		swapSize = 8
	case DT_FLOAT128:
		nByper = 16
		swapSize = 16
	case DT_COMPLEX128:
		nByper = 16
		swapSize = 8
	case DT_COMPLEX256:
		nByper = 32
		swapSize = 16
	default:
	}
	return nByper, swapSize
}

// needHeaderSwap checks whether byte swapping is needed. dim0 should be in [0,7], and headerSize should be accurate.
//
// Returns:
//
// > 0 : needs swap
//
// = 0 : does not need swap
//
// < 0 : error condition
func needHeaderSwap(dim0 int16) int {
	d0 := dim0
	if d0 != 0 {
		if d0 > 0 && d0 < 7 {
			return 0
		}

		d0 = swapInt16(d0)
		if d0 > 0 && d0 < 7 {
			return 1
		}
		return -1
	}
	return -2
}

// swapInt16 swaps int16 from native endian to the other
func swapInt16(in int16) int16 {
	b := make([]byte, 2)

	switch system.NativeEndian {
	case binary.LittleEndian:
		binary.LittleEndian.PutUint16(b, uint16(in))
		return int16(binary.BigEndian.Uint16(b))
	default:
		binary.BigEndian.PutUint16(b, uint16(in))
		return int16(binary.LittleEndian.Uint16(b))
	}
}

// swapInt32 swaps int32 from native endian to the other
func swapInt32(in int32) int32 {
	b := make([]byte, 4)

	switch system.NativeEndian {
	case binary.LittleEndian:
		binary.LittleEndian.PutUint32(b, uint32(in))
		return int32(binary.BigEndian.Uint16(b))
	default:
		binary.BigEndian.PutUint32(b, uint32(in))
		return int32(binary.LittleEndian.Uint32(b))
	}
}

// swapInt64 swaps int64 from native endian to the other
func swapInt64(in int64) int64 {
	b := make([]byte, 8)

	switch system.NativeEndian {
	case binary.LittleEndian:
		binary.LittleEndian.PutUint64(b, uint64(in))
		return int64(binary.BigEndian.Uint64(b))
	default:
		binary.BigEndian.PutUint64(b, uint64(in))
		return int64(binary.LittleEndian.Uint64(b))
	}
}

// swapFloat32 swaps float32 from native endian to the other
func swapFloat32(in float32) (float32, error) {
	buf := new(bytes.Buffer)

	switch system.NativeEndian {
	case binary.LittleEndian:
		err := binary.Write(buf, binary.LittleEndian, in)
		if err != nil {
			return 0, err
		}
		bits := binary.BigEndian.Uint32(buf.Bytes())
		res := math.Float32frombits(bits)
		return res, nil
	default:
		err := binary.Write(buf, binary.BigEndian, in)
		if err != nil {
			return 0, err
		}
		bits := binary.LittleEndian.Uint32(buf.Bytes())
		res := math.Float32frombits(bits)
		return res, nil
	}
}

// swapFloat64 swaps float64 from native endian to the other
func swapFloat64(in float64) (float64, error) {
	buf := new(bytes.Buffer)

	switch system.NativeEndian {
	case binary.LittleEndian:
		err := binary.Write(buf, binary.LittleEndian, in)
		if err != nil {
			return 0, err
		}
		bits := binary.BigEndian.Uint64(buf.Bytes())
		res := math.Float64frombits(bits)
		return res, nil
	default:
		err := binary.Write(buf, binary.BigEndian, in)
		if err != nil {
			return 0, err
		}
		bits := binary.LittleEndian.Uint64(buf.Bytes())
		res := math.Float64frombits(bits)
		return res, nil
	}
}

func convertToF64(ar [4]float32) [4]float64 {
	newar := [4]float64{}
	var v float32
	var i int
	for i, v = range ar {
		newar[i] = float64(v)
	}
	return newar
}

func dimInfoToFreqDim(DimInfo uint8) uint8 {
	return DimInfo & 0x03
}

func dimInfoToPhaseDim(DimInfo uint8) uint8 {
	return (DimInfo >> 2) & 0x03
}

func dimInfoToSliceDim(DimInfo uint8) uint8 {
	return (DimInfo >> 4) & 0x03
}

// convertSpaceTimeToXYZT converts xyzUnit, timeUnit back to uint8 representation of XyztUnits field
func convertSpaceTimeToXYZT(xyzUnit, timeUnit int32) uint8 {
	return uint8((xyzUnit & 0x07) | (timeUnit & 0x38))
}

// convertFPSIntoDimInfo converts freqDim, phaseDim, sliceDim back to uint8 representation of DimInfo
func convertFPSIntoDimInfo(freqDim, phaseDim, sliceDim int32) uint8 {
	return uint8((freqDim & 0x03) | ((phaseDim & 0x03) << 2) | ((sliceDim & 0x03) << 4))
}

func MakeNewNii1Header(inDim *[8]int16, inDatatype int32) *Nii1Header {
	// Default Dim value
	defaultDim := [8]int16{3, 1, 1, 1, 1, 1, 1, 1}

	header := new(Nii1Header)
	var dim [8]int16

	// If no input Dim is provided then we use the default value
	if inDim != nil {
		dim = *inDim
	} else {
		dim = defaultDim
	}

	// validate Dim: if there is any problem, apply default Dim
	if dim[0] < 0 || dim[0] > 7 {
		dim = defaultDim
	} else {
		for c := 1; c <= int(dim[0]); c++ {
			if dim[c] < 1 {
				fmt.Printf("bad dim: %d: %d\n", c, dim[c])
				dim = defaultDim
				break
			}
		}
	}

	// Validate datatype
	datatype := inDatatype
	if !IsValidDatatype(datatype) {
		datatype = DT_FLOAT32
	}

	// Populate the header struct
	header.SizeofHdr = NII1HeaderSize
	header.Regular = 'r'

	// Init dim and pixdim
	header.Dim[0] = dim[0]
	header.Pixdim[0] = 0.0
	for c := 1; c <= int(dim[0]); c++ {
		header.Dim[c] = dim[c]
		header.Pixdim[c] = 1.0
	}

	header.Datatype = int16(datatype)

	nByper, _ := AssignDatatypeSize(datatype)
	header.Bitpix = 8 * nByper
	header.Magic = [4]byte{110, 43, 49, 0}

	return header
}

func MakeNewNii2Header(inDim *[8]int64, inDatatype int32) *Nii2Header {
	// Default Dim value
	defaultDim := [8]int64{3, 1, 1, 1, 1, 1, 1, 1}

	header := new(Nii2Header)
	var dim [8]int64

	// If no input Dim is provided then we use the default value
	if inDim != nil {
		dim = *inDim
	} else {
		dim = defaultDim
	}

	// validate Dim: if there is any problem, apply default Dim
	if dim[0] < 0 || dim[0] > 7 {
		dim = defaultDim
	} else {
		for c := 1; c <= int(dim[0]); c++ {
			if dim[c] < 1 {
				fmt.Printf("bad dim: %d: %d\n", c, dim[c])
				dim = defaultDim
				break
			}
		}
	}

	// Validate datatype
	datatype := inDatatype
	if !IsValidDatatype(datatype) {
		datatype = DT_FLOAT32
	}

	// Populate the header struct
	header.SizeofHdr = NII2HeaderSize

	// Init dim and pixdim
	header.Dim[0] = dim[0]
	header.Pixdim[0] = 0.0
	for c := 1; c <= int(dim[0]); c++ {
		header.Dim[c] = dim[c]
		header.Pixdim[c] = 1.0
	}

	header.Datatype = int16(datatype)

	nByper, _ := AssignDatatypeSize(datatype)
	header.Bitpix = 8 * nByper
	header.Magic = NIFTI_2_MAGIC_SINGLE

	return header
}

func MakeNewNiiImageFromHdr(hdr interface{}) (*Nii, error) {
	var img Nii
	switch h := hdr.(type) {
	case *Nii1Header:
		img.PixDim = [8]float64{
			float64(h.Pixdim[0]),
			float64(h.Pixdim[1]),
			float64(h.Pixdim[2]),
			float64(h.Pixdim[3]),
			float64(h.Pixdim[4]),
			float64(h.Pixdim[5]),
			float64(h.Pixdim[6]),
			float64(h.Pixdim[7]),
		}
		img.Dim = [8]int64{
			int64(h.Dim[0]),
			int64(h.Dim[1]),
			int64(h.Dim[2]),
			int64(h.Dim[3]),
			int64(h.Dim[4]),
			int64(h.Dim[5]),
			int64(h.Dim[6]),
			int64(h.Dim[7]),
		}
		img.NDim = int64(h.Dim[0])
		img.Nx = int64(h.Dim[1])
		img.Ny = int64(h.Dim[2])
		img.Nz = int64(h.Dim[3])
		img.Nt = int64(h.Dim[4])
		img.Nu = int64(h.Dim[5])
		img.Nv = int64(h.Dim[6])
		img.Nw = int64(h.Dim[7])
		img.Datatype = int32(h.Datatype)
		img.ByteOrder = binary.LittleEndian
		img.SclSlope = float64(h.SclSlope)
		img.SclInter = float64(h.SclInter)
		img.VoxOffset = float64(h.VoxOffset)
	case *Nii2Header:
		img.PixDim = h.Pixdim
		img.Dim = h.Dim
		img.NDim = h.Dim[0]
		img.Nx = h.Dim[1]
		img.Ny = h.Dim[2]
		img.Nz = h.Dim[3]
		img.Nt = h.Dim[4]
		img.Nu = h.Dim[5]
		img.Nv = h.Dim[6]
		img.Nw = h.Dim[7]
		img.Datatype = int32(h.Datatype)
		img.ByteOrder = binary.LittleEndian
		img.SclSlope = h.SclSlope
		img.SclInter = h.SclInter
		img.VoxOffset = float64(h.VoxOffset)
	default:
		return nil, fmt.Errorf("unknown header type")
	}

	return &img, nil
}

// MakeEmptyImageFromImg returns a zero-filled byte slice from existing Nii image structure
func MakeEmptyImageFromImg(img *Nii) ([]byte, error) {
	var bDataLength int64

	if img == nil {
		return nil, errors.New("NIfTI image structure nil")
	}

	// Need at least Nx, Ny
	if img.Nx == 0 {
		return nil, errors.New("x dimension must not be zero")
	}
	if img.Ny == 0 {
		return nil, errors.New("y dimension must not be zero")
	}
	bDataLength = img.Nx * img.Ny

	if img.Nz > 0 {
		bDataLength = bDataLength * img.Nz
	}
	if img.Nt > 0 {
		bDataLength = bDataLength * img.Nt
	}
	if img.Nu > 0 {
		bDataLength = bDataLength * img.Nu
	}
	if img.Nv > 0 {
		bDataLength = bDataLength * img.Nv
	}
	if img.Nw > 0 {
		bDataLength = bDataLength * img.Nw
	}

	nByper, _ := AssignDatatypeSize(img.Datatype)
	bDataLength = bDataLength * int64(nByper)

	// Init a slice of bytes with capacity of bDataLength and initial value of 0
	bData := make([]byte, bDataLength)

	return bData, nil
}

// MakeEmptyImageFromHdr initializes a zero-filled byte slice from existing header structure
func MakeEmptyImageFromHdr(hdr *Nii1Header) ([]byte, error) {
	var bDataLength int64

	if hdr == nil {
		return nil, errors.New("NIfTI image structure nil")
	}

	if hdr.Dim[1] == 0 {
		return nil, errors.New("x dimension must not be zero")
	}
	if hdr.Dim[2] == 0 {
		return nil, errors.New("y dimension must not be zero")
	}
	bDataLength = int64(hdr.Dim[1]) * int64(hdr.Dim[2])

	if hdr.Dim[3] > 0 {
		bDataLength = bDataLength * int64(hdr.Dim[3])
	}
	if hdr.Dim[4] > 0 {
		bDataLength = bDataLength * int64(hdr.Dim[4])
	}
	if hdr.Dim[5] > 0 {
		bDataLength = bDataLength * int64(hdr.Dim[5])
	}
	if hdr.Dim[6] > 0 {
		bDataLength = bDataLength * int64(hdr.Dim[6])
	}
	if hdr.Dim[7] > 0 {
		bDataLength = bDataLength * int64(hdr.Dim[7])
	}

	nByper, _ := AssignDatatypeSize(int32(hdr.Datatype))
	bDataLength = bDataLength * int64(nByper)

	// Init a slice of bytes with capacity of bDataLength and initial value of 0
	bData := make([]byte, bDataLength)

	return bData, nil
}

func uint64ToFloat64(v uint64, datatype int32) float64 {
	var value float64

	switch datatype {
	case DT_FLOAT64:
		value = float64(v)
	case DT_INT64:
		value = float64(int64(v))
	case DT_UINT64, DT_COMPLEX64:
		value = math.Float64frombits(v)
	}
	return value
}

func uint32ToFloat64(v uint32, datatype int32) float64 {
	var value float64

	switch datatype {
	case DT_INT32:
		value = float64(int32(v))
	case DT_UINT32:
		value = float64(v)
	case DT_FLOAT32, DT_RGBA32:
		value = float64(math.Float32frombits(v))
	}
	return value
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

// ConvertVoxelToBytes converts the voxel in float64 back to bytes slice based on datatype and NByPer
func ConvertVoxelToBytes(voxel, slope, intercept float64, datatype int32, binaryOrder binary.ByteOrder, nByPer int32) ([]byte, error) {
	// Check if we need to rescale
	if slope != 0 && datatype != DT_RGB24 {
		voxel = (voxel - intercept) / slope
	}

	switch nByPer {
	case 0:
		return nil, errors.New("nByPer is 0")
	case 1: // 1 byte per voxel includes Uint8 and Int8
		var buf bytes.Buffer
		switch datatype {
		case DT_UINT8:
			err := binary.Write(&buf, binaryOrder, uint8(voxel))
			if err != nil {
				return nil, err
			}
		case DT_INT8:
			err := binary.Write(&buf, binaryOrder, int8(voxel))
			if err != nil {
				return nil, err
			}
		}
		return buf.Bytes(), nil
	case 2: // This fits Uint16
		v := uint16(voxel)
		b := make([]byte, 2)
		switch binaryOrder {
		case binary.LittleEndian:
			binary.LittleEndian.PutUint16(b, v)
		case binary.BigEndian:
			binary.BigEndian.PutUint16(b, v)
		}
		return b, nil
	case 3: // This fits Uint32 -> RGB24
		v := math.Float32bits(float32(voxel))
		b := make([]byte, 4)
		switch binaryOrder {
		case binary.LittleEndian:
			binary.LittleEndian.PutUint32(b, v)
		case binary.BigEndian:
			binary.BigEndian.PutUint32(b, v)
		}
		return b[:3], nil
	case 4: // This fits Uint32
		v := math.Float32bits(float32(voxel))
		b := make([]byte, 4)
		switch binaryOrder {
		case binary.LittleEndian:
			binary.LittleEndian.PutUint32(b, v)
		case binary.BigEndian:
			binary.BigEndian.PutUint32(b, v)
		}
		return b, nil
	case 8:
		v := uint64(voxel)
		b := make([]byte, 8)
		switch binaryOrder {
		case binary.LittleEndian:
			binary.LittleEndian.PutUint64(b, v)
		case binary.BigEndian:
			binary.BigEndian.PutUint64(b, v)
		}
		return b, nil
	case 16: // Unsupported
	case 32: // Unsupported
	default:
	}
	return nil, errors.New("unsupported datatype")
}

func WriteToFile(filePath string, compression bool, dataset []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if compression { // If the compression is set to true, then write a compressed file
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

func RLEEncode(original []float64) ([]float64, error) {
	var rleEncoded []float64

	if len(original) == 0 {
		return nil, errors.New("array has length zero")
	}
	for i := 0; i < len(original); i++ {
		var count float64 = 1
		if i == 0 && original[i] != 0 {
			rleEncoded = append(rleEncoded, 0)
		}
		for {
			if i < len(original)-1 && original[i] == original[i+1] {
				count++
				i++
			} else {
				break
			}
		}
		rleEncoded = append(rleEncoded, count)
	}

	return rleEncoded, nil
}
