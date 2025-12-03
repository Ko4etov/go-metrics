package middlewares

import (
	"bytes"
	"net/http"
)

// loggingWriter перехватывает ответ для логирования
type loggingWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
	buffer     *bytes.Buffer
	header     http.Header
}

func newLoggingWriter(w http.ResponseWriter) *loggingWriter {
	return &loggingWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		buffer:         &bytes.Buffer{},
		header:         make(http.Header),
	}
}

func (w *loggingWriter) Header() http.Header {
	return w.header
}

func (w *loggingWriter) Write(data []byte) (int, error) {
	size, err := w.buffer.Write(data)
	w.size += size
	return size, err
}

func (w *loggingWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

// WithLogging middleware для логирования запросов и ответов
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		// Создаем writer для перехвата ответа
		logWriter := newLoggingWriter(res)

		// Вызываем следующий обработчик
		next.ServeHTTP(logWriter, req)

		// Копируем заголовки и отправляем ответ
		for key, values := range logWriter.header {
			for _, value := range values {
				res.Header().Set(key, value)
			}
		}

		res.WriteHeader(logWriter.statusCode)
		if _, err := res.Write(logWriter.buffer.Bytes()); err != nil {
			return
		}
	})
}