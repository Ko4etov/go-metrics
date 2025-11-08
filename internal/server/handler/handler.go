package handler

import (
	"github.com/Ko4etov/go-metrics/internal/server/interfaces"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	storage interfaces.Storage
	pgx *pgxpool.Pool
}

func New(s interfaces.Storage, pgx *pgxpool.Pool) *Handler {
	return &Handler {
		storage: s,
		pgx: pgx,
	}
}