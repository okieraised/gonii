package gonii

import (
	"bytes"
	"encoding/binary"
	"github.com/okieraised/gonii/internal/utils"
	"github.com/okieraised/gonii/pkg/nifti"
	"net/http"
	"os"
)

//----------------------------------------------------------------------------------------------------------------------
// Define Reader methods
//----------------------------------------------------------------------------------------------------------------------

// NewNiiReader returns a new NIfTI reader
//
// Parameters:
//     - `filePath`         : Input NIfTI file. In case of separate .hdr/.img file, this is the image file (.img)
func NewNiiReader(filePath string, options ...func(*nifti.NiiReader) error) (nifti.Reader, error) {
	// Init new reader
	reader := new(nifti.NiiReader)
	reader.SetBinaryOrder(binary.LittleEndian)
	reader.SetDataset(&nifti.Nii{})

	for _, opt := range options {
		err := opt(reader)
		if err != nil {
			return nil, err
		}
	}

	// This is inefficient since it read the whole file to the memory
	// TODO: improve this for large file
	bData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Check the content type to see if the file is gzipped. Do not depend on just the extensions of the file
	bData, err = deflateFileContent(bData)
	reader.SetReader(bytes.NewReader(bData))
	return reader, nil
}

// WithInMemory allows option to read the whole file into memory. The default is true.
// This is for future implementation. Currently, all file is read into memory before parsing
func WithInMemory(inMemory bool) func(*nifti.NiiReader) error {
	return func(w *nifti.NiiReader) error {
		w.SetInMemory(inMemory)
		return nil
	}
}

// WithRetainHeader allows option to keep the header after parsing instead of just keeping the NIfTI data structure
func WithRetainHeader(retainHeader bool) func(*nifti.NiiReader) error {
	return func(w *nifti.NiiReader) error {
		w.SetRetainHeader(retainHeader)
		return nil
	}
}

// WithHeaderFile allows option to specify the separate header file in case of NIfTI pair .hdr/.img
func WithHeaderFile(headerFile string) func(*nifti.NiiReader) error {
	return func(w *nifti.NiiReader) error {
		bData, err := os.ReadFile(headerFile)
		if err != nil {
			return err
		}
		// Check the content type to see if the file is gzipped. Do not depend on just the extensions of the file
		bData, err = deflateFileContent(bData)
		w.SetHdrReader(bytes.NewReader(bData))
		return nil
	}
}

//----------------------------------------------------------------------------------------------------------------------
// Define Writer methods
//----------------------------------------------------------------------------------------------------------------------

// NewNiiWriter returns a new NIfTI writer. If no version is specified, the writer will default to write to NIfTI version 1
func NewNiiWriter(filePath string, options ...func(*nifti.NiiWriter)) (nifti.Writer, error) {
	writer := new(nifti.NiiWriter)

	writer.SetFilePath(filePath)
	writer.SetWriteHeaderFile(false)     // Default to false. Write to a single file only
	writer.SetCompression(false)         // Default to false. No compression
	writer.SetVersion(nifti.NIIVersion1) // Default to version 1

	// Other options
	for _, opt := range options {
		opt(writer)
	}
	return writer, nil
}

// WithWriteHeaderFile sets the option to write NIfTI image to a header/image (.hdr/.img) file pair
//
// If true, output will be two files for the header and the image. Default is false.
func WithWriteHeaderFile(writeHeaderFile bool) func(*nifti.NiiWriter) {
	return func(w *nifti.NiiWriter) {
		w.SetWriteHeaderFile(writeHeaderFile)
	}
}

// WithCompression sets the option to write compressed NIfTI image to a single file (.nii.gz)
//
// If true, the whole file will be compressed. Default is false.
func WithCompression(withCompression bool) func(writer *nifti.NiiWriter) {
	return func(w *nifti.NiiWriter) {
		w.SetCompression(withCompression)
	}
}

// WithHeader sets the option to allow user to provide predefined NIfTI-1 header structure.
//
// If no header provided, the header will be converted from the NIfTI image structure
func WithHeader(header *nifti.Nii1Header) func(*nifti.NiiWriter) {
	return func(w *nifti.NiiWriter) {
		w.SetHeader(header)
	}
}

// WithNIfTIData sets the option to allow user to provide predefined NIfTI-1 data structure.
func WithNIfTIData(data *nifti.Nii) func(writer *nifti.NiiWriter) {
	return func(w *nifti.NiiWriter) {
		w.SetNiiData(data)
	}
}

// WithVersion sets the option to specify the exported NIfTI version (NIfTI-1 or 2). Default is NIfTI-1
func WithVersion(version int) func(writer *nifti.NiiWriter) {
	return func(w *nifti.NiiWriter) {
		w.SetVersion(version)
	}
}

//----------------------------------------------------------------------------------------------------------------------
// Define Support function
//----------------------------------------------------------------------------------------------------------------------

// deflateFileContent deflates the gzipped binary to its original content
func deflateFileContent(bData []byte) ([]byte, error) {
	var err error
	mimeType := http.DetectContentType(bData[:512])
	if mimeType == "application/x-gzip" {
		bData, err = utils.DeflateGzip(bData)
		if err != nil {
			return nil, err
		}
	}
	return bData, nil
}
