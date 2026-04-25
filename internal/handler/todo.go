package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rsavuliak/todo/internal/middleware"
	"github.com/rsavuliak/todo/internal/store"
)

type TodoHandler struct {
	lists store.ListStore
	todos store.TodoStore
}

func NewTodoHandler(lists store.ListStore, todos store.TodoStore) *TodoHandler {
	return &TodoHandler{lists: lists, todos: todos}
}

// ownerList parses the listId URL param and verifies the list belongs to userID.
// Writes the appropriate error response and returns false if ownership fails.
func (h *TodoHandler) ownerList(w http.ResponseWriter, r *http.Request, userID string) (uuid.UUID, bool) {
	listID, err := uuid.Parse(chi.URLParam(r, "listId"))
	if err != nil {
		NotFound(w)
		return uuid.Nil, false
	}
	list, err := h.lists.GetList(r.Context(), listID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && list.UserID != userID) {
		NotFound(w)
		return uuid.Nil, false
	}
	if err != nil {
		InternalError(w, err)
		return uuid.Nil, false
	}
	return listID, true
}

func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		InternalError(w, fmt.Errorf("auth middleware not active"))
		return
	}

	listID, ok := h.ownerList(w, r, userID)
	if !ok {
		return
	}

	todos, err := h.todos.GetTodosByList(r.Context(), listID)
	if err != nil {
		InternalError(w, err)
		return
	}
	JSON(w, http.StatusOK, todos)
}

func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		InternalError(w, fmt.Errorf("auth middleware not active"))
		return
	}

	listID, ok := h.ownerList(w, r, userID)
	if !ok {
		return
	}

	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Errors(w, http.StatusBadRequest, []string{"body: malformed"})
		return
	}
	if !ValidateRequired(w, "text", body.Text) {
		return
	}

	todo, err := h.todos.CreateTodo(r.Context(), listID, userID, body.Text)
	if err != nil {
		InternalError(w, err)
		return
	}
	JSON(w, http.StatusCreated, todo)
}

func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		InternalError(w, fmt.Errorf("auth middleware not active"))
		return
	}

	listID, ok := h.ownerList(w, r, userID)
	if !ok {
		return
	}

	todoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		NotFound(w)
		return
	}

	existing, err := h.todos.GetTodo(r.Context(), todoID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && (existing.UserID != userID || existing.ListID != listID)) {
		NotFound(w)
		return
	}
	if err != nil {
		InternalError(w, err)
		return
	}

	var body struct {
		Text *string `json:"text"`
		Done *bool   `json:"done"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Errors(w, http.StatusBadRequest, []string{"body: malformed"})
		return
	}
	if body.Text == nil && body.Done == nil {
		Errors(w, http.StatusBadRequest, []string{"at least one of text or done is required"})
		return
	}

	updated, err := h.todos.UpdateTodo(r.Context(), todoID, body.Text, body.Done)
	if err != nil {
		InternalError(w, err)
		return
	}
	JSON(w, http.StatusOK, updated)
}

func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		InternalError(w, fmt.Errorf("auth middleware not active"))
		return
	}

	listID, ok := h.ownerList(w, r, userID)
	if !ok {
		return
	}

	todoID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		NotFound(w)
		return
	}

	existing, err := h.todos.GetTodo(r.Context(), todoID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && (existing.UserID != userID || existing.ListID != listID)) {
		NotFound(w)
		return
	}
	if err != nil {
		InternalError(w, err)
		return
	}

	if err := h.todos.DeleteTodo(r.Context(), todoID); err != nil {
		InternalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TodoHandler) DeleteAllTodos(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		InternalError(w, fmt.Errorf("auth middleware not active"))
		return
	}

	listID, ok := h.ownerList(w, r, userID)
	if !ok {
		return
	}

	if err := h.todos.DeleteAllTodosInList(r.Context(), listID); err != nil {
		InternalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
