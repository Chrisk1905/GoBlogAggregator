-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES(
        $1, $2, $3, $4, $5
    )
    RETURNING *
)
SELECT 
    inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds
ON inserted_feed_follow.feed_id = feeds.id
INNER JOIN users
ON inserted_feed_follow.user_id = users.id;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*, users.name as user_name, feeds.name as feeds_name FROM feed_follows 
INNER JOIN users
ON users.id = feed_follows.user_id
INNER JOIN feeds
ON feed_follows.feed_id = feeds.id
WHERE $1 = users.id;

-- name: DeleteFeedFollowsByUser :exec
DELETE FROM feed_follows
USING feeds
WHERE @user_id = feed_follows.user_id
AND feed_follows.feed_id = feeds.id
AND @feed_url = feeds.url;