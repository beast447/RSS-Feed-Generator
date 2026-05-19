-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES(
  $1,
  $2,
  $3,
  $4,
  $5,
  $6
)
RETURNING *;

-- name: ListAllFeeds :many
SELECT * FROM feeds;

-- name: GetFeedsWithUsers :many
SELECT feeds.url, feeds.name, users.name
FROM users
INNER JOIN feeds
ON users.id = feeds.user_id;

-- name: GetFeedByUrl :one
SELECT feeds.id
FROM feeds
WHERE feeds.url = $1;

