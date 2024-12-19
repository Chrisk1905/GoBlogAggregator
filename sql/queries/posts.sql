-- name: CreatePosts :one
INSERT INTO posts(id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES(
    @id,
    @created_at,
    @updated_at,
    @title,
    @url,
    @description,
    @published_at,
    @feed_id
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.title, posts.description, posts.published_at, posts.url, feeds.name FROM posts
INNER JOIN feeds
ON feeds.id = feed_id 
WHERE feeds.user_id = $1
ORDER BY published_at DESC 
LIMIT $2;