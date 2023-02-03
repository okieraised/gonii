package nifti

// Nii1Header defines the structure of the NIFTI-1 header
type Nii1Header struct {
	SizeofHdr      int32      `json:"sizeof_hdr"`
	DataTypeUnused [10]uint8  `json:"data_type"`
	DbName         [18]uint8  `json:"db_name"`
	Extents        int32      `json:"extents"`
	SessionError   int16      `json:"session_error"`
	Regular        uint8      `json:"regular"`
	DimInfo        uint8      `json:"dim_info"`
	Dim            [8]int16   `json:"dim"`
	IntentP1       float32    `json:"intent_p1"`
	IntentP2       float32    `json:"intent_p2"`
	IntentP3       float32    `json:"intent_p3"`
	IntentCode     int16      `json:"intent_code"`
	Datatype       int16      `json:"datatype"`
	Bitpix         int16      `json:"bitpix"`
	SliceStart     int16      `json:"slice_start"`
	Pixdim         [8]float32 `json:"pixdim"`
	VoxOffset      float32    `json:"vox_offset"`
	SclSlope       float32    `json:"scl_slope"`
	SclInter       float32    `json:"scl_inter"`
	SliceEnd       int16      `json:"slice_end"`
	SliceCode      uint8      `json:"slice_code"`
	XyztUnits      uint8      `json:"xyzt_units"`
	CalMax         float32    `json:"cal_max"`
	CalMin         float32    `json:"cal_min"`
	SliceDuration  float32    `json:"slice_duration"`
	Toffset        float32    `json:"toffset"`
	Glmax          int32      `json:"glmax"`
	Glmin          int32      `json:"glmin"`
	Descrip        [80]uint8  `json:"descrip"`
	AuxFile        [24]uint8  `json:"aux_file"`
	QformCode      int16      `json:"qform_code"`
	SformCode      int16      `json:"sform_code"`
	QuaternB       float32    `json:"quatern_b"`
	QuaternC       float32    `json:"quatern_c"`
	QuaternD       float32    `json:"quatern_d"`
	QoffsetX       float32    `json:"qoffset_x"`
	QoffsetY       float32    `json:"qoffset_y"`
	QoffsetZ       float32    `json:"qoffset_z"`
	SrowX          [4]float32 `json:"srow_x"`
	SrowY          [4]float32 `json:"srow_y"`
	SrowZ          [4]float32 `json:"srow_z"`
	IntentName     [16]uint8  `json:"intent_name"`
	Magic          [4]uint8   `json:"magic"`
}

// Nii2Header defines the structure of the NIFTI-2 header
type Nii2Header struct {
	SizeofHdr     int32      `json:"sizeof_hdr"`
	Magic         [8]uint8   `json:"magic"`
	Datatype      int16      `json:"datatype"`
	Bitpix        int16      `json:"bitpix"`
	Dim           [8]int64   `json:"dim"`
	IntentP1      float64    `json:"intent_p1"`
	IntentP2      float64    `json:"intent_p2"`
	IntentP3      float64    `json:"intent_p3"`
	Pixdim        [8]float64 `json:"pixdim"`
	VoxOffset     int64      `json:"vox_offset"`
	SclSlope      float64    `json:"scl_slope"`
	SclInter      float64    `json:"scl_inter"`
	CalMax        float64    `json:"cal_max"`
	CalMin        float64    `json:"cal_min"`
	SliceDuration float64    `json:"slice_duration"`
	Toffset       float64    `json:"toffset"`
	SliceStart    int64      `json:"slice_start"`
	SliceEnd      int64      `json:"slice_end"`
	Descrip       [80]uint8  `json:"descrip"`
	AuxFile       [24]uint8  `json:"aux_file"`
	QformCode     int32      `json:"qform_code"`
	SformCode     int32      `json:"sform_code"`
	QuaternB      float64    `json:"quatern_b"`
	QuaternC      float64    `json:"quatern_c"`
	QuaternD      float64    `json:"quatern_d"`
	QoffsetX      float64    `json:"qoffset_x"`
	QoffsetY      float64    `json:"qoffset_y"`
	QoffsetZ      float64    `json:"qoffset_z"`
	SrowX         [4]float64 `json:"srow_x"`
	SrowY         [4]float64 `json:"srow_y"`
	SrowZ         [4]float64 `json:"srow_z"`
	SliceCode     int32      `json:"slice_code"`
	XyztUnits     int32      `json:"xyzt_units"`
	IntentCode    int32      `json:"intent_code"`
	IntentName    [16]uint8  `json:"intent_name"`
	DimInfo       uint8      `json:"dim_info"`
	UnusedStr     [15]uint8  `json:"unused_str"`
}
