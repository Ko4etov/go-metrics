package handler

import (
	"github.com/Ko4etov/go-metrics/internal/server/interfaces"
)

type Handler struct {
	storage interfaces.Storage
}

func New(s interfaces.Storage) *Handler {
	return &Handler {
		storage: s,
	}
}