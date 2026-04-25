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

type ListHandler struct {
	lists store.ListStore
}

func NewListHandler(lists store.ListStore) *ListHandler {
	return &ListHandler{lists: lists}
}

func (h *ListHandler) GetLists(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		InternalError(w, fmt.Errorf("auth middleware not active"))
		return
	}

	lists, err := h.lists.GetListsByUser(r.Context(), userID)
	if err != nil {
		InternalError(w, err)
		return
	}
	JSON(w, http.StatusOK, lists)
}

func (h *ListHandler) CreateList(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		InternalError(w, fmt.Errorf("auth middleware not active"))
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Errors(w, http.StatusBadRequest, []string{"body: malformed"})
		return
	}
	if !ValidateRequired(w, "name", body.Name) {
		return
	}

	list, err := h.lists.CreateList(r.Context(), userID, body.Name)
	if err != nil {
		InternalError(w, err)
		return
	}
	JSON(w, http.StatusCreated, list)
}

func (h *ListHandler) UpdateList(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		InternalError(w, fmt.Errorf("auth middleware not active"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		NotFound(w)
		return
	}

	list, err := h.lists.GetList(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && list.UserID != userID) {
		NotFound(w)
		return
	}
	if err != nil {
		InternalError(w, err)
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		Errors(w, http.StatusBadRequest, []string{"body: malformed"})
		return
	}
	if !ValidateRequired(w, "name", body.Name) {
		return
	}

	updated, err := h.lists.UpdateListName(r.Context(), id, body.Name)
	if err != nil {
		InternalError(w, err)
		return
	}
	JSON(w, http.StatusOK, updated)
}

func (h *ListHandler) DeleteList(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		InternalError(w, fmt.Errorf("auth middleware not active"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		NotFound(w)
		return
	}

	list, err := h.lists.GetList(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && list.UserID != userID) {
		NotFound(w)
		return
	}
	if err != nil {
		InternalError(w, err)
		return
	}

	if err := h.lists.DeleteList(r.Context(), id); err != nil {
		InternalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
