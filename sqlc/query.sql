-- name: GetArticles :many
SELECT * FROM articles
ORDER BY date ASC;

-- name: GetArticle :one
SELECT * FROM articles
WHERE article_id = ?
LIMIT 1;

-- name: GetNewestArticle :one
SELECT * FROM articles
ORDER BY date DESC
LIMIT 1;

-- name: CreateArticle :execresult
INSERT INTO articles(title, content, date)
VALUES (?, ?, ?);

-- name: EditArticle :execresult
UPDATE articles
SET 
    title = ?,
    content = ?
WHERE article_id = ?;

-- name: DeleteArticle :execresult
DELETE FROM articles
WHERE article_id = ?;

-- name: GetUser :one
SELECT * FROM users
WHERE user_id = ?
LIMIT 1;

-- name: GetUserName :one
SELECT * FROM users
WHERE name = ?
LIMIT 1;

-- name: CreateUser :execresult
INSERT INTO users(name, password)
VALUES(?, ?);












