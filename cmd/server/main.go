package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rsavuliak/todo/internal/config"
	"github.com/rsavuliak/todo/internal/handler"
	"github.com/rsavuliak/todo/internal/middleware"
	"github.com/rsavuliak/todo/internal/store"
	migrationsfs "github.com/rsavuliak/todo/migrations"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := store.OpenPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	runMigrations(cfg.DatabaseURL)

	listStore := store.NewPgListStore(pool)
	todoStore := store.NewPgTodoStore(pool)
	listHandler := handler.NewListHandler(listStore)
	todoHandler := handler.NewTodoHandler(listStore, todoStore)

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.CORS(cfg.CORSOrigins))
	r.Use(middleware.MaxBodySize(1 << 20)) // 1 MB

	r.Get("/health", handler.Health)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(cfg))

		r.Route("/api/v1/lists", func(r chi.Router) {
			r.Get("/", listHandler.GetLists)
			r.Post("/", listHandler.CreateList)
			r.Patch("/{id}", listHandler.UpdateList)
			r.Delete("/{id}", listHandler.DeleteList)

			r.Route("/{listId}/todos", func(r chi.Router) {
				r.Get("/", todoHandler.GetTodos)
				r.Post("/", todoHandler.CreateTodo)
				r.Delete("/", todoHandler.DeleteAllTodos)
				r.Patch("/{id}", todoHandler.UpdateTodo)
				r.Delete("/{id}", todoHandler.DeleteTodo)
			})
		})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	<-quit
	slog.Info("shutting down")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		slog.Error("shutdown", "err", err)
	}
}

func runMigrations(databaseURL string) {
	pgx5URL, err := toPgx5URL(databaseURL)
	if err != nil {
		slog.Error("migrate url", "err", err)
		os.Exit(1)
	}

	d, err := iofs.New(migrationsfs.FS, ".")
	if err != nil {
		slog.Error("migrate source", "err", err)
		os.Exit(1)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, pgx5URL)
	if err != nil {
		slog.Error("migrate init", "err", err)
		os.Exit(1)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			slog.Warn("migrate source close", "err", srcErr)
		}
		if dbErr != nil {
			slog.Warn("migrate db close", "err", dbErr)
		}
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		slog.Error("migrate up", "err", err)
		os.Exit(1)
	}
	slog.Info("migrations applied")
}

// toPgx5URL converts a postgres:// or postgresql:// URL to pgx5:// as
// required by the golang-migrate pgx/v5 database driver.
func toPgx5URL(databaseURL string) (string, error) {
	for _, prefix := range []string{"postgres://", "postgresql://"} {
		if rest, ok := strings.CutPrefix(databaseURL, prefix); ok {
			return "pgx5://" + rest, nil
		}
	}
	return "", fmt.Errorf("unsupported DATABASE_URL scheme (expected postgres:// or postgresql://)")
}
