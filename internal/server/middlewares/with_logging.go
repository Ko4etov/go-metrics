package middlewares

import (
	"net/http"
	"time"

	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

type (
    // берём структуру для хранения сведений об ответе
    responseData struct {
        status int
        size int
    }

    // добавляем реализацию http.ResponseWriter
    loggingResponseWriter struct {
        http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
        responseData *responseData
    }
)

func WithLogging(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        responseData := &responseData {
            status: 0,
            size: 0,
        }

        lw := loggingResponseWriter {
            ResponseWriter: w,
            responseData: responseData,
        }

        h.ServeHTTP(&lw, r)

        duration := time.Since(start)

        // отправляем сведения о запросе в zap
        logger.Logger.Infoln(
            "uri", r.RequestURI,
            "method", r.Method,
            "duration", duration,
            "status", responseData.status,
            "size", responseData.size,
        )
        
    })
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
    // записываем ответ, используя оригинальный http.ResponseWriter
    size, err := r.ResponseWriter.Write(b) 
    r.responseData.size += size // захватываем размер
    return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
    // записываем код статуса, используя оригинальный http.ResponseWriter
    r.ResponseWriter.WriteHeader(statusCode) 
    r.responseData.status = statusCode // захватываем код статуса
} 