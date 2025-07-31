package compress

import (
	"compress/gzip"
	"net/http"
)

type GzipWriter struct {
	rw http.ResponseWriter
	gw *gzip.Writer
}

func NewGzipWriter(w http.ResponseWriter) *GzipWriter {
	return &GzipWriter{
		rw: w,
		gw: gzip.NewWriter(w),
	}
}

func (w GzipWriter) Header() http.Header {
	return w.rw.Header()
}

func (w GzipWriter) Write(p []byte) (int, error) {
	return w.gw.Write(p)
}

func (w GzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		w.rw.Header().Set("Content-Encoding", "gzip")
	}
	w.rw.WriteHeader(statusCode)
}

func (w GzipWriter) Close() error {
	return w.gw.Close()
}
