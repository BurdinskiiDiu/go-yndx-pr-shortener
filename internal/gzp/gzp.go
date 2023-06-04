package gzp

import (
	"compress/gzip"
	"io"
	"net/http"
)

type CompressWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func NewCompressWriter(w http.ResponseWriter, wr io.Writer) *CompressWriter {
	return &CompressWriter{
		ResponseWriter: w,
		writer:         wr,
	}
}

func (cW *CompressWriter) Write(p []byte) (int, error) {
	return cW.writer.Write(p)
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

/*
type CompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}



func (c *CompressWriter) Close() error {
	return c.zw.Close()
}
*/
/*
func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *CompressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

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
