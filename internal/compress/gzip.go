// Package compress provides GZIP compression utilities for HTTP requests and responses.
// It includes pooled GZIP readers and writers for efficient memory usage.
package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

// gzipWriterPool maintains a pool of GZIP writers to reduce allocation overhead.
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return w
	},
}

// gzipReaderPool maintains a pool of GZIP readers to reduce allocation overhead.
var gzipReaderPool = sync.Pool{
	New: func() interface{} {
		return new(gzip.Reader)
	},
}

// GzipWriter wraps http.ResponseWriter with GZIP compression capabilities.
// It transparently compresses response data and sets appropriate headers.
type GzipWriter struct {
	rw     http.ResponseWriter
	gw     *gzip.Writer
	status int
	header http.Header
}

// NewGzipWriter creates a new GzipWriter that compresses data written to the response.
// Uses a pooled GZIP writer for better performance.
//
// Parameters:
//   - w: the underlying http.ResponseWriter to wrap
//
// Returns:
//   - *GzipWriter: initialized GZIP response writer
func NewGzipWriter(w http.ResponseWriter) *GzipWriter {
	gw := gzipWriterPool.Get().(*gzip.Writer)
	gw.Reset(w)

	return &GzipWriter{
		rw:     w,
		gw:     gw,
		header: w.Header().Clone(),
	}
}

// Header returns the HTTP header map that will be sent by WriteHeader.
// The headers are captured and applied when WriteHeader is called.
//
// Returns:
//   - http.Header: the response headers
func (w *GzipWriter) Header() http.Header {
	return w.header
}

// Write writes compressed data to the response.
// If WriteHeader hasn't been called, it calls WriteHeader(http.StatusOK) first.
//
// Parameters:
//   - p: byte slice to write
//
// Returns:
//   - int: number of bytes written
//   - error: error if writing fails
func (w *GzipWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.gw.Write(p)
}

// WriteHeader sends an HTTP response header with the provided status code.
// Sets Content-Encoding header to "gzip" for compressible status codes.
//
// Parameters:
//   - statusCode: HTTP status code to send
func (w *GzipWriter) WriteHeader(statusCode int) {
	if w.status != 0 {
		return
	}

	w.status = statusCode

	for k, v := range w.header {
		w.rw.Header()[k] = v
	}

	if statusCode < http.StatusNoContent || statusCode == http.StatusConflict {
		w.rw.Header().Set("Content-Encoding", "gzip")
	}
	w.rw.WriteHeader(statusCode)
}

// Close closes the GZIP writer and returns it to the pool.
// This should be called when done writing to ensure proper cleanup.
//
// Returns:
//   - error: error if closing the GZIP writer fails
func (w *GzipWriter) Close() error {
	if w.gw == nil {
		return nil
	}

	err := w.gw.Close()
	gzipWriterPool.Put(w.gw)
	w.gw = nil

	return err
}

// GzipReader wraps an io.ReadCloser with GZIP decompression capabilities.
// It transparently decompresses request data from GZIP format.
type GzipReader struct {
	rc io.ReadCloser
	gr *gzip.Reader
}

// NewGzipReader creates a new GzipReader that decompresses data from the request.
// Uses a pooled GZIP reader for better performance.
//
// Parameters:
//   - r: the underlying io.ReadCloser to wrap
//
// Returns:
//   - *GzipReader: initialized GZIP reader
//   - error: error if GZIP reader initialization fails
func NewGzipReader(r io.ReadCloser) (*GzipReader, error) {
	gr := gzipReaderPool.Get().(*gzip.Reader)

	if err := gr.Reset(r); err != nil {
		gzipReaderPool.Put(gr)
		return nil, err
	}
	return &GzipReader{
		rc: r,
		gr: gr,
	}, nil
}

// Read reads decompressed data from the GZIP stream.
//
// Parameters:
//   - p: byte slice to read into
//
// Returns:
//   - int: number of bytes read
//   - error: error if reading fails
func (r *GzipReader) Read(p []byte) (n int, err error) {
	return r.gr.Read(p)
}

// Close closes the GZIP reader and returns it to the pool.
// Also closes the underlying io.ReadCloser.
//
// Returns:
//   - error: error if closing fails
func (r *GzipReader) Close() error {
	if r.gr != nil {
		gzipReaderPool.Put(r.gr)
		r.gr = nil
	}
	return r.rc.Close()
}
