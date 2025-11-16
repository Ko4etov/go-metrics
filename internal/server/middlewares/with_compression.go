package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// compressionWriter перехватывает ответ для компрессии
type compressionWriter struct {
	http.ResponseWriter
	buffer       *bytes.Buffer
	header       http.Header
	statusCode   int
	wroteHeader  bool
}

func newCompressionWriter(w http.ResponseWriter) *compressionWriter {
	return &compressionWriter{
		ResponseWriter: w,
		buffer:         &bytes.Buffer{},
		header:         make(http.Header),
		statusCode:     http.StatusOK,
	}
}

func (w *compressionWriter) Header() http.Header {
	return w.header
}

func (w *compressionWriter) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

func (w *compressionWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.wroteHeader = true
}

// shouldDecompressRequest проверяет нужно ли декомпрессировать запрос
func shouldDecompressRequest(req *http.Request) bool {
	return req.Header.Get("Content-Encoding") == "gzip" &&
		(req.Header.Get("Content-Type") == "application/json" || 
		 req.Header.Get("Content-Type") == "text/html")
}

// shouldCompressResponse проверяет нужно ли компрессировать ответ
func shouldCompressResponse(req *http.Request, responseContentType string) bool {
	acceptEncoding := req.Header.Get("Accept-Encoding")
	
	return strings.Contains(acceptEncoding, "gzip") &&
		(responseContentType == "application/json" || responseContentType == "text/html")
}

// decompressRequestBody декомпрессирует тело запроса
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

// compressResponseBody компрессирует тело ответа
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

// WithCompression middleware для компрессии/декомпрессии
func WithCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// Создаем копию запроса для безопасной модификации
		newReq := req.Clone(req.Context())
		
		// Декомпрессия входящего запроса если нужно
		if shouldDecompressRequest(req) {
			if err := decompressRequestBody(newReq); err != nil {
				http.Error(res, "Error decompressing request: "+err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			// Копируем тело если не декомпрессируем
			if req.Body != nil && req.Body != http.NoBody {
				body, err := io.ReadAll(req.Body)
				if err != nil {
					http.Error(res, "Error reading request body", http.StatusInternalServerError)
					return
				}
				newReq.Body = io.NopCloser(bytes.NewBuffer(body))
				newReq.ContentLength = int64(len(body))
			}
		}

		// Создаем writer для перехвата ответа
		compWriter := newCompressionWriter(res)

		// Вызываем следующий обработчик
		next.ServeHTTP(compWriter, newReq)

		// Получаем Content-Type ответа
		responseContentType := compWriter.header.Get("Content-Type")
		shouldCompress := shouldCompressResponse(req, responseContentType)

		var finalBody []byte
		
		// Компрессия ответа если нужно
		if shouldCompress && compWriter.buffer.Len() > 0 {
			compressedBody, err := compressResponseBody(compWriter.buffer.Bytes())
			if err != nil {
				http.Error(res, "Error compressing response: "+err.Error(), http.StatusInternalServerError)
				return
			}
			finalBody = compressedBody
		} else {
			finalBody = compWriter.buffer.Bytes()
		}

		// Копируем заголовки из перехваченного ответа
		for key, values := range compWriter.header {
			for _, value := range values {
				res.Header().Set(key, value)
			}
		}

		// Устанавливаем заголовки компрессии если нужно
		if shouldCompress {
			res.Header().Set("Content-Encoding", "gzip")
			res.Header().Set("Vary", "Accept-Encoding")
		}

		// Устанавливаем Content-Length
		res.Header().Set("Content-Length", strconv.Itoa(len(finalBody)))

		// Отправляем статус и тело
		if compWriter.wroteHeader {
			res.WriteHeader(compWriter.statusCode)
		} else {
			res.WriteHeader(http.StatusOK)
		}
		
		if _, err := res.Write(finalBody); err != nil {
			// Логируем ошибку, но не можем вернуть её
			return
		}
	})
}