-- name: CreateFile :execresult
INSERT INTO files (
    file_sha1, file_name, file_size, file_addr )
     VALUES (?, ?, ?, ?);
    
-- name: GetFileBySha1 :one
SELECT * FROM files WHERE file_sha1 = ? limit 1;

-- name: GetFileById :one
SELECT * FROM files WHERE id = ? limit 1;

-- name: UpdateFile :execresult
UPDATE files SET 
    file_name = coalesce(sqlc.narg(file_name), file_name),
    file_size = coalesce(sqlc.narg(file_size), file_size),
    file_addr = coalesce(sqlc.narg(file_addr), file_addr),
    status = coalesce(sqlc.narg(status), status),
    update_at = CURRENT_TIMESTAMP,
    extend = coalesce(sqlc.narg(extend), extend),   
    extend2 = coalesce(sqlc.narg(extend2), extend2)
WHERE id = ?;