package middleware

import (
	"log"
	"net/http"
	"time"
)

type LoggerMW struct {
	next http.Handler
}

func NewLogger(next http.Handler) *LoggerMW {
	return &LoggerMW{
		next: next,
	}
}

func (l *LoggerMW) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	l.next.ServeHTTP(lrw, r)

	log.Printf("%s %s %d %v", r.Method, r.RequestURI, lrw.statusCode, time.Since(start))
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
