-- name: VerifyEmail :one
INSERT INTO verify_emails (
	secret_code,  
	username, 
	email 
) VALUES (
	$1, $2, $3
) RETURNING *;

-- name: UpdateVerifyEmail :one
UPDATE verify_emails 
SET 
	is_used = true
WHERE id = @id 
	AND secret_code = @secret_code
	AND is_used = FALSE
	AND expired_at > now()
RETURNING *;
