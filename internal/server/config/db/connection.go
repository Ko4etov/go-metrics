package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewDbConnection(address string) *pgxpool.Pool {
	pool, err := pgxpool.New(
		context.Background(), 
		fmt.Sprintf("user=%s port=%s dbname=%s host=%s password=%s", "metrics", address, "metrics", "localhost", "v8Te8krwy4uIDBF7"))

	if err != nil {
		panic(err)
	}

	return pool
}