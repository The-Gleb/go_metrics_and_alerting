package compressor

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
)

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

func GzipMiddleware(h http.Handler) http.HandlerFunc {
	gzipMiddleware := func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("\nReq body encoded: %v\n", r.Body)
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
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	}
	return gzipMiddleware
}
