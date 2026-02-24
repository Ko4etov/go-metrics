package middlewares

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/service/crypto"
)

// CryptoConfig содержит конфигурацию для расшифровки.
type CryptoConfig struct {
	CryptoKey string // Файл с ключом для расшифровки
}

// shouldDecrypt проверяет, нужно ли расшифровывать запрос.
func shouldDecrypt(req *http.Request) bool {
	if (req.Method != http.MethodPost && req.Method != http.MethodPut) ||
		req.Body == nil || req.Body == http.NoBody {
		return false
	}

	return true
}

// WithDecryption возвращает middleware для расшифровки запросов.
func WithDecryption(config *CryptoConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			privateKey, err := crypto.LoadPrivateKey(config.CryptoKey)
			if err != nil {
				log.Fatal("Failed to load private key:", err)
			}

			if !shouldDecrypt(req) {
				next.ServeHTTP(res, req)
				return
			}

			originalContentType := req.Header.Get("Content-Type")

			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				http.Error(res, "Error reading request body", http.StatusBadRequest)
				return
			}
			defer req.Body.Close()

			if len(bodyBytes) == 0 {
				next.ServeHTTP(res, req)
				return
			}

			decryptedBytes, err := crypto.Decrypt(bodyBytes, privateKey)
			if err != nil {
				http.Error(res, "Decryption failed", http.StatusBadRequest)
				return
			}

			req.Body = io.NopCloser(bytes.NewReader(decryptedBytes))

			req.ContentLength = int64(len(decryptedBytes))

			req.Header.Set("Content-Type", "application/json")

			if originalContentType == "application/octet-stream" {
				req.Header.Set("Content-Encoding", "gzip")
			}

			next.ServeHTTP(res, req)
		})
	}
}
