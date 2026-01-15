package retriableagent

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

type RetriableAgent struct {
	MaxRetries    int
	RetryDelays   []time.Duration
}

func New(retries int, retryDelays []time.Duration) *RetriableAgent {
	return &RetriableAgent {
		MaxRetries:    retries,
		RetryDelays:   retryDelays,
	}
}

func (r *RetriableAgent) Send(operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= r.MaxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		if !r.isRetriableError(err) {
			return fmt.Errorf("non-retriable error: %w", err)
		}

		if attempt < r.MaxRetries {
			delay := r.RetryDelays[attempt]
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("failed after %d retries: %w", r.MaxRetries, lastErr)
}

func (r *RetriableAgent) isRetriableError(err error) bool {
	if err == nil {
		return false
	}

	if r.isNetworkError(err) {
		return true
	}

	if r.isRetriableByContent(err) {
		return true
	}

	if r.isRetriableHTTPStatus(err) {
		return true
	}

	return false
}

// isNetworkError проверяет сетевые ошибки.
func (r *RetriableAgent) isNetworkError(err error) bool {
	switch e := err.(type) {
	case *url.Error:
		return true
	case net.Error:
		return e.Timeout()
	}
	return false
}

// isRetriableByContent проверяет ошибки по их текстовому содержимому.
func (r *RetriableAgent) isRetriableByContent(err error) bool {
	errorStr := strings.ToLower(err.Error())

	retriablePatterns := []string{
		"timeout", "connection refused", "connection reset",
		"network", "temporary", "unavailable", "dial tcp",
		"no such host", "EOF", "broken pipe", "connection aborted",
		"i/o timeout", "network is unreachable", "reset by peer",
		"service unavailable", "bad gateway", "gateway timeout",
	}

	for _, pattern := range retriablePatterns {
		if strings.Contains(errorStr, pattern) {
			return true
		}
	}

	return false
}

func (r *RetriableAgent) isRetriableHTTPStatus(err error) bool {
	errorStr := strings.ToLower(err.Error())

	return strings.Contains(errorStr, "server error: 400")
}