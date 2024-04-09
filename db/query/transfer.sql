-- name: CreateTransfer :one
INSERT INTO transfers (
  from_account_id,
  to_account_id,
  amount
) VALUES (
  $1, $2, $3
)RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers 
WHERE id = $1
LIMIT 1;

-- name: ListTransfers :many
SELECT * FROM transfers
WHERE 
    from_account_id = $1 OR
    to_account_id = $2
ORDER BY id
LIMIT $3
OFFSET $4;



-- ONLY update the amount of an entry by from_account_id AND to_account_id 
-- name: UpdateTransfer :one
UPDATE transfers
SET amount = $3 
WHERE from_account_id = $1 AND to_account_id = $2
RETURNING *;
