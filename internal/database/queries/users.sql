
-- name: GetUserByNormalizedEmail :one
SELECT * FROM users
WHERE normalized_email = $1;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: InsertUser :exec
INSERT INTO users (
    id,
    full_name,
    user_name,
    normalized_user_name,
    email,
    email_confirmed,
    normalized_email,
    password_hash,
    security_stamp,
    concurrency_stamp
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);
