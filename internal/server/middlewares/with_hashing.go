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
			// Если ключ не установлен - пропускаем проверку
			if config.SecretKey == "" {
				next.ServeHTTP(res, req)
				return
			}

			// Проверяем входящую подпись для запросов с телом
			if shouldValidateHash(req) {
				// Читаем тело запроса
				body, err := io.ReadAll(req.Body)
				if err != nil {
					logger.Logger.Warnln("Error reading request body for hash validation:", err)
					http.Error(res, "Error reading request body", http.StatusBadRequest)
					return
				}
				
				// Восстанавливаем тело для следующих обработчиков
				req.Body = io.NopCloser(bytes.NewBuffer(body))

				// Проверяем подпись если она есть в заголовках
				receivedHash := req.Header.Get("HashSHA256")
				if receivedHash != "" {
					expectedHash := calculateHash(body, config.SecretKey)
					
					if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
						logger.Logger.Warnln(
							"Hash validation failed",
							"uri", req.RequestURI,
							"received_hash", receivedHash,
							"expected_hash", expectedHash,
						)
						http.Error(res, "Invalid hash signature", http.StatusBadRequest)
						return
					}
					
					logger.Logger.Debugln("Hash validation successful for", req.RequestURI)
				} else {
					logger.Logger.Debugln("No hash header for request", req.RequestURI)
				}
			}

			// Создаем writer для перехвата ответа
			hashWriter := newHashWriter(res, config.SecretKey)

			// Вызываем следующий обработчик
			next.ServeHTTP(hashWriter, req)

			// Добавляем подпись к ответу если есть данные
			if hashWriter.buffer.Len() > 0 {
				hash := calculateHash(hashWriter.buffer.Bytes(), config.SecretKey)
				hashWriter.header.Set("HashSHA256", hash)
			}

			// Копируем заголовки и отправляем ответ
			for key, values := range hashWriter.header {
				for _, value := range values {
					res.Header().Set(key, value)
				}
			}

			if hashWriter.wroteHeader {
				res.WriteHeader(hashWriter.statusCode)
			} else {
				res.WriteHeader(http.StatusOK)
			}
			
			if _, err := res.Write(hashWriter.buffer.Bytes()); err != nil {
				logger.Logger.Errorln("Error writing response:", err)
			}
		})
	}
}