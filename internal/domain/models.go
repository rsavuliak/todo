package domain

import (
	"time"

	"github.com/google/uuid"
)

type List struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	UserID    string    `db:"user_id"    json:"userId"`
	Name      string    `db:"name"       json:"name"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type Todo struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	ListID    uuid.UUID `db:"list_id"    json:"listId"`
	UserID    string    `db:"user_id"    json:"userId"`
	Text      string    `db:"text"       json:"text"`
	Done      bool      `db:"done"       json:"done"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}
