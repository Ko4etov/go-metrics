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
		header       http.Header
	}
)

func (w *customResponseWriter) Write(b []byte) (int, error) {
	size, err := w.buffer.Write(b)
	w.responseData.size += size
	return size, err
}

func (w *customResponseWriter) WriteHeader(statusCode int) {
	w.responseData.status = statusCode
}

func (w *customResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
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

// WithLoggingAndCompress middleware для логирования и компрессии
func WithLoggingAndCompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()

		newReq := req.Clone(req.Context())
		
		if shouldDecompressRequest(req) {
			if err := decompressRequestBody(newReq); err != nil {
				http.Error(res, "Error decompressing request: "+err.Error(), http.StatusBadRequest)
				return
			}
		} else {
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

		responseData := &responseData{status: http.StatusOK}
		responseWriter := &customResponseWriter{
			ResponseWriter: res,
			responseData:   responseData,
			buffer:         &bytes.Buffer{},
			header:         make(http.Header),
		}

		next.ServeHTTP(responseWriter, newReq)

		responseContentType := responseWriter.header.Get("Content-Type")
		shouldCompress := shouldCompressResponse(req, responseContentType)

		var finalBody []byte
		
		if shouldCompress && responseWriter.buffer.Len() > 0 {
			compressedBody, err := compressResponseBody(responseWriter.buffer.Bytes())
			if err != nil {
				http.Error(res, "Error compressing response: "+err.Error(), http.StatusInternalServerError)
				return
			}
			finalBody = compressedBody
		} else {
			finalBody = responseWriter.buffer.Bytes()
		}

		for key, values := range responseWriter.header {
			for _, value := range values {
				res.Header().Set(key, value)
			}
		}

		if shouldCompress {
			res.Header().Set("Content-Encoding", "gzip")
			res.Header().Set("Vary", "Accept-Encoding")
		}

		res.Header().Set("Content-Length", strconv.Itoa(len(finalBody)))

		res.WriteHeader(responseData.status)
		if _, err := res.Write(finalBody); err != nil {
			logger.Logger.Errorln("Error writing response:", err)
			return
		}

		// Логирование
		duration := time.Since(start)
		logger.Logger.Infoln(
			"uri", req.RequestURI,
			"method", req.Method,
			"duration", duration,
			"status", responseData.status,
			"size", len(finalBody),
			"compressed", shouldCompress,
		)
	})
}