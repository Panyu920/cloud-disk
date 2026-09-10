-- name: CreateUser :execresult
INSERT INTO users (
    username, password,  email, phone    )
     VALUES (?, ?, ?, ?);

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = ? LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ? LIMIT 1;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone = ? LIMIT 1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = ? LIMIT 1;

-- name: UpdateUser :execresult
UPDATE users SET
    username = coalesce(sqlc.narg(username), username),
    password = coalesce(sqlc.narg(password), password),
    email = coalesce(sqlc.narg(email), email),
    phone = coalesce(sqlc.narg(phone), phone),
    email_verified = coalesce(sqlc.narg(email_verified), email_verified),
    phone_verified = coalesce(sqlc.narg(phone_verified), phone_verified),
    profile = coalesce(sqlc.narg(profile), profile),
    status = coalesce(sqlc.narg(status), status),
    update_at = CURRENT_TIMESTAMP,
    last_login_at = coalesce(sqlc.narg(last_login_at), last_login_at)
WHERE id = ?;
