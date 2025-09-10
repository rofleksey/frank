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

-- name: CreateScheduledJob :one
INSERT INTO scheduled_jobs (data)
VALUES ($1)
    RETURNING id;

-- name: GetScheduledJob :one
SELECT * FROM scheduled_jobs
WHERE id = $1;

-- name: ListScheduledJobs :many
SELECT * FROM scheduled_jobs
ORDER BY id;

-- name: DeleteScheduledJob :exec
DELETE FROM scheduled_jobs
WHERE id = $1;

-- name: CountScheduledJobs :one
SELECT COUNT(*) FROM scheduled_jobs;

-- name: GetMigrations :many
SELECT *
FROM migration
ORDER BY id;

-- name: CreateMigration :one
INSERT INTO migration (id, applied)
VALUES ($1, $2) RETURNING id;
