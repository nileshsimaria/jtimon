package main

import (
	"compress/zlib"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
)

func newDEFLATEDecompressor() grpc.Decompressor {
	return &deflateDecompressor{}
}

type deflateDecompressor struct {
}

func (d *deflateDecompressor) Do(r io.Reader) ([]byte, error) {
	z, err := zlib.NewReader(r)
	if err != nil {
		return nil, err
	}

	a, err := ioutil.ReadAll(z)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (d *deflateDecompressor) Type() string {
	return "deflate"
}

// End Deflate Decompression
