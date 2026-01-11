package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

type HashConfig struct {
	SecretKey string
}

type hashWriter struct {
	http.ResponseWriter
	secretKey   string
	buffer      *bytes.Buffer
	header      http.Header
	statusCode  int
	wroteHeader bool
}

func newHashWriter(w http.ResponseWriter, secretKey string) *hashWriter {
	return &hashWriter{
		ResponseWriter: w,
		secretKey:      secretKey,
		buffer:         &bytes.Buffer{},
		header:         make(http.Header),
		statusCode:     http.StatusOK,
	}
}

func (w *hashWriter) Header() http.Header {
	return w.header
}

func (w *hashWriter) Write(data []byte) (int, error) {
	return w.buffer.Write(data)
}

func (w *hashWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.wroteHeader = true
}

func calculateHash(data []byte, secretKey string) string {
	if secretKey == "" {
		return ""
	}

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func shouldValidateHash(req *http.Request) bool {
	return (req.Method == http.MethodPost || req.Method == http.MethodPut) &&
		req.Body != nil && req.Body != http.NoBody
}

func WithHashing(config *HashConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if config.SecretKey == "" {
				next.ServeHTTP(res, req)
				return
			}

			if shouldValidateHash(req) {
				bodyBytes, err := io.ReadAll(req.Body)
				if err != nil {
					http.Error(res, "Error reading request body", http.StatusBadRequest)
					return
				}

				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				receivedHash := req.Header.Get("HashSHA256")

				if receivedHash != "" {
					expectedHash := calculateHash(bodyBytes, config.SecretKey)

					if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
						logger.Logger.Warnln("Hash validation failed - JSON mismatch")
					}
				}
			}

			hashWriter := newHashWriter(res, config.SecretKey)
			next.ServeHTTP(hashWriter, req)

			if hashWriter.buffer.Len() > 0 {
				hash := calculateHash(hashWriter.buffer.Bytes(), config.SecretKey)
				hashWriter.header.Set("HashSHA256", hash)
			}

			for key, values := range hashWriter.header {
				res.Header()[key] = values
			}
			res.WriteHeader(hashWriter.statusCode)
			res.Write(hashWriter.buffer.Bytes())
		})
	}
}
