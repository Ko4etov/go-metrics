// internal/server/middlewares/with_ip_check.go
package middlewares

import (
	"net"
	"net/http"
)

// IPConfig содержит конфигурацию для проверки IP
type IPConfig struct {
	TrustedNet *net.IPNet // доверенная подсеть в CIDR формате
}

// WithIPCheck возвращает middleware для проверки IP-адреса агента
func WithIPCheck(config *IPConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(ipStr)
			if ip == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			if !config.TrustedNet.Contains(ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Передаем управление дальше
			next.ServeHTTP(w, r)
		})
	}
}