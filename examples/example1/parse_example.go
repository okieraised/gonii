package main

import (
	"fmt"
	"github.com/okieraised/gonii"
)

func main() {
	filePath := "./test_data/int16.nii.gz"

	// Init new reader with option to keep the header structure after parsing
	rd, err := gonii.NewNiiReader(gonii.WithImageFile(filePath), gonii.WithRetainHeader(true))
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
