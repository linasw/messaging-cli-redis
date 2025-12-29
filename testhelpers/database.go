package testhelpers

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"testing"
)

func SetupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	connString := "host=localhost port=5432 user=orderuser password=orderpass dbname=orderdb"

	poolConfig, err := pgxpool.ParseConfig(connString)
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	require.NoError(t, err, "failed to create database pool")

	_, err = pool.Exec(context.Background(), "DELETE FROM orders")
	require.NoError(t, err, "failed to clean orders table")

	return pool
}
