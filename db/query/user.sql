-- name: CreateUser :one
INSERT INTO users (
  username, 
  hashed_password,
  full_name,
  email
) VALUES (
  $1, $2, $3, $4
)RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- -- name: GetUserForUpdate :one
-- SELECT * FROM users 
-- WHERE username = $1 LIMIT 1
-- FOR NO KEY UPDATE;

-- -- name: ListUsers :many
-- SELECT * FROM users
-- ORDER BY username 
-- LIMIT $1
-- OFFSET $2;


-- -- ONLY update the balance of an account by username, and we want to get return after update
-- -- name: UpdatePassword :one
-- UPDATE users 
-- SET hashed_password = $2 
-- WHERE username = $1
-- RETURNING *;

-- -- name: UpdateEmail :one
-- UPDATE users 
-- SET email = $2 
-- WHERE username = $1
-- RETURNING *;

-- -- maybe we don't want get return after delete
-- -- name: DeleteUser :exec
-- DELETE FROM users
-- WHERE username = $1;