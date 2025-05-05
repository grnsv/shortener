package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/grnsv/shortener/internal/logger"
)

var supportedContentTypes = [2]string{"application/json", "text/html"}

func isSupportedContentType(contentType string) bool {
	for _, v := range supportedContentTypes {
		if contentType == v {
			return true
		}
	}

	return false
}

// compressWriter wraps http.ResponseWriter and gzip.Writer to provide gzip compression for responses.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// newCompressWriter creates a new compressWriter for the given http.ResponseWriter.
func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header returns the header map that will be sent by WriteHeader.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write writes compressed data to the underlying gzip.Writer.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader sends an HTTP response header with the provided status code and sets Content-Encoding if appropriate.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes the underlying gzip.Writer.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader wraps an io.ReadCloser and gzip.Reader to provide gzip decompression for requests.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// newCompressReader creates a new compressReader for the given io.ReadCloser.
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

// Read reads decompressed data from the underlying gzip.Reader.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes the underlying gzip.Reader and io.ReadCloser.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// WithCompressing is a middleware that handles gzip compression and decompression for supported requests and responses.
func WithCompressing(logger logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					logger.Errorf("failed to create gzip reader for request: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer func() {
					if err := cr.Close(); err != nil {
						logger.Errorf("failed to close gzip reader for request: %v", err)
					}
				}()
			}

			ow := w
			if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && isSupportedContentType(r.Header.Get("Content-Type")) {
				cw := newCompressWriter(w)
				ow = cw
				defer func() {
					if err := cw.Close(); err != nil {
						logger.Errorf("failed to close gzip writer for response: %v", err)
					}
				}()
			}

			next.ServeHTTP(ow, r)
		})
	}
}
