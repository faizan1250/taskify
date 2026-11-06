CREATE TABLE todo_share_tokens (
    token TEXT PRIMARY KEY,
    todo_id UUID NOT NULL REFERENCES todos ON DELETE CASCADE,
    created_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_todo_share_tokens_todo_id ON todo_share_tokens(todo_id);

---- create above / drop below ----

DROP INDEX IF EXISTS idx_todo_share_tokens_todo_id;

DROP TABLE IF EXISTS todo_share_tokens;
