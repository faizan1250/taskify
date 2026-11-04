ALTER TABLE todos
    ADD COLUMN reminder_sent_at TIMESTAMPTZ,
    ADD COLUMN overdue_notified_at TIMESTAMPTZ;

CREATE INDEX idx_todos_reminder_sent_at ON todos(reminder_sent_at);
CREATE INDEX idx_todos_overdue_notified_at ON todos(overdue_notified_at);

---- create above / drop below ----

DROP INDEX IF EXISTS idx_todos_overdue_notified_at;
DROP INDEX IF EXISTS idx_todos_reminder_sent_at;

ALTER TABLE todos
    DROP COLUMN IF EXISTS overdue_notified_at,
    DROP COLUMN IF EXISTS reminder_sent_at;
