// internal/server/middlewares/with_ip_check.go
package middlewares

import (
	"net"
	"net/http"

	"github.com/Ko4etov/go-metrics/internal/server/service/logger"
)

// IPConfig содержит конфигурацию для проверки IP
type IPConfig struct {
	TrustedSubnet string // доверенная подсеть в CIDR формате
}

// WithIPCheck возвращает middleware для проверки IP-адреса агента
func WithIPCheck(config *IPConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если доверенная подсеть не указана - пропускаем без проверки
			if config.TrustedSubnet == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Парсим доверенную подсеть
			_, trustedNet, err := net.ParseCIDR(config.TrustedSubnet)
			if err != nil {
				logger.Logger.Errorf("Invalid trusted subnet format: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Получаем IP из заголовка X-Real-IP
			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				logger.Logger.Warn("X-Real-IP header is missing")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Парсим IP адрес
			ip := net.ParseIP(ipStr)
			if ip == nil {
				logger.Logger.Warnf("Invalid IP address format: %s", ipStr)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Проверяем, входит ли IP в доверенную подсеть
			if !trustedNet.Contains(ip) {
				logger.Logger.Warnf("IP %s is not in trusted subnet %s", ipStr, config.TrustedSubnet)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			logger.Logger.Debugf("IP %s is allowed (trusted subnet: %s)", ipStr, config.TrustedSubnet)

			// Передаем управление дальше
			next.ServeHTTP(w, r)
		})
	}
}

// GetRealIP возвращает реальный IP адрес из заголовка X-Real-IP
func GetRealIP(r *http.Request) string {
	return r.Header.Get("X-Real-IP")
}