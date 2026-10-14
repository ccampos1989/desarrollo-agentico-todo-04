ALTER TABLE tasks ADD COLUMN number integer;

WITH ordered AS (
  SELECT id, row_number() OVER (ORDER BY created_at ASC, id ASC) AS n
  FROM tasks
)
UPDATE tasks SET number = ordered.n
FROM ordered
WHERE tasks.id = ordered.id;

CREATE SEQUENCE tasks_number_seq AS integer OWNED BY tasks.number;

SELECT setval('tasks_number_seq', COALESCE((SELECT MAX(number) FROM tasks), 0) + 1, false);

ALTER TABLE tasks ALTER COLUMN number SET DEFAULT nextval('tasks_number_seq');
ALTER TABLE tasks ALTER COLUMN number SET NOT NULL;
ALTER TABLE tasks ADD CONSTRAINT tasks_number_key UNIQUE (number);
