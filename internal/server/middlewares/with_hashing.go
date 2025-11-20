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

// HashConfig конфигурация для middleware подписи
type HashConfig struct {
	SecretKey string
}

// hashWriter перехватывает ответ для вычисления хеша
type hashWriter struct {
	http.ResponseWriter
	secretKey  string
	buffer     *bytes.Buffer
	header     http.Header
	statusCode int
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

// calculateHash вычисляет HMAC-SHA256 хеш для данных
func calculateHash(data []byte, secretKey string) string {
	if secretKey == "" {
		return ""
	}

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// shouldValidateHash проверяет нужно ли валидировать хеш для запроса
func shouldValidateHash(req *http.Request) bool {
	// Проверяем только для методов с телом
	return (req.Method == http.MethodPost || req.Method == http.MethodPut) &&
		req.Body != nil && req.Body != http.NoBody
}

// WithHash middleware для проверки входящих и подписи исходящих данных
func WithHashing(config *HashConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
            logger.Logger.Infof("Server hash key: '%s' (length: %d)", config.SecretKey, len(config.SecretKey))
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

                logger.Logger.Infof("Hash validation - Received hash: %s", receivedHash)
				logger.Logger.Infof("Hash validation - Body length: %d", len(bodyBytes))
				logger.Logger.Infof("Hash validation - Body (first 100 chars): %s", string(bodyBytes[:min(100, len(bodyBytes))]))
				logger.Logger.Infof("Hash validation - Content-Encoding: %s", req.Header.Get("Content-Encoding"))
				logger.Logger.Infof("Hash validation - Content-Type: %s", req.Header.Get("Content-Type"))

                if receivedHash != "" {
                    expectedHash := calculateHash(bodyBytes, config.SecretKey)
                    logger.Logger.Infof("Hash validation - Expected hash: %s", expectedHash)

                    if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
                        logger.Logger.Warnln("Hash validation failed - JSON mismatch")
                        logger.Logger.Infof("Hash validation - Received: %s", receivedHash)
						logger.Logger.Infof("Hash validation - Expected: %s", expectedHash)
                        http.Error(res, "Invalid hash signature", http.StatusBadRequest)
                        return
                    }
                    logger.Logger.Infof("Hash validation successful")
                }
            }

            hashWriter := newHashWriter(res, config.SecretKey)
            next.ServeHTTP(hashWriter, req)

            // Для ответа хеш считаем от JSON (до компрессии)
            if hashWriter.buffer.Len() > 0 {
                hash := calculateHash(hashWriter.buffer.Bytes(), config.SecretKey)
                hashWriter.header.Set("HashSHA256", hash)
            }

            // Отправляем ответ
            for key, values := range hashWriter.header {
                res.Header()[key] = values
            }
            res.WriteHeader(hashWriter.statusCode)
            res.Write(hashWriter.buffer.Bytes())
        })
    }
}