package store_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rsavuliak/todo/internal/store"
)

func mustParseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return id
}

func TestTodoStore_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	ls := store.NewPgListStore(testPool)
	ts := store.NewPgTodoStore(testPool)

	list, _ := ls.CreateList(ctx, "user-todo-1", "Test List")

	todo, err := ts.CreateTodo(ctx, list.ID, "user-todo-1", "Buy milk")
	if err != nil {
		t.Fatalf("create todo: %v", err)
	}
	if todo.Text != "Buy milk" || todo.Done {
		t.Errorf("unexpected todo state: %+v", todo)
	}

	got, err := ts.GetTodo(ctx, todo.ID)
	if err != nil {
		t.Fatalf("get todo: %v", err)
	}
	if got.ID != todo.ID {
		t.Error("id mismatch")
	}
}

func TestTodoStore_Update_PartialFields(t *testing.T) {
	ctx := context.Background()
	ls := store.NewPgListStore(testPool)
	ts := store.NewPgTodoStore(testPool)

	list, _ := ls.CreateList(ctx, "user-todo-2", "Test List 2")
	todo, _ := ts.CreateTodo(ctx, list.ID, "user-todo-2", "Original")

	newText := "Updated"
	updated, err := ts.UpdateTodo(ctx, todo.ID, &newText, nil)
	if err != nil {
		t.Fatalf("update text: %v", err)
	}
	if updated.Text != "Updated" || updated.Done {
		t.Errorf("unexpected state after text update: %+v", updated)
	}

	done := true
	updated, err = ts.UpdateTodo(ctx, todo.ID, nil, &done)
	if err != nil {
		t.Fatalf("update done: %v", err)
	}
	if !updated.Done {
		t.Error("expected done=true")
	}
}

func TestTodoStore_CascadeDelete(t *testing.T) {
	ctx := context.Background()
	ls := store.NewPgListStore(testPool)
	ts := store.NewPgTodoStore(testPool)

	list, _ := ls.CreateList(ctx, "user-cascade", "Cascade List")
	todo, _ := ts.CreateTodo(ctx, list.ID, "user-cascade", "Will be cascade deleted")

	if err := ls.DeleteList(ctx, list.ID); err != nil {
		t.Fatalf("delete list: %v", err)
	}
	_, err := ts.GetTodo(ctx, todo.ID)
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("todo should be cascade deleted, got %v", err)
	}
}

func TestTodoStore_DeleteAllInList(t *testing.T) {
	ctx := context.Background()
	ls := store.NewPgListStore(testPool)
	ts := store.NewPgTodoStore(testPool)

	list, _ := ls.CreateList(ctx, "user-wipe", "Wipe List")
	_, _ = ts.CreateTodo(ctx, list.ID, "user-wipe", "Todo 1")
	_, _ = ts.CreateTodo(ctx, list.ID, "user-wipe", "Todo 2")

	if err := ts.DeleteAllTodosInList(ctx, list.ID); err != nil {
		t.Fatalf("delete all: %v", err)
	}
	todos, err := ts.GetTodosByList(ctx, list.ID)
	if err != nil {
		t.Fatalf("get todos: %v", err)
	}
	if len(todos) != 0 {
		t.Errorf("expected 0 todos after wipe, got %d", len(todos))
	}
}
