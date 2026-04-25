package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rsavuliak/todo/internal/domain"
)

type ListStore interface {
	GetListsByUser(ctx context.Context, userID string) ([]domain.List, error)
	GetList(ctx context.Context, id uuid.UUID) (domain.List, error)
	CreateList(ctx context.Context, userID, name string) (domain.List, error)
	UpdateListName(ctx context.Context, id uuid.UUID, name string) (domain.List, error)
	DeleteList(ctx context.Context, id uuid.UUID) error
}

type TodoStore interface {
	GetTodosByList(ctx context.Context, listID uuid.UUID) ([]domain.Todo, error)
	GetTodo(ctx context.Context, id uuid.UUID) (domain.Todo, error)
	CreateTodo(ctx context.Context, listID uuid.UUID, userID, text string) (domain.Todo, error)
	UpdateTodo(ctx context.Context, id uuid.UUID, text *string, done *bool) (domain.Todo, error)
	DeleteTodo(ctx context.Context, id uuid.UUID) error
	DeleteAllTodosInList(ctx context.Context, listID uuid.UUID) error
}

func OpenPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}
