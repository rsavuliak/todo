CREATE TABLE todos (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    list_id    UUID        NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    user_id    TEXT        NOT NULL,
    text       TEXT        NOT NULL,
    done       BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_todos_list_id ON todos(list_id);
CREATE INDEX idx_todos_user_id ON todos(user_id);
