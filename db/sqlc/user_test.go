package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/Panyu920/cloud-disk/utils"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) CreateUserParams {
	// 生成随机用户
	user := CreateUserParams{
		Username: utils.RandomString(10),
		Password: utils.RandomString(10),
		Email:    utils.RandomString(10) + "@test.com",
		Phone:    utils.RandomPhone(),
	}

	res, err := testQueries.CreateUser(context.Background(), user)
	require.NoError(t, err)
	_, err = res.LastInsertId()
	require.NoError(t, err)
	affectedRow, err := res.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), affectedRow)
	return user

}
func TestCreateUser(t *testing.T) {
	// 测试创建用户
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	// 测试获取用户
	user := createRandomUser(t)
	res, err := testQueries.GetUserByUsername(context.Background(), user.Username)
	require.NoError(t, err)
	require.Equal(t, user.Username, res.Username)
	require.Equal(t, user.Email, res.Email)
	require.Equal(t, user.Phone, res.Phone)
}

func TestUpdateUser(t *testing.T) {
	// 测试更新用户
	userParam := createRandomUser(t)
	oldUser, err := testQueries.GetUserByUsername(context.Background(), userParam.Username)
	require.NoError(t, err)

	// 更新用户
	updateUserParams := UpdateUserParams{
		ID:            oldUser.ID,
		Username:      sql.NullString{String: utils.RandomString(10), Valid: true},
		Password:      sql.NullString{String: utils.RandomString(10), Valid: true},
		Email:         sql.NullString{String: utils.RandomString(10) + "@test.com", Valid: true},
		Phone:         sql.NullString{String: utils.RandomPhone(), Valid: true},
		EmailVerified: sql.NullBool{Valid: true},
		PhoneVerified: sql.NullBool{Valid: true},
		Profile:       sql.NullString{String: utils.RandomString(10), Valid: true},
		Status:        sql.NullInt16{Int16: int16(utils.RandomInt(2, 0)), Valid: true},
		LastLoginAt:   sql.NullTime{Valid: true, Time: time.Now()},
	}
	res, err := testQueries.UpdateUser(context.Background(), updateUserParams)
	require.NoError(t, err)
	_, err = res.LastInsertId()
	require.NoError(t, err)
	affectedRow, err := res.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), affectedRow)
}
