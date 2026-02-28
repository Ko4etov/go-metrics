// Package config предоставляет конфигурацию для агента сбора метрик.
package config

import (
	"errors"
	"time"
)

// AgentConfig содержит конфигурационные параметры агента.
type AgentConfig struct {
	Address        string        // адрес сервера для отправки метрик
	PollInterval   time.Duration // интервал опроса метрик системы
	ReportInterval time.Duration // интервал отправки метрик на сервер
	HashKey        string        // ключ для хеширования (опционально)
	RateLimit      int           // лимит одновременных запросов
	CryptoKey      string        // Файл крипто ключа
	UseGRPC        bool
	GRPCAddress    string
}

// New создает новую конфигурацию агента.
func New() (*AgentConfig, error) {
	parameters := parseAgentParameters()

	if parameters.UseGRPC {
		if parameters.GRPCAddress == "" {
			return nil, ErrGRPCAddressMissed
		}
	}

	return &AgentConfig{
		Address:        parameters.Address,
		PollInterval:   time.Duration(parameters.PollInterval) * time.Second,
		ReportInterval: time.Duration(parameters.ReportInterval) * time.Second,
		HashKey:        parameters.HashKey,
		RateLimit:      parameters.RateLimit,
		CryptoKey:      parameters.CryptoKey,
		UseGRPC:        parameters.UseGRPC,
		GRPCAddress:    parameters.GRPCAddress,
	}, nil
}

var (
	ErrGRPCAddressMissed = errors.New("grpc address missed error")
)
