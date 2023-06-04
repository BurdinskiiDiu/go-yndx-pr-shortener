package gzp

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
)

type CompressWriter struct {
	http.ResponseWriter
	bf     bytes.Buffer
	writer io.Writer
}

func NewCompressWriter(w http.ResponseWriter, bf bytes.Buffer) *CompressWriter {
	return &CompressWriter{
		ResponseWriter: w,
		bf:             bf,
		writer:         gzip.NewWriter(&bf),
	}
}

func (cW *CompressWriter) Write(p []byte) (int, error) {
	//cW.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	lenBf, _ := cW.writer.Write(p)
	cW.ResponseWriter.Header().Set("Content-Length", strconv.Itoa(lenBf))
	return cW.ResponseWriter.Write(cW.bf.Bytes())
}

func (cW *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		cW.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	cW.ResponseWriter.WriteHeader(statusCode)
}

/*
func (cW *CompressWriter) Close() error {
	return cW.writer.Close()
}*/

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
