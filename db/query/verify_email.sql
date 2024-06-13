-- name: CreateVerifyEmail :one
INSERT INTO verify_emails (
  secret_code,  
  username, 
  email 
) VALUES (
    $1, $2, $3
) RETURNING *;