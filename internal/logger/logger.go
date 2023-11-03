package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log *zap.SugaredLogger = zap.NewNop().Sugar()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl.Sugar()
	return nil
}

type (
	// берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	r.responseData.status = 200
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// log.Println("WriteHeader from Logger called")
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func LogRequest(handleFunc http.HandlerFunc) http.HandlerFunc {
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
		handleFunc(&lw, r)
		buf := make([]byte, 0)
		r.Body.Read(buf)
		duration := time.Since(start)

		Log.Infow(
			"Got request ",
			"method", r.Method,
			"uri", r.RequestURI,
			"request body", string(buf),
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	return logFn
}
