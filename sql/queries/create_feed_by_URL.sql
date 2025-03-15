-- name: GetFeedByURL :one
SELECT id, created_at, updated_at, name, url, user_id, last_fetched_at
FROM feeds
WHERE url = $1
LIMIT 1;
