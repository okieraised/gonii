# gonii
Standalone, pure golang NIfTI file parser that provide functionalities to read and write NIfTI file format (including NIfTI-1 and NIfTI-2)

** This package is under active development and can be in a broken state. Please use the latest released version **

To install this package, run:
```shell
go get github.com/okieraised/gonii
```

To parse a single NIfTI file:
```go
package main

import (
	"fmt"
	"github.com/okieraised/gonii"
)

func main() {
	filePath := "./test_data/int16.nii.gz"

	// Init new reader with option to keep the header structure after parsing
	rd, err := gonii.NewNiiReader(filePath, gonii.WithRetainHeader(true))
	if err != nil {
		panic(err)
	}

	// Parse the image
	err = rd.Parse()
	if err != nil {
		panic(err)
	}

	// Access the NIfTI image structure
	img := rd.GetNiiData()

	// to see the raw byte data
	fmt.Println(img.Volume[60000:70000])

	// Transform the raw byte slices to a 1-D slice of voxel value in float64
	voxels := rd.GetNiiData().GetVoxels()
	// to see the transformed value as a float64 slice
	fmt.Println(voxels.GetDataset()[60000:70000])
}
```


## TODO
- [ ] Improve NIfTI reader parsing speed for large file size
- [X] Add support for NIfTI writer to export data as NIfTI-2 format
