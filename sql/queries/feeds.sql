-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;


-- name: GetFeedByID :one
SELECT * FROM feeds
WHERE feeds.id = $1;

-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE feeds.url = $1;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: MarkFeedFetched :exec
UPDATE feeds SET last_fetched_at = @time, updated_at = @time WHERE id = @id;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST;