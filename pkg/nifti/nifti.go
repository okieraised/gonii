package nifti

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/okieraised/gonii/pkg/matrix"
	"math"
	"strings"
)

const (
	INVALID = "INVALID"
	UNKNOWN = "UNKNOWN"
	ILLEGAL = "ILLEGAL"
)

// Nii defines the structure of the NIFTI-1 data for I/O purpose
type Nii struct {
	NDim          int64            // last dimension greater than 1 (1..7)
	Nx            int64            // dimensions of grid array
	Ny            int64            // dimensions of grid array
	Nz            int64            // dimensions of grid array
	Nt            int64            // dimensions of grid array
	Nu            int64            // dimensions of grid array
	Nv            int64            // dimensions of grid array
	Nw            int64            // dimensions of grid array
	Dim           [8]int64         // dim[0] = ndim, dim[1] = nx, etc
	NVox          int64            // number of voxels = nx*ny*nz*...*nw
	NByPer        int32            // bytes per voxel, matches datatype (Datatype)
	Datatype      int32            // type of data in voxels: DT_* code
	Dx            float64          // grid spacings
	Dy            float64          // grid spacings
	Dz            float64          // grid spacings
	Dt            float64          // grid spacings
	Du            float64          // grid spacings
	Dv            float64          // grid spacings
	Dw            float64          // grid spacings tEStataILSTERIOn
	PixDim        [8]float64       // pixdim[1]=dx, etc
	SclSlope      float64          // scaling parameter: slope
	SclInter      float64          // scaling parameter: intercept
	CalMin        float64          // calibration parameter: minimum
	CalMax        float64          // calibration parameter: maximum
	QformCode     int32            // codes for (x,y,z) space meaning
	SformCode     int32            // codes for (x,y,z) space meaning
	FreqDim       int32            // indices (1,2,3, or 0) for MRI
	PhaseDim      int32            // directions in dim[]/pixdim[]
	SliceDim      int32            // directions in dim[]/pixdim[]
	SliceCode     int32            // code for slice timing pattern
	SliceStart    int64            // index for start of slices
	SliceEnd      int64            // index for end of slices
	SliceDuration float64          // time between individual slices
	QuaternB      float64          // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QuaternC      float64          // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QuaternD      float64          // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QoffsetX      float64          // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QoffsetY      float64          // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QoffsetZ      float64          // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QFac          float64          // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QtoXYZ        matrix.DMat44    // qform: transform (i,j,k) to (x,y,z)
	QtoIJK        matrix.DMat44    // qform: transform (x,y,z) to (i,j,k)
	StoXYZ        matrix.DMat44    // sform: transform (i,j,k) to (x,y,z)
	StoIJK        matrix.DMat44    // sform: transform (x,y,z) to (i,j,k)
	TOffset       float64          // time coordinate offset
	XYZUnits      int32            // dx,dy,dz units: NIFTI_UNITS_* code
	TimeUnits     int32            // dt units: NIFTI_UNITS_* code
	NiftiType     int32            // 0==Analyze, 1==NIFTI-1 (file), 2==NIFTI-1 (2 files), 3==NIFTI-ASCII (1 file)
	IntentCode    int32            // statistic type (or something)
	IntentP1      float64          // intent parameters
	IntentP2      float64          // intent parameters
	IntentP3      float64          // intent parameters
	IntentName    [16]byte         // optional description of intent data
	Descrip       [80]byte         // optional text to describe dataset
	AuxFile       [24]byte         // auxiliary filename
	FName         *byte            // header filename
	IName         *byte            // image filename
	INameOffset   int32            // offset into IName where data start
	SwapSize      int32            // swap unit in image data (might be 0)
	ByteOrder     binary.ByteOrder // byte order on disk (MSB_ or LSB_FIRST)
	Volume        []byte           // slice of data: nbyper*nvox bytes
	NumExt        int32            // number of extensions in extList
	Nifti1Ext     []Nifti1Ext      // array of extension structs (with data)
	IJKOrient     [3]int32         // self-add. Orientation ini, j, k coordinate
	Affine        matrix.DMat44    // self-add. Affine matrix
	VoxOffset     float64          // self-add. Voxel offset
	Version       int              // self-add. Used for version identification when writing
}

type Nifti1Ext struct {
	ECode int32
	Edata []byte
	ESize int32
}

//----------------------------------------------------------------------------------------------------------------------
// Get methods
//----------------------------------------------------------------------------------------------------------------------

// getSliceCode returns the slice code of the NIFTI image
func (n *Nii) getSliceCode() string {
	return getSliceCode(n.SliceCode)
}

// getQFormCode returns the QForm code
func (n *Nii) getQFormCode() string {
	qForm, ok := NiiPatientOrientationInfo[n.QformCode]
	if !ok {
		return INVALID
	}
	return qForm
}

// getSFormCode returns the SForm code
func (n *Nii) getSFormCode() string {
	sForm, ok := NiiPatientOrientationInfo[n.SformCode]
	if !ok {
		return INVALID
	}
	return sForm
}

// getDatatype returns the corresponding NIfTI datatype
func (n *Nii) getDatatype() string {
	return getDatatype(n.Datatype)
}

// getOrientation returns the image orientation
func (n *Nii) getOrientation() [3]string {
	res := [3]string{}

	ijk := n.IJKOrient

	iOrient, ok := OrietationToString[int(ijk[0])]
	if !ok {
		res[0] = OrietationToString[NIFTI_UNKNOWN_ORIENT]
	}
	res[0] = iOrient

	jOrient, ok := OrietationToString[int(ijk[1])]
	if !ok {
		res[1] = OrietationToString[NIFTI_UNKNOWN_ORIENT]
	}
	res[1] = jOrient

	kOrient, ok := OrietationToString[int(ijk[2])]
	if !ok {
		res[2] = OrietationToString[NIFTI_UNKNOWN_ORIENT]
	}
	res[2] = kOrient

	return res
}

func (n *Nii) getVoxel() *Voxels {
	vox := NewVoxels(n.Nx, n.Ny, n.Nz, n.Nt, n.Datatype)
	for x := int64(0); x < n.Nx; x++ {
		for y := int64(0); y < n.Ny; y++ {
			for z := int64(0); z < n.Nz; z++ {
				for t := int64(0); t < n.Nt; t++ {
					vox.Set(x, y, z, t, n.getAt(x, y, z, t))
				}
			}
		}
	}
	return vox
}

// getAt returns the value at (x, y, z, t) location
func (n *Nii) getAt(x, y, z, t int64) float64 {
	tIndex := t * n.Nx * n.Ny * n.Nz
	zIndex := n.Nx * n.Ny * z
	yIndex := n.Nx * y
	xIndex := x
	index := tIndex + zIndex + yIndex + xIndex
	nByPer := int64(n.NByPer)

	dataPoint := n.Volume[index*nByPer : (index+1)*nByPer]

	var value float64
	switch n.NByPer {
	case 0, 1:
		if len(dataPoint) > 0 {
			value = float64(dataPoint[0])
		}
	case 2: // This fits Uint16
		var v uint16
		switch n.ByteOrder {
		case binary.LittleEndian:
			v = binary.LittleEndian.Uint16(dataPoint)
		case binary.BigEndian:
			v = binary.BigEndian.Uint16(dataPoint)
		}
		switch n.Datatype {
		case DT_INT16:
			value = float64(int16(v))
		case DT_UINT16:
			value = float64(v)
		}
	case 3, 4: // This fits Uint32
		var v uint32
		switch n.ByteOrder {
		case binary.LittleEndian:
			switch len(dataPoint) {
			case 3:
				v = uint32(dataPoint[0]) | uint32(dataPoint[1])<<8 | uint32(dataPoint[2])<<16
				value = float64(math.Float32frombits(v))
			case 4:
				v = binary.LittleEndian.Uint32(dataPoint)
				value = uint32ToFloat64(v, n.Datatype)
			}
		case binary.BigEndian:
			switch len(dataPoint) {
			case 3:
				v = uint32(dataPoint[2]) | uint32(dataPoint[1])<<8 | uint32(dataPoint[0])<<16
				value = float64(math.Float32frombits(v))
			case 4:
				v = binary.BigEndian.Uint32(dataPoint)
				value = uint32ToFloat64(v, n.Datatype)
			}
		}
	case 8:
		var v uint64
		switch n.ByteOrder {
		case binary.LittleEndian:
			v = binary.LittleEndian.Uint64(dataPoint)
		case binary.BigEndian:
			v = binary.BigEndian.Uint64(dataPoint)
		}
		value = uint64ToFloat64(v, n.Datatype)
	case 16: // Unsupported
	case 32: // Unsupported
	default:
	}

	if n.SclSlope != 0 && n.Datatype != DT_RGB24 {
		value = n.SclSlope*value + n.SclInter
	}
	return value
}

// getTimeSeries returns the time-series of a point
func (n *Nii) getTimeSeries(x, y, z int64) ([]float64, error) {
	timeSeries := make([]float64, 0, n.Dim[4])

	sliceX := n.Nx
	sliceY := n.Ny
	sliceZ := n.Nx

	if x >= sliceX {
		return nil, fmt.Errorf("invalid x value %d", x)
	}

	if y >= sliceY {
		return nil, fmt.Errorf("invalid y value %d", y)
	}

	if z >= sliceZ {
		return nil, fmt.Errorf("invalid z value %d", z)
	}

	for t := 0; t < int(n.Dim[4]); t++ {
		timeSeries = append(timeSeries, n.getAt(x, y, z, int64(t)))
	}
	return timeSeries, nil
}

// getSlice returns the image in x-y dimension
func (n *Nii) getSlice(z, t int64) ([][]float64, error) {
	sliceX := n.Nx
	sliceY := n.Ny
	sliceZ := n.Nz
	sliceT := n.Nt

	if z >= sliceZ {
		return nil, fmt.Errorf("invalid z value %d", z)
	}

	if t >= sliceT || t < 0 {
		return nil, fmt.Errorf("invalid time value %d", t)
	}

	slice := make([][]float64, sliceX)
	for i := range slice {
		slice[i] = make([]float64, sliceY)
	}
	for x := 0; x < int(sliceX); x++ {
		for y := 0; y < int(sliceY); y++ {
			slice[x][y] = n.getAt(int64(x), int64(y), z, t)
		}
	}
	return slice, nil
}

// getVolume return the whole image volume at time t
func (n *Nii) getVolume(t int64) ([][][]float64, error) {
	sliceX := n.Nx
	sliceY := n.Ny
	sliceZ := n.Nz
	sliceT := n.Nt

	if t >= sliceT || t < 0 {
		return nil, fmt.Errorf("invalid time value %d", t)
	}
	volume := make([][][]float64, sliceX)
	for i := range volume {
		volume[i] = make([][]float64, sliceY)
		for j := range volume[i] {
			volume[i][j] = make([]float64, sliceZ)
		}
	}
	for x := 0; x < int(sliceX); x++ {
		for y := 0; y < int(sliceY); y++ {
			for z := 0; z < int(sliceZ); z++ {
				volume[x][y][z] = n.getAt(int64(x), int64(y), int64(z), t)
			}
		}
	}
	return volume, nil
}

// getUnitsOfMeasurements returns the spatial and temporal units of measurements
func (n *Nii) getUnitsOfMeasurements() ([2]string, error) {
	units := [2]string{}
	spatialUnit, ok := NiiMeasurementUnits[uint8(n.XYZUnits)]
	if !ok {
		return units, fmt.Errorf("invalid spatial unit %d", n.XYZUnits)
	}

	temporalUnit, ok := NiiMeasurementUnits[uint8(n.TimeUnits)]
	if !ok {
		return units, fmt.Errorf("invalid temporal unit %d", n.TimeUnits)
	}

	units[0] = spatialUnit
	units[1] = temporalUnit

	return units, nil
}

// getAffine returns the 4x4 affine matrix
func (n *Nii) getAffine() matrix.DMat44 {
	return n.Affine
}

// getImgShape returns the image shape in terms of x, y, z, t
func (n *Nii) getImgShape() [4]int64 {
	dim := [4]int64{}

	for index, _ := range dim {
		dim[index] = n.Dim[index+1]
	}
	return dim
}

// getVoxelSize returns the voxel size of the image
func (n *Nii) getVoxelSize() [4]float64 {
	size := [4]float64{}
	for index, _ := range size {
		size[index] = n.PixDim[index+1]
	}
	return size
}

// getDescrip returns the description with trailing null bytes removed
func (n *Nii) getDescrip() string {
	return strings.ReplaceAll(string(n.Descrip[:]), "\x00", "")
}

// getIntentName returns the intent name with trailing null bytes removed
func (n *Nii) getIntentName() string {
	return strings.ReplaceAll(string(n.IntentName[:]), "\x00", "")
}

// getSliceDuration returns the slice duration info
func (n *Nii) getSliceDuration() float64 {
	return n.SliceDuration
}

// getSliceStart returns the slice start info
func (n *Nii) getSliceStart() int64 {
	return n.SliceStart
}

// getSliceEnd returns the slice end info
func (n *Nii) getSliceEnd() int64 {
	return n.SliceEnd
}

// getRawData returns the raw byte array of image
func (n *Nii) getRawData() []byte {
	return n.Volume
}

//----------------------------------------------------------------------------------------------------------------------
// Set methods
//----------------------------------------------------------------------------------------------------------------------

// setSliceCode sets the new slice code of the NIFTI image
func (n *Nii) setSliceCode(sliceCode int32) error {
	_, ok := NiiSliceAcquistionInfo[sliceCode]
	if ok {
		n.SliceCode = sliceCode
		return nil
	}
	return fmt.Errorf("unknown sliceCode %d", sliceCode)
}

// setQFormCode sets the new QForm code
func (n *Nii) setQFormCode(qFormCode int32) error {
	_, ok := NiiPatientOrientationInfo[qFormCode]
	if ok {
		n.QformCode = qFormCode
		return nil
	}
	return fmt.Errorf("unknown qFormCode %d", qFormCode)
}

// setSFormCode sets the new SForm code
func (n *Nii) setSFormCode(sFormCode int32) error {
	_, ok := NiiPatientOrientationInfo[n.SformCode]
	if ok {
		n.SformCode = sFormCode
		return nil
	}
	return fmt.Errorf("unknown sFormCode %d", sFormCode)
}

// setDatatype sets the new NIfTI datatype
func (n *Nii) setDatatype(datatype int32) error {
	_, ok := ValidDatatype[datatype]
	if ok {
		n.Datatype = datatype
		return nil
	}
	return fmt.Errorf("unknown datatype value %d", datatype)
}

// setAffine sets the new 4x4 affine matrix
func (n *Nii) setAffine(mat matrix.DMat44) {
	n.Affine = mat
}

// setDescrip returns the description with trailing null bytes removed
func (n *Nii) setDescrip(descrip string) error {

	if len([]byte(descrip)) > 79 {
		return errors.New("description must be fewer than 80 characters")
	}

	var bDescrip [80]byte
	copy(bDescrip[:], descrip)

	n.Descrip = bDescrip

	return nil
}

// setIntentName sets the new intent name
func (n *Nii) setIntentName(intentName string) error {

	if len([]byte(intentName)) > 15 {
		return errors.New("intent name must be fewer than 16 characters")
	}

	var bDescrip [80]byte
	copy(bDescrip[:], intentName)

	n.Descrip = bDescrip

	return nil
}

// setSliceDuration sets the new slice duration info
func (n *Nii) setSliceDuration(sliceDuration float64) {
	n.SliceDuration = sliceDuration
}

// setSliceStart sets the new slice start info
func (n *Nii) setSliceStart(sliceStart int64) {
	n.SliceStart = sliceStart
}

// setSliceEnd sets the new slice end info
func (n *Nii) setSliceEnd(sliceEnd int64) {
	n.SliceEnd = sliceEnd
}

// setXYZUnits sets the new spatial unit of measurements
func (n *Nii) setXYZUnits(xyzUnit int32) {
	n.XYZUnits = xyzUnit
}

// setTimeUnits sets the new temporal unit of measurements
func (n *Nii) setTimeUnits(timeUnit int32) {
	n.TimeUnits = timeUnit
}

// setAt sets the new value in bytes at (x, y, z, t) location
func (n *Nii) setAt(newVal float64, x, y, z, t int64) error {

	tIndex := t * n.Nx * n.Ny * n.Nz
	zIndex := n.Nx * n.Ny * z
	yIndex := n.Nx * y
	xIndex := x
	index := tIndex + zIndex + yIndex + xIndex
	nByPer := int64(n.NByPer)

	if index*nByPer > int64(len(n.Volume)) || (index+1)*nByPer > int64(len(n.Volume)) {
		return fmt.Errorf("index out of range. Max volume size is %d", len(n.Volume))
	}

	dataPoint := n.Volume[index*nByPer : (index+1)*nByPer]

	switch nByPer {
	case 0:
		return nil
	case 1:
		if len(dataPoint) > 0 {
			count := 0
			var buf bytes.Buffer
			err := binary.Write(&buf, n.ByteOrder, newVal)
			if err != nil {
				return err
			}
			for _, b := range buf.Bytes() {
				if b != 0x00 {
					count++
				}
			}
			if count == len(dataPoint) {
				copy(n.Volume[index*nByPer:(index+1)*nByPer], buf.Bytes())
			}
		}
	case 2: // This fits Uint16
		v := uint16(newVal)
		b := make([]byte, 2)

		switch n.ByteOrder {
		case binary.LittleEndian:
			binary.LittleEndian.PutUint16(b, v)
		case binary.BigEndian:
			binary.BigEndian.PutUint16(b, v)
		}
		copy(n.Volume[index*nByPer:(index+1)*nByPer], b)
	case 3:
		v := math.Float32bits(float32(newVal))
		b := make([]byte, 4)
		switch n.ByteOrder {
		case binary.LittleEndian:
			binary.LittleEndian.PutUint32(b, v)
		case binary.BigEndian:
			binary.BigEndian.PutUint32(b, v)
		}
		copy(n.Volume[index*nByPer:(index+1)*nByPer], b[:3])
	case 4: // This fits Uint32
		v := uint32(newVal)
		b := make([]byte, 4)
		switch n.ByteOrder {
		case binary.LittleEndian:
			binary.LittleEndian.PutUint32(b, v)
		case binary.BigEndian:
			binary.BigEndian.PutUint32(b, v)
		}
		copy(n.Volume[index*nByPer:(index+1)*nByPer], b)
	case 8:
		v := uint64(newVal)
		b := make([]byte, 8)
		switch n.ByteOrder {
		case binary.LittleEndian:
			binary.LittleEndian.PutUint64(b, v)
		case binary.BigEndian:
			binary.BigEndian.PutUint64(b, v)
		}
		copy(n.Volume[index*nByPer:(index+1)*nByPer], b)
	case 16: // Unsupported
	case 32: // Unsupported
	default:
	}
	return nil
}

// setVolume sets the new volume
func (n *Nii) setVolume(vol []byte) error {
	var bDataLength int64

	// Need at least nx, ny
	if n.Nx == 0 {
		return errors.New("x dimension must not be zero")
	}
	if n.Ny == 0 {
		return errors.New("y dimension must not be zero")
	}
	bDataLength = n.Nx * n.Ny

	if n.Nz > 0 {
		bDataLength = bDataLength * n.Nz
	}
	if n.Nt > 0 {
		bDataLength = bDataLength * n.Nt
	}
	if n.Nu > 0 {
		bDataLength = bDataLength * n.Nu
	}
	if n.Nv > 0 {
		bDataLength = bDataLength * n.Nv
	}
	if n.Nw > 0 {
		bDataLength = bDataLength * n.Nw
	}

	nByper, _ := assignDatatypeSize(n.Datatype)
	bDataLength = bDataLength * int64(nByper)

	if int64(len(vol)) != bDataLength {
		return fmt.Errorf("expected length of volume does not match. Expected %d Actual %d", bDataLength, len(vol))
	}

	n.Volume = vol
	return nil
}

// setVoxelToRawVolume converts the 1-D slice of float64 back to byte array
func (n *Nii) setVoxelToRawVolume(vox *Voxels) error {
	result := make([]byte, vox.GetRawByteSize(), vox.GetRawByteSize())
	nByPer := n.NByPer

	for index, voxel := range vox.voxel {
		switch nByPer {
		case 0:
			continue
		case 1: // 1 byte per voxel includes Uint8 and Int8
			var buf bytes.Buffer
			switch n.Datatype {
			case DT_UINT8:
				err := binary.Write(&buf, n.ByteOrder, uint8(voxel))
				if err != nil {
					return err
				}
			case DT_INT8:
				err := binary.Write(&buf, n.ByteOrder, int8(voxel))
				if err != nil {
					return err
				}
			}
			copy(result[index*int(nByPer):(index+1)*int(nByPer)], buf.Bytes())
		case 2: // This fits Uint16
			v := uint16(voxel)
			b := make([]byte, 2)
			switch n.ByteOrder {
			case binary.LittleEndian:
				binary.LittleEndian.PutUint16(b, v)
			case binary.BigEndian:
				binary.BigEndian.PutUint16(b, v)
			}
			copy(result[index*int(nByPer):(index+1)*int(nByPer)], b)
		case 3: // This fits Uint32 -> RGB24
			v := math.Float32bits(float32(voxel))
			b := make([]byte, 4)
			switch n.ByteOrder {
			case binary.LittleEndian:
				binary.LittleEndian.PutUint32(b, v)
			case binary.BigEndian:
				binary.BigEndian.PutUint32(b, v)
			}
			copy(result[index*int(nByPer):(index+1)*int(nByPer)], b[:3])
		case 4: // This fits Uint32
			v := uint32(voxel)
			b := make([]byte, 4)
			switch n.ByteOrder {
			case binary.LittleEndian:
				binary.LittleEndian.PutUint32(b, v)
			case binary.BigEndian:
				binary.BigEndian.PutUint32(b, v)
			}
			copy(result[index*int(nByPer):(index+1)*int(nByPer)], b)
		case 8:
			v := uint64(voxel)
			b := make([]byte, 8)
			switch n.ByteOrder {
			case binary.LittleEndian:
				binary.LittleEndian.PutUint64(b, v)
			case binary.BigEndian:
				binary.BigEndian.PutUint64(b, v)
			}
			copy(result[index*int(nByPer):(index+1)*int(nByPer)], b)
		case 16: // Unsupported
		case 32: // Unsupported
		default:
		}
	}
	n.Volume = result
	return nil
}
