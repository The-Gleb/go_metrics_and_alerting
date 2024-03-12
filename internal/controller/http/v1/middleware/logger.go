package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type loggerMiddleware struct {
	log *zap.SugaredLogger
}

func NewLoggerMiddleware(l *zap.SugaredLogger) *loggerMiddleware {
	return &loggerMiddleware{l}
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.status = 200
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (md *loggerMiddleware) Do(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		buf := make([]byte, 0)
		r.Body.Read(buf)
		duration := time.Since(start)

		md.log.Infow(
			"Got request ",
			"method", r.Method,
			"uri", r.RequestURI,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}
