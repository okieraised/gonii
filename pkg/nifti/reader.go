package nifti

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/okieraised/gonii/pkg/matrix"
)

type Reader interface {
	// Parse parses the input file(s) and returns the input NIfTI as header and image data
	Parse() error
	// GetBinaryOrder returns the binary order of the NIfTI image
	GetBinaryOrder() binary.ByteOrder
	// GetNiiData returns the raw NIfTI header and image data
	GetNiiData() *Nii
	// GetHeader returns the NIfTI header
	GetHeader(prettyShow bool) interface{}
}

// NiiReader define the NIfTI reader structure.
type NiiReader struct {
	reader       *bytes.Reader
	hReader      *bytes.Reader
	binaryOrder  binary.ByteOrder // Default system order
	retainHeader bool             // Whether to keep the header after parsing
	inMemory     bool             // Whether to read the whole NIfTI image to memory
	data         *Nii             // Contains the NIFTI data structure
	header       interface{}      // Contains the NIFTI header
	version      int              // Define the version of NIFTI image (1 or 2)
}

func (r *NiiReader) SetBinaryOrder(bo binary.ByteOrder) {
	r.binaryOrder = bo
}

func (r *NiiReader) SetHdrReader(hdrRd *bytes.Reader) {
	r.hReader = hdrRd
}

func (r *NiiReader) SetReader(rd *bytes.Reader) {
	r.reader = rd
}

func (r *NiiReader) SetDataset(ds *Nii) {
	r.data = ds
}

func (r *NiiReader) SetRetainHeader(retainHeader bool) {
	r.retainHeader = retainHeader
}

func (r *NiiReader) SetInMemory(inMemory bool) {
	r.inMemory = inMemory
}

func (r *NiiReader) GetHeader(prettyShow bool) interface{} {
	if r.header != nil {
		if r.version == NIIVersion1 {
			hdr := r.header.(*Nii1Header)
			if prettyShow {
				fmt.Println(prettyPrint(hdr))
			}
			return hdr
		}
		if r.version == NIIVersion2 {
			hdr := r.header.(*Nii2Header)
			if prettyShow {
				fmt.Println(prettyPrint(hdr))
			}
			return hdr
		}
	}
	return r.header
}

// GetVersion returns the NIfTI version based on the header information
func (r *NiiReader) GetVersion() int {
	return r.version
}

// GetNiiData returns the NIfTI image structure
func (r *NiiReader) GetNiiData() *Nii {
	return r.data
}

// GetBinaryOrder returns the NIfTI file binary order
func (r *NiiReader) GetBinaryOrder() binary.ByteOrder {
	return r.binaryOrder
}

// Parse returns the raw byte array into NIfTI-1/2 header and dataset structure
func (r *NiiReader) Parse() error {
	err := r.getVersion()
	if err != nil {
		return err
	}

	err = r.parseNIfTI()
	if err != nil {
		return err
	}
	return nil
}

// parseNIfTI parse the NIfTI header and the data
func (r *NiiReader) parseNIfTI() error {
	var hReader *bytes.Reader
	if r.hReader != nil {
		hReader = r.hReader
	} else {
		hReader = r.reader
	}

	_, err := hReader.Seek(0, 0)
	if err != nil {
		return err
	}

	var dim0 int64
	var header interface{}

	switch r.version {
	case NIIVersion1:
		n1Header := new(Nii1Header)
		err = binary.Read(hReader, r.binaryOrder, n1Header)
		if err != nil {
			return err
		}
		if n1Header.Magic != NIFTI_1_MAGIC_SINGLE && n1Header.Magic != NIFTI_1_MAGIC_PAIR {
			return errors.New("invalid NIFTI-1 magic string")
		}
		dim0 = int64(n1Header.Dim[0])

		if dim0 < 0 || dim0 > 7 {
			if r.binaryOrder == binary.LittleEndian {
				r.binaryOrder = binary.BigEndian
			} else {
				r.binaryOrder = binary.LittleEndian
			}
		}
		header = n1Header
	case NIIVersion2:
		n2Header := new(Nii2Header)
		err = binary.Read(hReader, r.binaryOrder, n2Header)
		if err != nil {
			return err
		}
		if n2Header.Magic != NIFTI_2_MAGIC_SINGLE && n2Header.Magic != NIFTI_2_MAGIC_PAIR {
			return errors.New("invalid NIFTI-2 magic string")
		}
		dim0 = n2Header.Dim[0]

		if dim0 < 0 || dim0 > 7 {
			if r.binaryOrder == binary.LittleEndian {
				r.binaryOrder = binary.BigEndian
			} else {
				r.binaryOrder = binary.LittleEndian
			}
		}
		header = n2Header
	default:
		return errors.New("invalid version")
	}
	err = r.parseData(header)
	if err != nil {
		return err
	}

	if r.retainHeader {
		r.header = header
	}

	return nil
}

// parseData parse the raw byte array into NIFTI-1 or NIFTI-2 data structure
func (r *NiiReader) parseData(header interface{}) error {
	var statDim int64 = 1
	var bitpix int16
	var qFormCode, sFormCode, intentCode, sliceCode, datatype, freqDim, phaseDim, sliceDim int32
	var pixDim0, sclSlope, sclInter, intentP1, intentP2, intentP3, quaternB, quaternC, quaternD, sliceDuration, calMin, calMax float64
	var sRowX, sRowY, sRowZ [4]float64
	var intentName [16]uint8
	var descrip [80]uint8
	var sliceStart, sliceEnd, voxOffset int64

	switch r.version {
	case NIIVersion1:
		n1Header := header.(*Nii1Header)

		freqDim = int32(dimInfoToFreqDim(n1Header.DimInfo))
		phaseDim = int32(dimInfoToPhaseDim(n1Header.DimInfo))
		sliceDim = int32(dimInfoToSliceDim(n1Header.DimInfo))

		voxOffset = int64(n1Header.VoxOffset)
		datatype = int32(n1Header.Datatype)

		// The bits 1-3 are used to store the spatial dimensions, the bits 4-6 are for temporal dimensions,
		// and the bits 6 and 7 are not used
		r.data.XYZUnits = int32(n1Header.XyztUnits % 8)
		r.data.TimeUnits = int32(n1Header.XyztUnits) - r.data.XYZUnits

		sliceCode = int32(n1Header.SliceCode)
		sliceStart = int64(n1Header.SliceStart)
		sliceEnd = int64(n1Header.SliceEnd)
		sliceDuration = float64(n1Header.SliceDuration)

		calMin = float64(n1Header.CalMin)
		calMax = float64(n1Header.CalMax)

		qFormCode = int32(n1Header.QformCode)
		sFormCode = int32(n1Header.SformCode)
		pixDim0 = float64(n1Header.Pixdim[0])

		sRowX = convertToF64(n1Header.SrowX)
		sRowY = convertToF64(n1Header.SrowY)
		sRowZ = convertToF64(n1Header.SrowZ)

		sclSlope = float64(n1Header.SclSlope)
		sclInter = float64(n1Header.SclInter)

		intentName = n1Header.IntentName
		intentCode = int32(n1Header.IntentCode)
		intentP1 = float64(n1Header.IntentP1)
		intentP2 = float64(n1Header.IntentP2)
		intentP3 = float64(n1Header.IntentP3)

		quaternB = float64(n1Header.QuaternB)
		quaternC = float64(n1Header.QuaternC)
		quaternD = float64(n1Header.QuaternD)
		descrip = n1Header.Descrip

		// Set the dimension of data array
		r.data.NDim, r.data.Dim[0] = int64(n1Header.Dim[0]), int64(n1Header.Dim[0])
		r.data.Nx, r.data.Dim[1] = int64(n1Header.Dim[1]), int64(n1Header.Dim[1])
		r.data.Ny, r.data.Dim[2] = int64(n1Header.Dim[2]), int64(n1Header.Dim[2])
		r.data.Nz, r.data.Dim[3] = int64(n1Header.Dim[3]), int64(n1Header.Dim[3])
		r.data.Nt, r.data.Dim[4] = int64(n1Header.Dim[4]), int64(n1Header.Dim[4])
		r.data.Nu, r.data.Dim[5] = int64(n1Header.Dim[5]), int64(n1Header.Dim[5])
		r.data.Nv, r.data.Dim[6] = int64(n1Header.Dim[6]), int64(n1Header.Dim[6])
		r.data.Nw, r.data.Dim[7] = int64(n1Header.Dim[7]), int64(n1Header.Dim[7])

		// Set the grid spacing
		r.data.Dx, r.data.PixDim[1] = float64(n1Header.Pixdim[1]), float64(n1Header.Pixdim[1])
		r.data.Dy, r.data.PixDim[2] = float64(n1Header.Pixdim[2]), float64(n1Header.Pixdim[2])
		r.data.Dz, r.data.PixDim[3] = float64(n1Header.Pixdim[3]), float64(n1Header.Pixdim[3])
		r.data.Dt, r.data.PixDim[4] = float64(n1Header.Pixdim[4]), float64(n1Header.Pixdim[4])
		r.data.Du, r.data.PixDim[5] = float64(n1Header.Pixdim[5]), float64(n1Header.Pixdim[5])
		r.data.Dv, r.data.PixDim[6] = float64(n1Header.Pixdim[6]), float64(n1Header.Pixdim[6])
		r.data.Dw, r.data.PixDim[7] = float64(n1Header.Pixdim[7]), float64(n1Header.Pixdim[7])

		bitpix = n1Header.Bitpix

		NByPerVoxel, SwapSize := AssignDatatypeSize(datatype)
		r.data.NByPer = int32(NByPerVoxel)
		r.data.SwapSize = int32(SwapSize)

		r.data.QuaternB, r.data.QuaternC, r.data.QuaternD = float64(n1Header.QuaternB), float64(n1Header.QuaternC), float64(n1Header.QuaternD)
		r.data.QoffsetX, r.data.QoffsetY, r.data.QoffsetZ = float64(n1Header.QoffsetX), float64(n1Header.QoffsetY), float64(n1Header.QoffsetZ)

		r.data.AuxFile = n1Header.AuxFile

	case NIIVersion2:
		n2Header := header.(*Nii2Header)

		freqDim = int32(dimInfoToFreqDim(n2Header.DimInfo))
		phaseDim = int32(dimInfoToPhaseDim(n2Header.DimInfo))
		sliceDim = int32(dimInfoToSliceDim(n2Header.DimInfo))

		voxOffset = n2Header.VoxOffset
		datatype = int32(n2Header.Datatype)

		// The bits 1-3 are used to store the spatial dimensions, the bits 4-6 are for temporal dimensions,
		// and the bits 6 and 7 are not used
		r.data.XYZUnits = n2Header.XyztUnits % 8
		r.data.TimeUnits = n2Header.XyztUnits - r.data.XYZUnits

		sliceCode = n2Header.SliceCode
		sliceStart = n2Header.SliceStart
		sliceEnd = n2Header.SliceEnd
		sliceDuration = n2Header.SliceDuration

		calMin = n2Header.CalMin
		calMax = n2Header.CalMax

		qFormCode = n2Header.QformCode
		pixDim0 = n2Header.Pixdim[0]
		sFormCode = n2Header.SformCode

		sclSlope = n2Header.SclSlope
		sclInter = n2Header.SclInter

		intentName = n2Header.IntentName
		intentCode = n2Header.IntentCode
		r.data.IntentP1 = n2Header.IntentP1
		r.data.IntentP2 = n2Header.IntentP2
		r.data.IntentP3 = n2Header.IntentP3

		r.data.QuaternB = n2Header.QuaternB
		r.data.QuaternC = n2Header.QuaternC
		r.data.QuaternD = n2Header.QuaternD
		descrip = n2Header.Descrip

		// Set the dimension of data array
		r.data.NDim, r.data.Dim[0] = n2Header.Dim[0], n2Header.Dim[0]
		r.data.Nx, r.data.Dim[1] = n2Header.Dim[1], n2Header.Dim[1]
		r.data.Ny, r.data.Dim[2] = n2Header.Dim[2], n2Header.Dim[2]
		r.data.Nz, r.data.Dim[3] = n2Header.Dim[3], n2Header.Dim[3]
		r.data.Nt, r.data.Dim[4] = n2Header.Dim[4], n2Header.Dim[4]
		r.data.Nu, r.data.Dim[5] = n2Header.Dim[5], n2Header.Dim[5]
		r.data.Nv, r.data.Dim[6] = n2Header.Dim[6], n2Header.Dim[6]
		r.data.Nw, r.data.Dim[7] = n2Header.Dim[7], n2Header.Dim[7]

		// Set the grid spacing
		r.data.Dx, r.data.PixDim[1] = n2Header.Pixdim[1], n2Header.Pixdim[1]
		r.data.Dy, r.data.PixDim[2] = n2Header.Pixdim[2], n2Header.Pixdim[2]
		r.data.Dz, r.data.PixDim[3] = n2Header.Pixdim[3], n2Header.Pixdim[3]
		r.data.Dt, r.data.PixDim[4] = n2Header.Pixdim[4], n2Header.Pixdim[4]
		r.data.Du, r.data.PixDim[5] = n2Header.Pixdim[5], n2Header.Pixdim[5]
		r.data.Dv, r.data.PixDim[6] = n2Header.Pixdim[6], n2Header.Pixdim[6]
		r.data.Dw, r.data.PixDim[7] = n2Header.Pixdim[7], n2Header.Pixdim[7]

		bitpix = n2Header.Bitpix

		// SRowX, SRowY, SRowZ
		sRowX, sRowY, sRowZ = n2Header.SrowX, n2Header.SrowY, n2Header.SrowZ

		NByPerVoxel, SwapSize := AssignDatatypeSize(datatype)
		r.data.NByPer = int32(NByPerVoxel)
		r.data.SwapSize = int32(SwapSize)

		r.data.QuaternB, r.data.QuaternC, r.data.QuaternD = n2Header.QuaternB, n2Header.QuaternC, n2Header.QuaternD
		r.data.QoffsetX, r.data.QoffsetY, r.data.QoffsetZ = n2Header.QoffsetX, n2Header.QoffsetY, n2Header.QoffsetZ

		r.data.AuxFile = n2Header.AuxFile
	}

	// Fix bad value in header
	if r.data.Nz <= 0 && r.data.Dim[3] <= 0 {
		r.data.Nz = 1
		r.data.Dim[3] = 1
	}
	if r.data.Nt <= 0 && r.data.Dim[4] <= 0 {
		r.data.Nt = 1
		r.data.Dim[4] = 1
	}
	if r.data.Nu <= 0 && r.data.Dim[5] <= 0 {
		r.data.Nu = 1
		r.data.Dim[5] = 1
	}
	if r.data.Nv <= 0 && r.data.Dim[6] <= 0 {
		r.data.Nv = 1
		r.data.Dim[6] = 1
	}
	if r.data.Nw <= 0 && r.data.Dim[7] <= 0 {
		r.data.Nw = 1
		r.data.Dim[7] = 1
	}

	// Set the byte order
	r.data.ByteOrder = r.binaryOrder

	if bitpix == 0 {
		return errors.New("number of bits per voxel value (bitpix) is zero")
	}

	r.data.NVox = 1
	for i := int64(1); i <= r.data.NDim; i++ {
		r.data.NVox *= r.data.Dim[i]
	}

	// compute QToXYK transformation from pixel indexes (i,j,k) to (x,y,z)
	if qFormCode <= 0 {
		r.data.QtoXYZ.M[0][0] = r.data.Dx
		r.data.QtoXYZ.M[1][1] = r.data.Dy
		r.data.QtoXYZ.M[2][2] = r.data.Dz

		// off diagonal is zero
		r.data.QtoXYZ.M[0][1] = 0
		r.data.QtoXYZ.M[0][2] = 0
		r.data.QtoXYZ.M[0][3] = 0

		r.data.QtoXYZ.M[1][0] = 0
		r.data.QtoXYZ.M[1][2] = 0
		r.data.QtoXYZ.M[1][3] = 0

		r.data.QtoXYZ.M[2][0] = 0
		r.data.QtoXYZ.M[2][1] = 0
		r.data.QtoXYZ.M[2][3] = 0

		// last row is [0, 0, 0, 1]
		r.data.QtoXYZ.M[3][0] = 0
		r.data.QtoXYZ.M[3][1] = 0
		r.data.QtoXYZ.M[3][2] = 0
		r.data.QtoXYZ.M[3][3] = 1.0

		r.data.QformCode = NIFTI_XFORM_UNKNOWN
	} else {
		if pixDim0 < 0 {
			r.data.QFac = -1
		} else {
			r.data.QFac = 1
		}
		r.data.QtoXYZ = r.data.QuaternToMatrix()
		r.data.QformCode = qFormCode
	}

	// Set QToIJK
	r.data.QtoIJK = matrix.Mat44Inverse(r.data.QtoXYZ)

	if sFormCode <= 0 {
		r.data.SformCode = NIFTI_XFORM_UNKNOWN
	} else {
		r.data.StoXYZ.M[0][0] = sRowX[0]
		r.data.StoXYZ.M[0][1] = sRowX[1]
		r.data.StoXYZ.M[0][2] = sRowX[2]
		r.data.StoXYZ.M[0][3] = sRowX[3]

		r.data.StoXYZ.M[1][0] = sRowY[0]
		r.data.StoXYZ.M[1][1] = sRowY[1]
		r.data.StoXYZ.M[1][2] = sRowY[2]
		r.data.StoXYZ.M[1][3] = sRowY[3]

		r.data.StoXYZ.M[2][0] = sRowZ[0]
		r.data.StoXYZ.M[2][1] = sRowZ[1]
		r.data.StoXYZ.M[2][2] = sRowZ[2]
		r.data.StoXYZ.M[2][3] = sRowZ[3]

		r.data.StoXYZ.M[3][0] = 0
		r.data.StoXYZ.M[3][1] = 0
		r.data.StoXYZ.M[3][2] = 0
		r.data.StoXYZ.M[3][3] = 1

		r.data.StoIJK = matrix.Mat44Inverse(r.data.StoXYZ)

		r.data.SformCode = sFormCode
	}

	// Other stuff
	r.data.SclSlope = sclSlope
	r.data.SclInter = sclInter

	r.data.IntentName = intentName
	r.data.IntentCode = intentCode
	r.data.IntentP1 = intentP1
	r.data.IntentP2 = intentP2
	r.data.IntentP3 = intentP3

	r.data.QuaternB = quaternB
	r.data.QuaternC = quaternC
	r.data.QuaternD = quaternD
	r.data.Descrip = descrip

	// Frequency dimension, phase dimension, slice dimension
	r.data.FreqDim = freqDim
	r.data.PhaseDim = phaseDim
	r.data.SliceDim = sliceDim

	r.data.SliceCode = sliceCode
	r.data.SliceStart = sliceStart
	r.data.SliceEnd = sliceEnd
	r.data.SliceDuration = sliceDuration

	r.data.CalMin = calMin
	r.data.CalMax = calMax

	r.data.Datatype = datatype

	if r.data.Dim[5] > 1 {
		statDim = r.data.Dim[5]
	}

	r.data.VoxOffset = float64(voxOffset)
	dataSize := r.data.Dim[1] * r.data.Dim[2] * r.data.Dim[3] * r.data.Dim[4] * statDim * (int64(bitpix) / 8)

	_, err := r.reader.Seek(voxOffset, 0)
	if err != nil {
		return err
	}

	buf := make([]byte, dataSize)
	_, err = io.ReadFull(r.reader, buf)
	if err != nil {
		return err
	}
	r.data.Volume = buf

	affine := matrix.DMat44{}
	affine.M[0] = sRowX
	affine.M[1] = sRowY
	affine.M[2] = sRowZ
	affine.M[3] = [4]float64{0, 0, 0, 1}

	r.data.Affine = affine
	r.data.MatrixToOrientation(affine)

	return nil
}

// getVersion checks the header to determine the NIFTI version
func (r *NiiReader) getVersion() error {
	var hSize int32
	var hReader *bytes.Reader

	if r.hReader != nil {
		hReader = r.hReader
	} else {
		hReader = r.reader
	}

	err := binary.Read(hReader, r.binaryOrder, &hSize)
	if err != nil {
		return err
	}

	switch hSize {
	case NII1HeaderSize:
		r.version = NIIVersion1
	case NII2HeaderSize:
		r.version = NIIVersion2
	default:
		r.binaryOrder = binary.BigEndian
		_, err := hReader.Seek(0, 0)
		if err != nil {
			return err
		}
		var hSize int32
		err = binary.Read(hReader, r.binaryOrder, &hSize)
		if err != nil {
			return err
		}
		switch hSize {
		case NII1HeaderSize:
			r.version = NIIVersion1
		case NII2HeaderSize:
			r.version = NIIVersion2
		default:
			return errors.New("invalid NIFTI file format")
		}
	}
	r.data.Version = r.version
	return nil
}
