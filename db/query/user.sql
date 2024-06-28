-- name: CreateUser :one
INSERT INTO users (
  username, 
  role,
  hashed_password,
  full_name,
  email
) VALUES (
  $1, $2, $3, $4, $5
)RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users 
SET 
  role = COALESCE(sqlc.narg(role), role),
  hashed_password = COALESCE(sqlc.narg(hashed_password), hashed_password), 
  full_name = COALESCE(sqlc.narg(full_name), full_name),
  email = COALESCE(sqlc.narg(email), email), 
  is_email_verified = COALESCE(sqlc.narg(is_email_verified), is_email_verified), 
  password_changed_at = now()
WHERE username = sqlc.arg(username)
RETURNING *;