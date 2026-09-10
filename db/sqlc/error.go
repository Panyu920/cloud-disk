package db

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

const (
	ErrUsernameAlready = "username already exists"
	ErrEmailAlready    = "email already exists"
	ErrPhoneAlready    = "phone already exists"
	ErrUserNotFound    = "user not found"
)

func IsUniqueError(err error) (bool, string) {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		if mysqlErr.Number == 1062 {
			return true, mysqlErr.Message
		}
	}
	return false, ""
}
