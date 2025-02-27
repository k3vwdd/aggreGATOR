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
-- name: GetFeeds :many
SELECT
    feeds.name AS Feeds_Name,
    feeds.url,
    users.name AS Users_Name
FROM feeds
JOIN users ON feeds.user_id = users.id;
-- name: CreateFeedFollow :many
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (user_id, feed_id, created_at, updated_at)  -- Insert a new follow record
    VALUES ($1, $2, $3, $4)  -- Assuming placeholders for user_id and feed_id
    RETURNING *  -- Return the inserted row
)
SELECT
    inserted_feed_follow.*,
    feeds.name AS feed_name,  -- Fetch feed name
    users.name AS user_name   -- Fetch user name
FROM inserted_feed_follow
INNER JOIN feeds ON inserted_feed_follow.feed_id = feeds.id  -- Join with feeds to get feed name
INNER JOIN users ON inserted_feed_follow.user_id = users.id;  -- Join with users to get user name
-- name: GetFeedByURL :one
SELECT *
FROM feeds
WHERE url = $1;
-- name: GetFeedFollowsForUser :many
SELECT
    feed_follows.id AS follow_id,  -- ID of the follow relationship
    feed_follows.created_at,  -- When the user followed the feed
    feeds.id AS feed_id,  -- ID of the followed feed
    feeds.name AS feed_name,  -- Name of the followed feed
    feeds.url AS feed_url  -- URL of the followed feed
FROM feed_follows
INNER JOIN feeds ON feed_follows.feed_id = feeds.id  -- Join feeds to get feed details
WHERE feed_follows.user_id = $1;  -- Filter by the specific user_id
-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
USING feeds
WHERE feed_follows.feed_id = feeds.id  -- Match the joined row
  AND feed_follows.user_id = $1        -- The user's ID
  AND feeds.url = $2;                  -- The feed's URL

