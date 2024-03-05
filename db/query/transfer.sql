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

-- name: ListFromIDTransfers :many
SELECT * FROM transfers
ORDER BY id
LIMIT $1
OFFSET $2;



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