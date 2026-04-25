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

type PgListStore struct {
	pool *pgxpool.Pool
}

func NewPgListStore(pool *pgxpool.Pool) *PgListStore {
	return &PgListStore{pool: pool}
}

func (s *PgListStore) GetListsByUser(ctx context.Context, userID string) ([]domain.List, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, name, created_at FROM lists WHERE user_id = $1 ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query lists: %w", err)
	}
	lists, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.List])
	if err != nil {
		return nil, fmt.Errorf("collect lists: %w", err)
	}
	if lists == nil {
		lists = []domain.List{}
	}
	return lists, nil
}

func (s *PgListStore) GetList(ctx context.Context, id uuid.UUID) (domain.List, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, name, created_at FROM lists WHERE id = $1`,
		id,
	)
	if err != nil {
		return domain.List{}, fmt.Errorf("query list: %w", err)
	}
	list, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.List])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.List{}, pgx.ErrNoRows
		}
		return domain.List{}, fmt.Errorf("collect list: %w", err)
	}
	return list, nil
}

func (s *PgListStore) CreateList(ctx context.Context, userID, name string) (domain.List, error) {
	rows, err := s.pool.Query(ctx,
		`INSERT INTO lists (user_id, name) VALUES ($1, $2) RETURNING id, user_id, name, created_at`,
		userID, name,
	)
	if err != nil {
		return domain.List{}, fmt.Errorf("insert list: %w", err)
	}
	list, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.List])
	if err != nil {
		return domain.List{}, fmt.Errorf("collect created list: %w", err)
	}
	return list, nil
}

func (s *PgListStore) UpdateListName(ctx context.Context, id uuid.UUID, name string) (domain.List, error) {
	rows, err := s.pool.Query(ctx,
		`UPDATE lists SET name = $1 WHERE id = $2 RETURNING id, user_id, name, created_at`,
		name, id,
	)
	if err != nil {
		return domain.List{}, fmt.Errorf("update list: %w", err)
	}
	list, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.List])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.List{}, pgx.ErrNoRows
		}
		return domain.List{}, fmt.Errorf("collect updated list: %w", err)
	}
	return list, nil
}

func (s *PgListStore) DeleteList(ctx context.Context, id uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM lists WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete list: %w", err)
	}
	return nil
}
