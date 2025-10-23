package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return w
	},
}

var gzipReaderPool = sync.Pool{
	New: func() interface{} {
		return new(gzip.Reader)
	},
}

type GzipWriter struct {
	rw     http.ResponseWriter
	gw     *gzip.Writer
	status int
	header http.Header
}

func NewGzipWriter(w http.ResponseWriter) *GzipWriter {
	gw := gzipWriterPool.Get().(*gzip.Writer)
	gw.Reset(w)

	return &GzipWriter{
		rw:     w,
		gw:     gw,
		header: w.Header().Clone(),
	}
}

func (w *GzipWriter) Header() http.Header {
	return w.header
}

func (w *GzipWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	return w.gw.Write(p)
}

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

func (w *GzipWriter) Close() error {
	if w.gw == nil {
		return nil
	}

	err := w.gw.Close()
	gzipWriterPool.Put(w.gw)
	w.gw = nil

	return err
}

type GzipReader struct {
	rc io.ReadCloser
	gr *gzip.Reader
}

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

func (r *GzipReader) Read(p []byte) (n int, err error) {
	return r.gr.Read(p)
}

func (r *GzipReader) Close() error {
	if r.gr != nil {
		gzipReaderPool.Put(r.gr)
		r.gr = nil
	}
	return r.rc.Close()
}
