// Package middlewares предоставляет промежуточное ПО (middleware) для HTTP-сервера.
package middlewares

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// compressionWriter оборачивает ResponseWriter для буферизации ответа.
type compressionWriter struct {
	http.ResponseWriter
	buffer      *bytes.Buffer
	header      http.Header
	statusCode  int
	wroteHeader bool
}

// newCompressionWriter создает новый compressionWriter.
func newCompressionWriter(w http.ResponseWriter) *compressionWriter {
	return &compressionWriter{
		ResponseWriter: w,
		buffer:         &bytes.Buffer{},
		header:         make(http.Header),
		statusCode:     http.StatusOK,
	}
}

// Header возвращает заголовки ответа.
func (w *compressionWriter) Header() http.Header {
	return w.header
}

// Write записывает данные в буфер.
func (w *compressionWriter) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

// WriteHeader устанавливает статус код ответа.
func (w *compressionWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.wroteHeader = true
}

// shouldDecompressRequest проверяет, нужно ли распаковывать тело запроса.
func shouldDecompressRequest(req *http.Request) bool {
	return req.Header.Get("Content-Encoding") == "gzip" &&
		(req.Header.Get("Content-Type") == "application/json" ||
			req.Header.Get("Content-Type") == "text/html")
}

// shouldCompressResponse проверяет, нужно ли сжимать тело ответа.
func shouldCompressResponse(req *http.Request, responseContentType string) bool {
	acceptEncoding := req.Header.Get("Accept-Encoding")

	return strings.Contains(acceptEncoding, "gzip") &&
		(responseContentType == "application/json" || responseContentType == "text/html")
}

// decompressRequestBody распаковывает сжатое тело запроса.
func decompressRequestBody(req *http.Request) error {
	gzReader, err := gzip.NewReader(req.Body)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	decompressedBody, err := io.ReadAll(gzReader)
	if err != nil {
		return fmt.Errorf("error decompressing request: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewBuffer(decompressedBody))
	req.ContentLength = int64(len(decompressedBody))
	req.Header.Del("Content-Encoding")

	return nil
}

// compressResponseBody сжимает тело ответа с использованием gzip.
func compressResponseBody(data []byte) ([]byte, error) {
	var compressedBuf bytes.Buffer
	gzWriter := gzip.NewWriter(&compressedBuf)

	if _, err := gzWriter.Write(data); err != nil {
		return nil, err
	}

	if err := gzWriter.Close(); err != nil {
		return nil, err
	}

	return compressedBuf.Bytes(), nil
}

// WithCompression возвра middleware для сжатия ответов и распаковки запросов.
func WithCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if shouldDecompressRequest(req) {
			if err := decompressRequestBody(req); err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}
		}

		compWriter := newCompressionWriter(res)
		next.ServeHTTP(compWriter, req)

		responseContentType := compWriter.header.Get("Content-Type")
		shouldCompress := shouldCompressResponse(req, responseContentType)

		var finalBody []byte

		if shouldCompress && compWriter.buffer.Len() > 0 {
			compressedBody, err := compressResponseBody(compWriter.buffer.Bytes())
			if err != nil {
				http.Error(res, "Error compressing response: "+err.Error(), http.StatusInternalServerError)
				return
			}
			finalBody = compressedBody
			res.Header().Set("Content-Encoding", "gzip")
			res.Header().Set("Vary", "Accept-Encoding")
		} else {
			finalBody = compWriter.buffer.Bytes()
		}

		for key, values := range compWriter.header {
			res.Header()[key] = values
		}

		res.Header().Set("Content-Length", strconv.Itoa(len(finalBody)))

		if compWriter.wroteHeader {
			res.WriteHeader(compWriter.statusCode)
		}

		res.Write(finalBody)
	})
}