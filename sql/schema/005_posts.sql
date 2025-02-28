-- +goose Up
CREATE TABLE posts (
    "id" UUID NOT NULL PRIMARY KEY,
    "created_at" TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP NOT NULL,
    "title" VARCHAR(100) NOT NULL,
    "url" VARCHAR(200) NOT NULL,
    "description" VARCHAR(500) NOT NULL,
    "published_at" TIMESTAMP NOT NULL,
    "feed_id" UUID NOT NULL,
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;

