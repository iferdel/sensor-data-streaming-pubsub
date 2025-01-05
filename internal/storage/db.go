package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Relational Database
const (
	PostgresConnString = "postgres://postgres:postgres@localhost:5432/iot"
)

// TODO: Single Pool instead of one per function
type DB struct {
	dbpool *pgxpool.Pool
}

func NewDBPool(connString string) (*DB, error) {
	ctx := context.Background()

	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("Unable to create connection pool: %v", err)
	}

	return &DB{
		dbpool: dbpool,
	}, nil
}

func (s *DB) Close() {
	s.dbpool.Close()
}
