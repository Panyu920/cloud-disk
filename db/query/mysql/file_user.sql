-- name: CreateFileUser :execresult
INSERT INTO file_users (
    username, file_sha1, file_size, filename )
     VALUES (?, ?, ?, ?);   

-- name: GetUserFilesByUsername :many
SELECT * FROM file_users WHERE username = ? AND status = 0 
ORDER BY create_at DESC
limit ? offset ?;

-- name: UpdateFileUserStatus :execresult
UPDATE file_users SET 
status = coalesce(sqlc.narg(status), status),
filename = coalesce(sqlc.narg(filename), filename),
update_at = now()
WHERE filename = ? AND file_sha1 = ? AND status = ? AND username = ?
