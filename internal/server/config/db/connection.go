package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDBConnection(address string) *pgxpool.Pool {
	pool, err := pgxpool.New(
		context.Background(),
		address)

	if err != nil {
		panic(err)
	}

	return pool
}