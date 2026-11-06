ALTER TABLE todos
    ADD COLUMN recurrence TEXT NOT NULL DEFAULT 'none',
    ADD COLUMN next_run_at TIMESTAMPTZ;

ALTER TABLE todos
    ADD CONSTRAINT todos_recurrence_check
    CHECK (recurrence IN ('none', 'daily', 'weekly', 'monthly'));

CREATE INDEX idx_todos_recurrence_next_run ON todos(recurrence, next_run_at) WHERE recurrence <> 'none';

---- create above / drop below ----

DROP INDEX IF EXISTS idx_todos_recurrence_next_run;

ALTER TABLE todos
    DROP CONSTRAINT IF EXISTS todos_recurrence_check;

ALTER TABLE todos
    DROP COLUMN IF EXISTS next_run_at,
    DROP COLUMN IF EXISTS recurrence;
