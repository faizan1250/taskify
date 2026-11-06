ALTER TABLE todos
    ADD COLUMN position DOUBLE PRECISION NOT NULL DEFAULT 0;

CREATE INDEX idx_todos_user_position ON todos(user_id, position);

---- create above / drop below ----

DROP INDEX IF EXISTS idx_todos_user_position;

ALTER TABLE todos
    DROP COLUMN IF EXISTS position;
