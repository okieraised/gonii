package utils

import (
	"bytes"
	gzip "github.com/klauspost/pgzip"
	"io"
)

func DeflateGzip(b []byte) ([]byte, error) {
	br := bytes.NewReader(b)
	g, err := gzip.NewReader(br)
	if err != nil {
		return nil, err
	}
	defer g.Close()

	p, err := io.ReadAll(g)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, err
	}

	return p, nil
}
