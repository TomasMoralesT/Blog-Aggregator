-- name: GetFeedFollowsForUser :many
SELECT
    feed_follows.id AS feed_follow_id,
    feed_follows.created_at,
    feed_follows.updated_at,
    users.name AS user_name,
    feeds.name AS feed_name,
    feeds.url AS feed_url
FROM feed_follows
INNER JOIN users ON feed_follows.user_id = users.id
INNER JOIN feeds ON feed_follows.feed_id = feeds.id
WHERE feed_follows.user_id = $1;
