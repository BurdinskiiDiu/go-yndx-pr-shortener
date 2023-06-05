package gzp

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strconv"
)

type CompressWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		ResponseWriter: w,
		writer:         gzip.NewWriter(w),
	}
}

func (cW *CompressWriter) Write(p []byte) (int, error) {
	//cW.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	var bf bytes.Buffer
	wrt := gzip.NewWriter(&bf)
	/*_, err := */ wrt.Write(p)
	/*if err != nil {

	} */
	wrt.Close()
	strng := bf.String()
	log.Println("compessed response is: ", strng, "and it len is: ", len(strng))
	cW.ResponseWriter.Header().Set("Content-lenght", strconv.Itoa(len(bf.Bytes())))
	return cW.writer.Write(p)

}

func (cW *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		cW.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	cW.ResponseWriter.WriteHeader(statusCode)
}

func (cW *CompressWriter) Close() error {
	return cW.writer.Close()
}

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
