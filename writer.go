package cache

import "net/http"

type responseWriter struct {
	http.ResponseWriter
	header      http.Header
	statusCode  int
	body        []byte
	wroteHeader bool
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.writeHeader()
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = b
	w.writeHeader()
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) writeHeader() {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	for key := range w.ResponseWriter.Header() {
		// clear out so below we just write the current state
		w.ResponseWriter.Header().Del(key)
	}

	for k, v := range w.Header() {
		for _, vv := range v {
			w.ResponseWriter.Header().Add(k, vv) // Header already internally handled the difference between Set and Add
		}
	}
}

func (w *responseWriter) Header() http.Header {
	if w.header == nil {
		w.header = w.ResponseWriter.Header().Clone()
	}
	return w.header
}
