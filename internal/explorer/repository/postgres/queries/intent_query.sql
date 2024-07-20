-- name: SaveIntent :exec
INSERT INTO intents (id, repository, since, created_at, is_active)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE
SET repository = EXCLUDED.repository,
    since = EXCLUDED.since,
    created_at = EXCLUDED.created_at,
    is_active = EXCLUDED.is_active;

-- name: GetIntentById :one
SELECT id, repository, since, created_at, is_active
FROM intents
WHERE id = $1;

-- name: GetIntentByRepoName :one
SELECT id, repository, since, created_at, is_active
FROM intents
WHERE repository = $1;

-- name: GetIntents :many
SELECT id, repository, since, created_at, is_active 
FROM intents 
WHERE is_active = COALESCE($1, is_active);

-- name: UpdateIntent :exec
UPDATE intents
SET is_active = COALESCE($2, is_active),
    since = COALESCE($3, since)
WHERE id = $1;
