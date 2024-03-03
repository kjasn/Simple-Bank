-- name: CreateEntry :one
INSERT INTO entries (
  account_id,
  amount
) VALUES (
  $1, $2
)RETURNING *;

-- name: GetEntry :one
SELECT * FROM entries 
WHERE id = $1 LIMIT 1;

-- name: ListEntries :many
SELECT * FROM entries
WHERE account_id = $1
LIMIT $2
OFFSET $3;


-- ONLY update the amount of an entry by account_id 
-- name: UpdateEntry :one
UPDATE entries
SET amount = $2 
WHERE id = $1
RETURNING *;

-- don't return 
-- delete one entry
-- name: DeleteEntry :exec
DELETE FROM entries
WHERE id = $1;


-- delete many entries by account_id
-- name: DeleteEntries :exec
DELETE FROM entries
WHERE account_id = $1;