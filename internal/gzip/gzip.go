package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
)

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
	//defer c.Close()
	return c.zw.Write(p)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.zw.Close()
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

/*
func GZipMiddleware(h http.Handler, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		accptEnc := r.Header.Get("Accept-Encoding")
		logger.Info("acceptEnc", zap.String("accptEnc", string(accptEnc)))
		cntntEnc := r.Header.Get("Content-Encoding")
		logger.Info("cntntEnc", zap.String("cntntEnc", string(cntntEnc)))
		sendGZip := strings.Contains(cntntEnc, "gzip")
		if sendGZip {
			cr, err := NewCompressReader(r.Body)
			if err != nil {
				logger.Debug("compersReader creation err", zap.String("err", err.Error()))
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		logger.Info("response", zap.String("response", r.RequestURI))

		h.ServeHTTP(ow, r)
	})
}*/
