package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/rsavuliak/todo/internal/store"
)

func TestListStore_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	s := store.NewPgListStore(testPool)

	list, err := s.CreateList(ctx, "user-1", "My List")
	if err != nil {
		t.Fatalf("create list: %v", err)
	}
	if list.Name != "My List" {
		t.Errorf("expected name 'My List', got %q", list.Name)
	}

	got, err := s.GetList(ctx, list.ID)
	if err != nil {
		t.Fatalf("get list: %v", err)
	}
	if got.ID != list.ID {
		t.Errorf("id mismatch")
	}
}

func TestListStore_GetListsByUser_OwnershipIsolation(t *testing.T) {
	ctx := context.Background()
	s := store.NewPgListStore(testPool)

	_, _ = s.CreateList(ctx, "user-a", "A's list")
	_, _ = s.CreateList(ctx, "user-b", "B's list")

	lists, err := s.GetListsByUser(ctx, "user-a")
	if err != nil {
		t.Fatalf("get lists: %v", err)
	}
	for _, l := range lists {
		if l.UserID != "user-a" {
			t.Errorf("got list belonging to %q, expected user-a only", l.UserID)
		}
	}
}

func TestListStore_GetList_NotFound(t *testing.T) {
	ctx := context.Background()
	s := store.NewPgListStore(testPool)

	nonExistent := mustParseUUID("00000000-0000-0000-0000-000000000001")
	_, err := s.GetList(ctx, nonExistent)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestListStore_UpdateAndDelete(t *testing.T) {
	ctx := context.Background()
	s := store.NewPgListStore(testPool)

	list, _ := s.CreateList(ctx, "user-update", "Old Name")

	updated, err := s.UpdateListName(ctx, list.ID, "New Name")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected 'New Name', got %q", updated.Name)
	}

	if err := s.DeleteList(ctx, list.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = s.GetList(ctx, list.ID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("expected not found after delete, got %v", err)
	}
}
