package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

type (
	responseData struct {
		status int
		size   int
	}

	customResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
		buffer       *bytes.Buffer
	}
)

func (w *customResponseWriter) Write(b []byte) (int, error) {
	size, err := w.buffer.Write(b)
	w.responseData.size += size
	return size, err
}

func (w *customResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.responseData.status = statusCode
}

func (w *customResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header()
}

func shouldDecompressRequest(req *http.Request) bool {
	return req.Header.Get("Content-Encoding") == "gzip" &&
		(req.Header.Get("Content-Type") == "application/json" || 
		 req.Header.Get("Content-Type") == "text/html")
}

func shouldCompressResponse(req *http.Request, responseContentType string) bool {
	acceptEncoding := req.Header.Get("Accept-Encoding")
	
	return strings.Contains(acceptEncoding, "gzip") &&
		(responseContentType == "application/json" || responseContentType == "text/html")
}

// decompressRequestBody декомпрессирует тело запроса если нужно
func decompressRequestBody(req *http.Request) error {
	gzReader, err := gzip.NewReader(req.Body)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	decompressedBody, err := io.ReadAll(gzReader)
	if err != nil {
		return err
	}

	req.Body = io.NopCloser(bytes.NewBuffer(decompressedBody))
	req.ContentLength = int64(len(decompressedBody))
	req.Header.Del("Content-Encoding")

	return nil
}

// WithLoggingAndCompress middleware для логирования и компрессии
func WithLoggingAndCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()

		logger.Logger.Infoln(
			"request_headers", req.Header,
			"shouldDecompressRequest", shouldDecompressRequest(req),
		)

		// Декомпрессия входящего запроса
		if shouldDecompressRequest(req) {
			if err := decompressRequestBody(req); err != nil {
				http.Error(res, "Error decompressing request: "+err.Error(), http.StatusBadRequest)
				return
			}
		}

		// Подготовка для перехвата ответа
		responseData := &responseData{}
		responseWriter := &customResponseWriter{
			ResponseWriter: res,
			responseData:   responseData,
			buffer:         &bytes.Buffer{},
		}

		// Вызов следующего обработчика
		next.ServeHTTP(responseWriter, req)

		resBody := responseWriter.buffer.Bytes()
		responseContentType := responseWriter.Header().Get("Content-Type")

		logger.Logger.Infoln(
			"compression_check",
			"response_content_type", responseContentType,
			"body_size", len(resBody),
			"should_compress", shouldCompressResponse(req, responseContentType),
		)

		// ✅ ПРАВИЛЬНЫЙ ПОРЯДОК: сначала копируем заголовки, потом решаем про компрессию
		headers := responseWriter.Header()
		for key, values := range headers {
			for _, value := range values {
				res.Header().Set(key, value)
			}
		}

		// Обработка ответа
		if shouldCompressResponse(req, responseContentType) && len(resBody) > 0 {
			// Компрессия ответа
			var compressedBuf bytes.Buffer
			gzWriter := gzip.NewWriter(&compressedBuf)

			if _, err := gzWriter.Write(resBody); err != nil {
				http.Error(res, "Error compressing response: "+err.Error(), http.StatusInternalServerError)
				return
			}

			if err := gzWriter.Close(); err != nil {
				http.Error(res, "Error closing gzip writer: "+err.Error(), http.StatusInternalServerError)
				return
			}

			compressedData := compressedBuf.Bytes()
			
			// Устанавливаем заголовки компрессии
			res.Header().Set("Content-Encoding", "gzip")
			res.Header().Set("Vary", "Accept-Encoding")
			res.Header().Set("Content-Length", strconv.Itoa(len(compressedData)))
			
			// Отправляем сжатый ответ
			res.WriteHeader(responseData.status)
			if _, err := res.Write(compressedData); err != nil {
				logger.Logger.Errorln("Error writing compressed response:", err)
				return
			}

			logger.Logger.Infoln(
				"response_sent", "compressed",
				"original_size", len(resBody),
				"compressed_size", len(compressedData),
			)

		} else {
			// Отправляем несжатый ответ
			res.Header().Set("Content-Length", strconv.Itoa(len(resBody)))
			res.WriteHeader(responseData.status)
			if _, err := res.Write(resBody); err != nil {
				logger.Logger.Errorln("Error writing response:", err)
				return
			}

			logger.Logger.Infoln("response_sent", "uncompressed")
		}

		// Логирование
		duration := time.Since(start)
		logger.Logger.Infoln(
			"request_completed",
			"uri", req.RequestURI,
			"method", req.Method,
			"duration", duration,
			"status", responseData.status,
			"size", responseData.size,
		)
	})
}