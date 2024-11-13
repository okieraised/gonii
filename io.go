package gonii

import (
	"bytes"
	"encoding/binary"
	"net/http"
	"os"

	"github.com/okieraised/gonii/internal/utils"
	"github.com/okieraised/gonii/pkg/nifti"
)

//----------------------------------------------------------------------------------------------------------------------
// Define Reader methods
//----------------------------------------------------------------------------------------------------------------------

// NewNiiReader returns a new NIfTI reader
//
// Options:
//   - `WithReadInMemory(inMemory bool)`         : Read the whole file into memory
//   - `WithReadRetainHeader(retainHeader bool)` : Whether to retain the header structure after parsing
//   - `WithReadHeaderFile(headerFile string)`   : Specify a header file path in case of separate .hdr/.img file
//   - `WithReadImageFile(niiFile string)`       : Specify an image file path
//   - `WithReadImageReader(r *bytes.Reader)`    : Specify a header file reader in case of separate .hdr/.img file
//   - `WithReadHeaderReader(r *bytes.Reader)`   : Specify an image file reader
func NewNiiReader(options ...func(*nifti.NiiReader) error) (nifti.Reader, error) {
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
	return reader, nil
}

// WithReadInMemory allows option to read the whole file into memory. The default is true.
// This is for future implementation. Currently, all file is read into memory before parsing
func WithReadInMemory(inMemory bool) func(*nifti.NiiReader) error {
	return func(w *nifti.NiiReader) error {
		w.SetInMemory(inMemory)
		return nil
	}
}

// WithReadRetainHeader allows option to keep the header after parsing instead of just keeping the NIfTI data structure
func WithReadRetainHeader(retainHeader bool) func(*nifti.NiiReader) error {
	return func(w *nifti.NiiReader) error {
		w.SetRetainHeader(retainHeader)
		return nil
	}
}

// WithReadHeaderFile allows option to specify the separate header file in case of NIfTI pair .hdr/.img
func WithReadHeaderFile(headerFile string) func(*nifti.NiiReader) error {
	return func(w *nifti.NiiReader) error {
		bData, err := os.ReadFile(headerFile)
		if err != nil {
			return err
		}
		// Check the content type to see if the file is gzipped. Do not depend on just the extensions of the file
		bData, err = deflateFileContent(bData)
		if err != nil {
			return err
		}
		w.SetHdrReader(bytes.NewReader(bData))
		return nil
	}
}

// WithReadImageFile allows option to specify the NIfTI file (.nii.gz or .nii)
func WithReadImageFile(niiFile string) func(*nifti.NiiReader) error {
	return func(w *nifti.NiiReader) error {
		bData, err := os.ReadFile(niiFile)
		if err != nil {
			return err
		}
		// Check the content type to see if the file is gzipped. Do not depend on just the extensions of the file
		bData, err = deflateFileContent(bData)
		if err != nil {
			return err
		}
		w.SetReader(bytes.NewReader(bData))
		return nil
	}
}

// WithReadImageReader allows option for users to specify the NIfTI bytes reader (.nii.gz or .nii)
func WithReadImageReader(r *bytes.Reader) func(*nifti.NiiReader) error {
	return func(w *nifti.NiiReader) error {
		bArr := make([]byte, r.Len())
		_, err := r.Read(bArr)
		if err != nil {
			return err
		}
		bArr, err = deflateFileContent(bArr)
		if err != nil {
			return err
		}
		w.SetReader(bytes.NewReader(bArr))
		return nil
	}
}

// WithReadHeaderReader allows option for users to specify the separate header file reader in case of NIfTI pair .hdr/.img
func WithReadHeaderReader(r *bytes.Reader) func(*nifti.NiiReader) error {
	return func(w *nifti.NiiReader) error {
		bArr := make([]byte, r.Len())
		_, err := r.Read(bArr)
		if err != nil {
			return err
		}
		bArr, err = deflateFileContent(bArr)
		if err != nil {
			return err
		}
		w.SetHdrReader(r)
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

// WithWriteCompression sets the option to write compressed NIfTI image to a single file (.nii.gz)
//
// If true, the whole file will be compressed. Default is false.
func WithWriteCompression(withCompression bool) func(writer *nifti.NiiWriter) {
	return func(w *nifti.NiiWriter) {
		w.SetCompression(withCompression)
	}
}

// WithWriteNii1Header sets the option to allow user to provide predefined NIfTI-1 header structure.
//
// If no header provided, the header will be converted from the NIfTI image structure
func WithWriteNii1Header(header *nifti.Nii1Header) func(*nifti.NiiWriter) {
	return func(w *nifti.NiiWriter) {
		w.SetHeader(header)
	}
}

// WithWriteNii2Header sets the option to allow user to provide predefined NIfTI-2 header structure.
//
// If no header provided, the header will be converted from the NIfTI image structure
func WithWriteNii2Header(header *nifti.Nii2Header) func(*nifti.NiiWriter) {
	return func(w *nifti.NiiWriter) {
		w.SetHeader(header)
	}
}

// WithWriteNIfTIData sets the option to allow user to provide predefined NIfTI-1 data structure.
func WithWriteNIfTIData(data *nifti.Nii) func(writer *nifti.NiiWriter) {
	return func(w *nifti.NiiWriter) {
		w.SetNiiData(data)
	}
}

// WithWriteVersion sets the option to specify the exported NIfTI version (NIfTI-1 or 2). Default is NIfTI-1
func WithWriteVersion(version int) func(writer *nifti.NiiWriter) {
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
