-- name: CreateTransfer :one
INSERT INTO transfers (
  from_account_id,
  to_account_id,
  amount
) VALUES (
  $1, $2, $3
)RETURNING *;

-- get by from_account_id AND to_account_id
-- name: GetTransfer :one
SELECT * FROM transfers 
WHERE from_account_id = $1 AND to_account_id = $2
LIMIT 1;

-- get by from_account_id
-- name: ListFromIDTransfers :many
SELECT * FROM transfers
WHERE from_account_id = $1
LIMIT $2
OFFSET $3;

-- get by to_account_id
-- name: ListToIDTransfers :many
SELECT * FROM transfers
WHERE to_account_id = $1
LIMIT $2
OFFSET $3;


-- ONLY update the amount of an entry by from_account_id AND to_account_id 
-- name: UpdateTransfer :one
UPDATE transfers
SET amount = $3 
WHERE from_account_id = $1 AND to_account_id = $2
RETURNING *;

-- delete one transfer
-- name: DeleteTransfer :exec
DELETE FROM transfers
WHERE from_account_id = $1 AND to_account_id = $2;


-- delete many transfers by from_account_id
-- name: DeleteFromIDTransfers :exec
DELETE FROM transfers
WHERE from_account_id = $1;


-- delete many transfers by to_account_id
-- name: DeleteTOIDTransfers :exec
DELETE FROM transfers
WHERE to_account_id = $1;