package httpproxy

import (
	"io"
	"net/http"
)

type WriterFlusher interface {
	io.Writer
	http.Flusher
}

type flushWriter struct {
	WriterFlusher
}

func newFlushWriter(w io.Writer) io.Writer {
	if wf, ok := w.(WriterFlusher); ok {
		return flushWriter{wf}
	}
	return w
}

func (fw flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.WriterFlusher.Write(p)
	fw.WriterFlusher.Flush()
	return
}
