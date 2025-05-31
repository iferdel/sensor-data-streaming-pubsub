package storage

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Relational Database
var (
	PostgresConnString = os.Getenv("POSTGRES_CONN_STRING")
)

type DB struct {
	pool *pgxpool.Pool
}

func NewDBPool(connString string) (*DB, error) {
	ctx := context.Background()

	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %v", err)
	}

	return &DB{
		pool: dbpool,
	}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}
