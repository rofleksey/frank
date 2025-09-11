-- name: CreateContextEntry :one
INSERT INTO context_entries (created, tags, text)
VALUES ($1, $2, $3) RETURNING id;

-- name: GetContextEntry :one
SELECT *
FROM context_entries
WHERE id = $1;

-- name: ListContextEntries :many
SELECT *
FROM context_entries
ORDER BY created DESC;

-- name: ListContextEntriesByTags :many
SELECT *
FROM context_entries
WHERE tags @> $1
ORDER BY created DESC;

-- name: ListContextEntriesByAnyTag :many
SELECT *
FROM context_entries
WHERE tags && $1
ORDER BY created DESC;

-- name: DeleteContextEntry :exec
DELETE
FROM context_entries
WHERE id = $1;

-- name: CreateScheduledJob :exec
INSERT INTO scheduled_jobs (name, created, data)
VALUES ($1, $2, $3);

-- name: GetScheduledJob :one
SELECT * FROM scheduled_jobs
WHERE name = $1;

-- name: ListScheduledJobs :many
SELECT * FROM scheduled_jobs
ORDER BY created DESC;

-- name: DeleteScheduledJob :exec
DELETE FROM scheduled_jobs
WHERE name = $1;

-- name: CountScheduledJobs :one
SELECT COUNT(*) FROM scheduled_jobs;

-- name: CreatePrompt :one
INSERT INTO prompts (created, data)
VALUES ($1, $2) RETURNING id;

-- name: GetPrompt :one
SELECT *
FROM prompts
WHERE id = $1;

-- name: UpdatePrompt :exec
UPDATE prompts
SET data = $2
WHERE id = $1;

-- name: GetMigrations :many
SELECT *
FROM migration
ORDER BY id;

-- name: CreateMigration :one
INSERT INTO migration (id, applied)
VALUES ($1, $2) RETURNING id;
