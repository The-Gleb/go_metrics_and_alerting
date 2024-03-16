package middleware

import (
	"compress/gzip"
	"fmt"
	"io"

	"net/http"
	"strings"

	"github.com/The-Gleb/go_metrics_and_alerting/internal/logger"
)

type gzipMiddleware struct {
}

func NewGzipMiddleware() *gzipMiddleware {
	return &gzipMiddleware{}
}

type compressWriter struct {
	rw http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(rw http.ResponseWriter) *compressWriter {
	return &compressWriter{
		rw: rw,
		zw: gzip.NewWriter(rw),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.rw.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if c.rw.Header().Get("Content-Type") == "application/json" || c.rw.Header().Get("Content-Type") == "text/html" {
		c.rw.Header().Set("Content-Encoding", "gzip")
		return c.zw.Write(p)
	}
	return c.rw.Write(p)
}
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.rw.Header().Set("Content-Encoding", "gzip")
	}
	c.rw.WriteHeader(statusCode)
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
		return nil, fmt.Errorf("gzip.NewReader: %w", err)
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

func (md *gzipMiddleware) Do(h http.Handler) http.Handler {
	gzipMiddleware := func(rw http.ResponseWriter, r *http.Request) {
		ow := rw

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := newCompressWriter(rw)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				logger.Log.Error(err)
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(gzipMiddleware)
}
