package store_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	migrationsfs "github.com/rsavuliak/todo/migrations"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := postgres.Run(ctx, "postgres:16",
		postgres.WithDatabase("todo_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		panic("start postgres container: " + err.Error())
	}

	connStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic("get connection string: " + err.Error())
	}

	d, err := iofs.New(migrationsfs.FS, ".")
	if err != nil {
		panic("create migration source: " + err.Error())
	}

	pgx5URL := "pgx5://" + strings.TrimPrefix(connStr, "postgres://")
	mig, err := migrate.NewWithSourceInstance("iofs", d, pgx5URL)
	if err != nil {
		panic("create migrator: " + err.Error())
	}
	if err := mig.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic("run migrations: " + err.Error())
	}
	mig.Close() //nolint:errcheck

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		panic("create pool: " + err.Error())
	}

	code := m.Run()

	testPool.Close()
	pg.Terminate(ctx) //nolint:errcheck

	os.Exit(code)
}
