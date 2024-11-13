package nifti

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/okieraised/gonii/pkg/matrix"
)

// Nii defines the structure of the NIFTI-1 data for I/O purpose
type Nii struct {
	NDim          int64            `json:"ndim"`           // last dimension greater than 1 (1..7)
	Nx            int64            `json:"Nx"`             // dimensions of grid array
	Ny            int64            `json:"Ny"`             // dimensions of grid array
	Nz            int64            `json:"nz"`             // dimensions of grid array
	Nt            int64            `json:"nt"`             // dimensions of grid array
	Nu            int64            `json:"nu"`             // dimensions of grid array
	Nv            int64            `json:"nv"`             // dimensions of grid array
	Nw            int64            `json:"nw"`             // dimensions of grid array
	Dim           [8]int64         `json:"dim"`            // dim[0] = ndim, dim[1] = Nx, etc
	NVox          int64            `json:"nvox"`           // number of voxels = Nx*Ny*nz*...*nw
	NByPer        int32            `json:"nbyper"`         // bytes per voxel, matches datatype (Datatype)
	Datatype      int32            `json:"datatype"`       // type of data in voxels: DT_* code
	Dx            float64          `json:"dx"`             // grid spacings
	Dy            float64          `json:"dy"`             // grid spacings
	Dz            float64          `json:"dz"`             // grid spacings
	Dt            float64          `json:"dt"`             // grid spacings
	Du            float64          `json:"du"`             // grid spacings
	Dv            float64          `json:"dv"`             // grid spacings
	Dw            float64          `json:"dw"`             // grid spacings
	PixDim        [8]float64       `json:"pix_dim"`        // pixdim[1]=dx, etc
	SclSlope      float64          `json:"scl_slope"`      // scaling parameter: slope
	SclInter      float64          `json:"scl_inter"`      // scaling parameter: intercept
	CalMin        float64          `json:"cal_min"`        // calibration parameter: minimum
	CalMax        float64          `json:"cal_max"`        // calibration parameter: maximum
	QformCode     int32            `json:"qform_code"`     // codes for (x,y,z) space meaning
	SformCode     int32            `json:"sform_code"`     // codes for (x,y,z) space meaning
	FreqDim       int32            `json:"freq_dim"`       // indices (1,2,3, or 0) for MRI
	PhaseDim      int32            `json:"phase_dim"`      // directions in dim[]/pixdim[]
	SliceDim      int32            `json:"slice_dim"`      // directions in dim[]/pixdim[]
	SliceCode     int32            `json:"slice_code"`     // code for slice timing pattern
	SliceStart    int64            `json:"slice_start"`    // index for start of slices
	SliceEnd      int64            `json:"slice_end"`      // index for end of slices
	SliceDuration float64          `json:"slice_duration"` // time between individual slices
	QuaternB      float64          `json:"quatern_b"`      // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QuaternC      float64          `json:"quatern_c"`      // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QuaternD      float64          `json:"quatern_d"`      // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QoffsetX      float64          `json:"qoffset_x"`      // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QoffsetY      float64          `json:"qoffset_y"`      // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QoffsetZ      float64          `json:"qoffset_z"`      // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QFac          float64          `json:"q_fac"`          // Quaternion transform parameters [when writing a dataset, these are used for qform, NOT qto_xyz]
	QtoXYZ        matrix.DMat44    `json:"qto_xyz"`        // qform: transform (i,j,k) to (x,y,z)
	QtoIJK        matrix.DMat44    `json:"qto_ijk"`        // qform: transform (x,y,z) to (i,j,k)
	StoXYZ        matrix.DMat44    `json:"sto_xyz"`        // sform: transform (i,j,k) to (x,y,z)
	StoIJK        matrix.DMat44    `json:"sto_ijk"`        // sform: transform (x,y,z) to (i,j,k)
	TOffset       float64          `json:"t_offset"`       // time coordinate offset
	XYZUnits      int32            `json:"xyz_units"`      // dx,dy,dz units: NIFTI_UNITS_* code
	TimeUnits     int32            `json:"time_units"`     // dt units: NIFTI_UNITS_* code
	NiftiType     int32            `json:"nifti_type"`     // 0==Analyze, 1==NIFTI-1 (file), 2==NIFTI-1 (2 files), 3==NIFTI-ASCII (1 file)
	IntentCode    int32            `json:"intent_code"`    // statistic type (or something)
	IntentP1      float64          `json:"intent_p1"`      // intent parameters
	IntentP2      float64          `json:"intent_p2"`      // intent parameters
	IntentP3      float64          `json:"intent_p3"`      // intent parameters
	IntentName    [16]byte         `json:"intent_name"`    // optional description of intent data
	Descrip       [80]byte         `json:"descrip"`        // optional text to describe dataset
	AuxFile       [24]byte         `json:"aux_file"`       // auxiliary filename
	FName         *byte            `json:"f_name"`         // header filename
	IName         *byte            `json:"i_name"`         // image filename
	INameOffset   int32            `json:"i_name_offset"`  // offset into IName where data start
	SwapSize      int32            `json:"swap_size"`      // swap unit in image data (might be 0)
	ByteOrder     binary.ByteOrder `json:"byte_order"`     // byte order on disk (MSB_ or LSB_FIRST)
	Volume        []byte           `json:"volume"`         // slice of data: nbyper*nvox bytes
	NumExt        int32            `json:"num_ext"`        // number of extensions in extList
	Nifti1Ext     []Nifti1Ext      `json:"nifti1_ext"`     // array of extension structs (with data)
	IJKOrient     [3]int32         `json:"ijk_orient"`     // self-add. Orientation ini, j, k coordinate
	Affine        matrix.DMat44    `json:"affine"`         // self-add. Affine matrix
	VoxOffset     float64          `json:"vox_offset"`     // self-add. Voxel offset
	Version       int              `json:"version"`        // self-add. Used for version identification when writing
}

// Nifti1Ext defines the NIfTI-1 extension
type Nifti1Ext struct {
	ECode int32  `json:"e_code"`
	EData []byte `json:"e_data"`
	ESize int32  `json:"e_size"`
}

//----------------------------------------------------------------------------------------------------------------------
// Get methods
//----------------------------------------------------------------------------------------------------------------------

// GetQuaternB returns the QuaternB parameter
func (n *Nii) GetQuaternB() float64 {
	return n.QuaternB
}

// GetQuaternC returns the QuaternC parameter
func (n *Nii) GetQuaternC() float64 {
	return n.QuaternC
}

// GetQuaternD returns the QuaternD parameter
func (n *Nii) GetQuaternD() float64 {
	return n.QuaternD
}

// GetQoffsetX returns the QoffsetX parameter
func (n *Nii) GetQoffsetX() float64 {
	return n.QoffsetX
}

// GetQoffsetY returns the QoffsetY parameter
func (n *Nii) GetQoffsetY() float64 {
	return n.QoffsetY
}

// GetQoffsetZ returns the QoffsetZ parameter
func (n *Nii) GetQoffsetZ() float64 {
	return n.QoffsetZ
}

// GetQtoXYZMat returns the QtoXYZ matrix as [4][4]float64
func (n *Nii) GetQtoXYZMat() matrix.DMat44 {
	return n.QtoXYZ
}

// GetQtoIJKMat returns the QtoIJK matrix as [4][4]float64
func (n *Nii) GetQtoIJKMat() matrix.DMat44 {
	return n.QtoIJK
}

// GetStoXYZMat returns the StoXYZ matrix as [4][4]float64
func (n *Nii) GetStoXYZMat() matrix.DMat44 {
	return n.StoXYZ
}

// GetStoIJKMat returns the StoIJK matrix as [4][4]float64
func (n *Nii) GetStoIJKMat() matrix.DMat44 {
	return n.StoIJK
}

// GetSliceCode returns the slice code of the NIFTI image
func (n *Nii) GetSliceCode() string {
	return getSliceCode(n.SliceCode)
}

// GetQFormCode returns the QformCode code parameter
func (n *Nii) GetQFormCode() string {
	qForm, ok := NiiPatientOrientationInfo[n.QformCode]
	if !ok {
		return INVALID
	}
	return qForm
}

// GetSFormCode returns the SformCode parameter
func (n *Nii) GetSFormCode() string {
	sForm, ok := NiiPatientOrientationInfo[n.SformCode]
	if !ok {
		return INVALID
	}
	return sForm
}

// GetDatatype returns the corresponding NIfTI datatype
func (n *Nii) GetDatatype() string {
	return getDatatype(n.Datatype)
}

// GetOrientation returns the image orientation
func (n *Nii) GetOrientation() [3]string {
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

// GetVoxels returns the 1-D slice of voxel values of type float64
func (n *Nii) GetVoxels() *Voxels {
	vox := NewVoxels(n.Nx, n.Ny, n.Nz, n.Nt, n.Datatype)
	for x := int64(0); x < n.Nx; x++ {
		for y := int64(0); y < n.Ny; y++ {
			for z := int64(0); z < n.Nz; z++ {
				for t := int64(0); t < n.Nt; t++ {
					vox.Set(x, y, z, t, n.GetAt(x, y, z, t))
				}
			}
		}
	}
	return vox
}

// GetAt returns the value at (x, y, z, t) location
func (n *Nii) GetAt(x, y, z, t int64) float64 {
	tIndex := t * n.Nx * n.Ny * n.Nz
	zIndex := n.Nx * n.Ny * z
	yIndex := n.Nx * y
	xIndex := x
	index := tIndex + zIndex + yIndex + xIndex
	nByPer := int64(n.NByPer)

	dataPoint := n.Volume[index*nByPer : (index+1)*nByPer]

	var value float64
	switch n.NByPer {
	case 0:
	case 1:
		value = float64(dataPoint[0])
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
	case 3: // This fits Uint32
		var v uint32
		switch n.ByteOrder {
		case binary.LittleEndian:
			v = uint32(dataPoint[0]) | uint32(dataPoint[1])<<8 | uint32(dataPoint[2])<<16
		case binary.BigEndian:
			v = uint32(dataPoint[2]) | uint32(dataPoint[1])<<8 | uint32(dataPoint[0])<<16
		}
		value = float64(math.Float32frombits(v))
	case 4: // This fits Uint32
		var v uint32
		switch n.ByteOrder {
		case binary.LittleEndian:
			v = binary.LittleEndian.Uint32(dataPoint)
		case binary.BigEndian:
			v = binary.BigEndian.Uint32(dataPoint)
		}
		value = uint32ToFloat64(v, n.Datatype)
	case 8: // THis fits Uint64
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

// GetTimeSeries returns the time-series of a point
func (n *Nii) GetTimeSeries(x, y, z int64) ([]float64, error) {
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
		timeSeries = append(timeSeries, n.GetAt(x, y, z, int64(t)))
	}
	return timeSeries, nil
}

// GetSlice returns the image in x-y dimension
func (n *Nii) GetSlice(z, t int64) ([][]float64, error) {
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
			slice[x][y] = n.GetAt(int64(x), int64(y), z, t)
		}
	}
	return slice, nil
}

// GetVolume return the whole image volume at time t
func (n *Nii) GetVolume(t int64) ([][][]float64, error) {
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
				volume[x][y][z] = n.GetAt(int64(x), int64(y), int64(z), t)
			}
		}
	}
	return volume, nil
}

// GetUnitsOfMeasurements returns the spatial and temporal units of measurements
func (n *Nii) GetUnitsOfMeasurements() ([2]string, error) {
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

// GetAffine returns the 4x4 affine matrix
func (n *Nii) GetAffine() matrix.DMat44 {
	return n.Affine
}

// GetImgShape returns the image shape in terms of x, y, z, t
func (n *Nii) GetImgShape() [4]int64 {
	dim := [4]int64{}

	for index := range dim {
		dim[index] = n.Dim[index+1]
	}
	return dim
}

// GetVoxelSize returns the voxel size of the image
func (n *Nii) GetVoxelSize() [4]float64 {
	size := [4]float64{}
	for index := range size {
		size[index] = n.PixDim[index+1]
	}
	return size
}

// GetDescrip returns the description with trailing null bytes removed
func (n *Nii) GetDescrip() string {
	return strings.ReplaceAll(string(n.Descrip[:]), "\x00", "")
}

// GetIntentName returns the intent name with trailing null bytes removed
func (n *Nii) GetIntentName() string {
	return strings.ReplaceAll(string(n.IntentName[:]), "\x00", "")
}

// GetAuxFile returns the AuxFile with trailing null bytes removed
func (n *Nii) GetAuxFile() string {
	return strings.ReplaceAll(string(n.AuxFile[:]), "\x00", "")
}

// GetSliceDuration returns the slice duration info
func (n *Nii) GetSliceDuration() float64 {
	return n.SliceDuration
}

// GetSliceStart returns the slice start info
func (n *Nii) GetSliceStart() int64 {
	return n.SliceStart
}

// GetSliceEnd returns the slice end info
func (n *Nii) GetSliceEnd() int64 {
	return n.SliceEnd
}

// GetRawData returns the raw byte array of image
func (n *Nii) GetRawData() []byte {
	return n.Volume
}

// GetSclSlope returns the SclSlope parameter
func (n *Nii) GetSclSlope() float64 {
	return n.SclSlope
}

// GetSclInter returns the SclInter parameter
func (n *Nii) GetSclInter() float64 {
	return n.SclInter
}

// GetPixDim returns the PixDim parameter
func (n *Nii) GetPixDim() [8]float64 {
	return n.PixDim
}

// GetDim returns the Dim parameter
func (n *Nii) GetDim() [8]int64 {
	return n.Dim
}

// GetNVox returns the NVox parameter
func (n *Nii) GetNVox() int64 {
	return n.NVox
}

// GetQFac returns the QFac parameters
func (n *Nii) GetQFac() float64 {
	return n.QFac
}

// GetTOffset returns the TOffset parameters
func (n *Nii) GetTOffset() float64 {
	return n.TOffset
}

// GetXYZUnits returns the XYZUnits parameters
func (n *Nii) GetXYZUnits() int32 {
	return n.XYZUnits
}

// GetTimeUnits returns the TimeUnits parameters
func (n *Nii) GetTimeUnits() int32 {
	return n.TimeUnits
}

// GetNiftiType returns the NiftiType parameters
func (n *Nii) GetNiftiType() int32 {
	return n.NiftiType
}

// GetIntentCode returns the IntentCode parameters
func (n *Nii) GetIntentCode() int32 {
	return n.IntentCode
}

// GetIntentP1 returns the IntentP1 parameters
func (n *Nii) GetIntentP1() float64 {
	return n.IntentP1
}

// GetIntentP2 returns the IntentP2 parameters
func (n *Nii) GetIntentP2() float64 {
	return n.IntentP2
}

// GetIntentP3 returns the IntentP3 parameters
func (n *Nii) GetIntentP3() float64 {
	return n.IntentP3
}

// GetFreqDim returns the FreqDim parameters
func (n *Nii) GetFreqDim() int32 {
	return n.FreqDim
}

// GetPhaseDim returns the PhaseDim parameters
func (n *Nii) GetPhaseDim() int32 {
	return n.PhaseDim
}

// GetSliceDim returns the SliceDim parameters
func (n *Nii) GetSliceDim() int32 {
	return n.SliceDim
}

//----------------------------------------------------------------------------------------------------------------------
// Set methods
//----------------------------------------------------------------------------------------------------------------------

// SetSliceCode sets the new slice code of the NIFTI image
func (n *Nii) SetSliceCode(sliceCode int32) error {
	_, ok := NiiSliceAcquistionInfo[sliceCode]
	if ok {
		n.SliceCode = sliceCode
		return nil
	}
	return fmt.Errorf("unknown sliceCode %d", sliceCode)
}

// SetQFormCode sets the new QForm code
func (n *Nii) SetQFormCode(qFormCode int32) error {
	_, ok := NiiPatientOrientationInfo[qFormCode]
	if ok {
		n.QformCode = qFormCode
		return nil
	}
	return fmt.Errorf("unknown qFormCode %d", qFormCode)
}

// SetSFormCode sets the new SForm code
func (n *Nii) SetSFormCode(sFormCode int32) error {
	_, ok := NiiPatientOrientationInfo[n.SformCode]
	if ok {
		n.SformCode = sFormCode
		return nil
	}
	return fmt.Errorf("unknown sFormCode %d", sFormCode)
}

// SetDatatype sets the new NIfTI datatype
func (n *Nii) SetDatatype(datatype int32) error {
	_, ok := ValidDatatype[datatype]
	if ok {
		n.Datatype = datatype
		return nil
	}
	return fmt.Errorf("unknown datatype value %d", datatype)
}

// SetAffine sets the new 4x4 affine matrix
func (n *Nii) SetAffine(mat matrix.DMat44) {
	n.Affine = mat
}

// SetDescrip returns the description with trailing null bytes removed
func (n *Nii) SetDescrip(descrip string) error {

	if len([]byte(descrip)) > 79 {
		return errors.New("description must be fewer than 80 characters")
	}

	var bDescrip [80]byte
	copy(bDescrip[:], descrip)

	n.Descrip = bDescrip

	return nil
}

// SetIntentName sets the new intent name
func (n *Nii) SetIntentName(intentName string) error {

	if len([]byte(intentName)) > 15 {
		return errors.New("intent name must be fewer than 16 characters")
	}

	var bIntentName [16]byte
	copy(bIntentName[:], intentName)

	n.IntentName = bIntentName

	return nil
}

// SetAuxFile sets the new AuxFile
func (n *Nii) SetAuxFile(auxFile string) error {

	if len([]byte(auxFile)) > 24 {
		return errors.New("AuxFile must be fewer than 24 characters")
	}

	var bAuxFile [24]byte
	copy(bAuxFile[:], auxFile)

	n.AuxFile = bAuxFile

	return nil
}

// SetSliceDuration sets the new slice duration info
func (n *Nii) SetSliceDuration(sliceDuration float64) {
	n.SliceDuration = sliceDuration
}

// SetSliceStart sets the new slice start info
func (n *Nii) SetSliceStart(sliceStart int64) {
	n.SliceStart = sliceStart
}

// SetSliceEnd sets the new slice end info
func (n *Nii) SetSliceEnd(sliceEnd int64) {
	n.SliceEnd = sliceEnd
}

// SetXYZUnits sets the new spatial unit of measurements
func (n *Nii) SetXYZUnits(xyzUnit int32) {
	n.XYZUnits = xyzUnit
}

// SetTimeUnits sets the new temporal unit of measurements
func (n *Nii) SetTimeUnits(timeUnit int32) {
	n.TimeUnits = timeUnit
}

// SetSclSlope sets the SclSlope parameter
func (n *Nii) SetSclSlope(sclSlope float64) {
	n.SclSlope = sclSlope
}

// SetSclInter sets the SclInter parameter
func (n *Nii) SetSclInter(sclInter float64) {
	n.SclInter = sclInter
}

// SetPixDim sets the PixDim parameter
func (n *Nii) SetPixDim(pixDim [8]float64) {
	n.PixDim = pixDim
}

// SetDim sets the Dim parameter
func (n *Nii) SetDim(dim [8]int64) {
	n.Dim = dim
}

// SetNVox sets the NVox parameter
func (n *Nii) SetNVox(nVox int64) {
	n.NVox = nVox
}

// SetQFac sets the QFac parameters
func (n *Nii) SetQFac(qFac float64) {
	n.QFac = qFac
}

// SetTOffset sets the TOffset parameters
func (n *Nii) SetTOffset(tOffset float64) {
	n.TOffset = tOffset
}

// SetIntentCode sets the IntentCode parameters
func (n *Nii) SetIntentCode(intentCode int32) {
	n.IntentCode = intentCode
}

// SetIntentP1 sets the IntentP1 parameters
func (n *Nii) SetIntentP1(intentP1 float64) {
	n.IntentP1 = intentP1
}

// SetIntentP2 sets the IntentP2 parameters
func (n *Nii) SetIntentP2(intentP2 float64) {
	n.IntentP2 = intentP2
}

// SetIntentP3 sets the IntentP3 parameters
func (n *Nii) SetIntentP3(intentP3 float64) {
	n.IntentP3 = intentP3
}

// SetFreqDim sets the FreqDim parameters
func (n *Nii) SetFreqDim(freqDim int32) {
	n.FreqDim = freqDim
}

// SetPhaseDim sets the PhaseDim parameters
func (n *Nii) SetPhaseDim(phaseDim int32) {
	n.PhaseDim = phaseDim
}

// SetSliceDim sets the SliceDim parameters
func (n *Nii) SetSliceDim(sliceDim int32) {
	n.SliceDim = sliceDim
}

// SetVolume sets the new volume
func (n *Nii) SetVolume(vol []byte) error {
	var bDataLength int64

	// Need at least Nx, Ny
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

	nByper, _ := AssignDatatypeSize(n.Datatype)
	bDataLength = bDataLength * int64(nByper)

	if int64(len(vol)) != bDataLength {
		return fmt.Errorf("expected length of volume does not match. Expected %d Actual %d", bDataLength, len(vol))
	}

	n.Volume = vol
	return nil
}

// SetAt sets the new value in bytes at (x, y, z, t) location
func (n *Nii) SetAt(newVal float64, x, y, z, t int64) error {

	tIndex := t * n.Nx * n.Ny * n.Nz
	zIndex := n.Nx * n.Ny * z
	yIndex := n.Nx * y
	xIndex := x
	index := tIndex + zIndex + yIndex + xIndex
	nByPer := int64(n.NByPer)

	if index*nByPer > int64(len(n.Volume)) || (index+1)*nByPer > int64(len(n.Volume)) {
		return fmt.Errorf("index out of range. Max volume size is %d", len(n.Volume))
	}
	bVal, err := ConvertVoxelToBytes(newVal, n.SclSlope, n.SclInter, n.Datatype, n.ByteOrder, n.NByPer)
	if err != nil {
		return err
	}
	copy(n.Volume[index*nByPer:(index+1)*nByPer], bVal)
	return nil
}

// SetVoxelToRawVolume converts the 1-D slice of float64 back to byte array
func (n *Nii) SetVoxelToRawVolume(vox *Voxels) error {
	result := make([]byte, vox.GetRawByteSize())
	nByPer := n.NByPer

	for index, voxel := range vox.voxel {
		bVal, err := ConvertVoxelToBytes(voxel, n.SclSlope, n.SclInter, n.Datatype, n.ByteOrder, nByPer)
		if err != nil {
			return err
		}
		copy(result[index*int(nByPer):(index+1)*int(nByPer)], bVal)
	}
	n.Volume = result
	return nil
}
