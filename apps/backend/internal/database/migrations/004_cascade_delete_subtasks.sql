ALTER TABLE todos DROP CONSTRAINT IF EXISTS todos_parent_todo_id_fkey;
ALTER TABLE todos
    ADD CONSTRAINT todos_parent_todo_id_fkey
    FOREIGN KEY (parent_todo_id)
    REFERENCES todos(id)
    ON DELETE CASCADE;

---- create above / drop below ----

ALTER TABLE todos DROP CONSTRAINT IF EXISTS todos_parent_todo_id_fkey;
ALTER TABLE todos
    ADD CONSTRAINT todos_parent_todo_id_fkey
    FOREIGN KEY (parent_todo_id)
    REFERENCES todos(id);
