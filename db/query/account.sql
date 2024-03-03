-- name: CreateAccount :one
INSERT INTO accounts (
  owner, 
  banlance,
  currency
) VALUES (
  $1, $2, $3
)RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts 
WHERE id = $1 LIMIT 1;

-- name: ListAccounts :many
SELECT * FROM accounts
ORDER BY id 
LIMIT $1
OFFSET $2;


-- ONLY update the balance of an account by id, and we want to get return after update
-- name: UpdateAccount :one
UPDATE accounts 
SET banlance = $2 
WHERE id = $1
RETURNING *;

-- maybe we don't want get return after delete
-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;