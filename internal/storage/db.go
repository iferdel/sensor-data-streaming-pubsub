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

func (db *DB) Close() {
	db.dbpool.Close()
}
