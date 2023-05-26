package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/BurdinskiiDiu/go-yndx-pr-shortener.git/internal/logger"
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

/*
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}*/
/*
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}*/
/*
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}*/

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

func GZipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w
		accptEnc := r.Header.Get("Accept-Encoding")
		suppGZip := strings.Contains(accptEnc, "gzip")
		if suppGZip {
			cw := newCompressWriter(w)
			//cw.Header().Set("Content-Encoding", "gzip")
			ow = cw
			ow.Header().Set("Content-Encoding", "gzip")
			defer cw.Close()
		}
		cntntEnc := r.Header.Get("Content-Encoding")
		sendGZip := strings.Contains(cntntEnc, "gzip")
		//sendGZip := suppGZip
		if sendGZip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				logger.Log.Debug("compersReader creation err", zap.String("err", err.Error()))
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		h.ServeHTTP(ow, r)
	}
}
