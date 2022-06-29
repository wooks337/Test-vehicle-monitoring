package httpwares

import (
	"bytes"
	"net/http"
)

type Middleware func(handler http.Handler) http.Handler

type WrappedResponseWriter interface {
	http.ResponseWriter
	StatusCode() int
	MessageLength() int
	Body() *bytes.Buffer
}

type wrappedResponseWriter struct {
	http.ResponseWriter
	code       int
	bytes      int
	body       bytes.Buffer
	enableBody bool
}

func (w *wrappedResponseWriter) StatusCode() int {
	if w.code == 0 {
		return http.StatusOK
	}

	return w.code
}

func (w *wrappedResponseWriter) MessageLength() int {
	return w.bytes
}

func (w *wrappedResponseWriter) Body() *bytes.Buffer {
	return &w.body
}

func (w *wrappedResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func (w *wrappedResponseWriter) WriteHeader(code int) {
	w.code = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *wrappedResponseWriter) Write(b []byte) (int, error) {
	if w.code == 0 {
		w.code = 200
	}

	n, err := w.ResponseWriter.Write(b)
	if w.enableBody {
		w.body.Write(b)
	}

	w.bytes += n

	return n, err

}

func WrapHandler(h http.Handler, wares ...Middleware) http.Handler {
	if len(wares) < 1 {
		return h
	}

	wrapped := h
	for i := len(wares) - 1; i >= 0; i-- {
		wrapped = wares[i](wrapped)
	}

	return wrapped
}

func NewWrappedResponseWriter(w http.ResponseWriter, body bool) WrappedResponseWriter {
	wrapped := &wrappedResponseWriter{
		ResponseWriter: w,
		enableBody:     body,
	}

	return wrapped
}
