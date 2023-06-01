package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type compressWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
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

func GZipMiddleware(h http.HandlerFunc, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w
		accptEnc := r.Header.Get("Accept-Encoding")
		logger.Info("acceptEnc", zap.String("accptEnc", string(accptEnc)))
		cntntEnc := r.Header.Get("Content-Encoding")
		logger.Info("cntntEnc", zap.String("cntntEnc", string(cntntEnc)))
		sendGZip := strings.Contains(cntntEnc, "gzip")
		if sendGZip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				logger.Debug("compersReader creation err", zap.String("err", err.Error()))
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		logger.Info("response", zap.String("response", r.RequestURI))

		h(ow, r)
	}
}
