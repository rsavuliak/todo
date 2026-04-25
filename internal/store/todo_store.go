package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rsavuliak/todo/internal/domain"
)

type PgTodoStore struct {
	pool *pgxpool.Pool
}

func NewPgTodoStore(pool *pgxpool.Pool) *PgTodoStore {
	return &PgTodoStore{pool: pool}
}

func (s *PgTodoStore) GetTodosByList(ctx context.Context, listID uuid.UUID) ([]domain.Todo, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, list_id, user_id, text, done, created_at FROM todos WHERE list_id = $1 ORDER BY created_at ASC`,
		listID,
	)
	if err != nil {
		return nil, fmt.Errorf("query todos: %w", err)
	}
	todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Todo])
	if err != nil {
		return nil, fmt.Errorf("collect todos: %w", err)
	}
	if todos == nil {
		todos = []domain.Todo{}
	}
	return todos, nil
}

func (s *PgTodoStore) GetTodo(ctx context.Context, id uuid.UUID) (domain.Todo, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, list_id, user_id, text, done, created_at FROM todos WHERE id = $1`,
		id,
	)
	if err != nil {
		return domain.Todo{}, fmt.Errorf("query todo: %w", err)
	}
	todo, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Todo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Todo{}, pgx.ErrNoRows
		}
		return domain.Todo{}, fmt.Errorf("collect todo: %w", err)
	}
	return todo, nil
}

func (s *PgTodoStore) CreateTodo(ctx context.Context, listID uuid.UUID, userID, text string) (domain.Todo, error) {
	rows, err := s.pool.Query(ctx,
		`INSERT INTO todos (list_id, user_id, text) VALUES ($1, $2, $3) RETURNING id, list_id, user_id, text, done, created_at`,
		listID, userID, text,
	)
	if err != nil {
		return domain.Todo{}, fmt.Errorf("insert todo: %w", err)
	}
	todo, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Todo])
	if err != nil {
		return domain.Todo{}, fmt.Errorf("collect created todo: %w", err)
	}
	return todo, nil
}

func (s *PgTodoStore) UpdateTodo(ctx context.Context, id uuid.UUID, text *string, done *bool) (domain.Todo, error) {
	rows, err := s.pool.Query(ctx,
		`UPDATE todos
		 SET text = COALESCE($1, text),
		     done = COALESCE($2, done)
		 WHERE id = $3
		 RETURNING id, list_id, user_id, text, done, created_at`,
		text, done, id,
	)
	if err != nil {
		return domain.Todo{}, fmt.Errorf("update todo: %w", err)
	}
	todo, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Todo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Todo{}, pgx.ErrNoRows
		}
		return domain.Todo{}, fmt.Errorf("collect updated todo: %w", err)
	}
	return todo, nil
}

func (s *PgTodoStore) DeleteTodo(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete todo: %w", err)
	}
	return nil
}

func (s *PgTodoStore) DeleteAllTodosInList(ctx context.Context, listID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM todos WHERE list_id = $1`, listID)
	if err != nil {
		return fmt.Errorf("delete todos in list: %w", err)
	}
	return nil
}
