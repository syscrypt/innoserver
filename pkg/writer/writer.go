package writer

import (
	"net/http"
)

type ExtRespWriter interface {
	http.ResponseWriter
}

type extRespWriter struct {
	w      http.ResponseWriter
	Status int
}

func New(w http.ResponseWriter) *extRespWriter {
	return &extRespWriter{
		w:      w,
		Status: 0,
	}
}

func (s *extRespWriter) WriteHeader(statusCode int) {
	s.Status = statusCode
	s.w.WriteHeader(statusCode)
}

func (s *extRespWriter) Header() http.Header {
	return s.w.Header()
}

func (s *extRespWriter) Write(data []byte) (int, error) {
	return s.w.Write(data)
}
