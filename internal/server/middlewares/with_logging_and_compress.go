package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
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

func shouldCompressRequest(req *http.Request) bool {
	return req.Header.Get("Content-Encoding") == "gzip" &&
		(req.Header.Get("Content-Type") == "application/json" || 
		 req.Header.Get("Content-Type") == "text/html")
}

func shouldCompressResponse(req *http.Request) bool {
	acceptEncoding := req.Header.Get("Accept-Encoding")
	contentType := req.Header.Get("Content-Type")
	
	return strings.Contains(acceptEncoding, "gzip") &&
		(contentType == "application/json" || contentType == "text/html")
}

// decompressRequestBody декомпрессирует тело запроса если нужно
func decompressRequestBody(req *http.Request) error {
	if !shouldCompressRequest(req) {
		return nil
	}

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

// compressResponseBody компрессирует тело ответа если нужно
func compressResponseBody(res http.ResponseWriter, req *http.Request, data []byte) error {
	if !shouldCompressResponse(req) || len(data) == 0 {
		_, err := res.Write(data)
		return err
	}

	var compressedBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&compressedBuf)

	if _, err := gzWriter.Write(data); err != nil {
		return err
	}

	if err := gzWriter.Close(); err != nil {
		return err
	}

	res.Header().Set("Content-Encoding", "gzip")
	res.Header().Del("Content-Length")

	_, err := res.Write(compressedBuf.Bytes())
	return err
}

// WithLoggingAndCompress middleware для логирования и компрессии
func WithLoggingAndCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()

		// Декомпрессия входящего запроса
		if err := decompressRequestBody(req); err != nil {
			http.Error(res, "Error decompressing request: "+err.Error(), http.StatusBadRequest)
			return
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

		// Компрессия исходящего ответа
		if err := compressResponseBody(res, req, responseWriter.buffer.Bytes()); err != nil {
			http.Error(res, "Error compressing response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Логирование
		duration := time.Since(start)
		logger.Logger.Infoln(
			"uri", req.RequestURI,
			"method", req.Method,
			"duration", duration,
			"status", responseData.status,
			"size", responseData.size,
		)
	})
}