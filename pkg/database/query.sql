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

-- name: GetMigrations :many
SELECT *
FROM migration
ORDER BY id;

-- name: CreateMigration :one
INSERT INTO migration (id, applied)
VALUES ($1, $2) RETURNING id;
