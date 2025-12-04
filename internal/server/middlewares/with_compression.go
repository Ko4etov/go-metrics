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
		return fmt.Errorf("error decompressing request: %w", err)
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

// WithCompression middleware для компрессии/декомпрессии
func WithCompression(next http.Handler) http.Handler {
    return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
        // Декомпрессия входящего запроса если нужно
        if shouldDecompressRequest(req) {
            if err := decompressRequestBody(req); err != nil {
                http.Error(res, err.Error(), http.StatusBadRequest)
                return
            }
        }

        compWriter := newCompressionWriter(res)
        next.ServeHTTP(compWriter, req)

        // Компрессия ответа если нужно
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

        // Копируем заголовки
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