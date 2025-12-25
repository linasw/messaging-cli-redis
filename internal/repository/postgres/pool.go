package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func NewPool() (*pgxpool.Pool, error) {
	connString := "host=localhost port=5432 user=orderuser password=orderpass dbname=orderdb"

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, err
	}

	//test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
