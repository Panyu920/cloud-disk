package router

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	db "github.com/Panyu920/cloud-disk/db/sqlc"
	"github.com/Panyu920/cloud-disk/token"
	"github.com/Panyu920/cloud-disk/utils"
	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=4,max=20"`
	Password string `json:"password" binding:"required,min=4,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"required,phone"`
}

type CreateUserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

// CreateUserHandler 创建用户
func CreateUserHandler(c *gin.Context) {
	var userParam CreateUserRequest
	if err := c.ShouldBindJSON(&userParam); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	// 创建用户
	res, err := db.StoreInstance.CreateUser(c, db.CreateUserParams{
		Username: userParam.Username,
		Password: utils.GeneratePasswordHash(userParam.Password),
		Email:    userParam.Email,
		Phone:    userParam.Phone,
	})
	// 处理错误
	if err != nil {
		isUnique, msg := db.IsUniqueError(err)
		// 处理唯一键约束错误
		if isUnique {
			utils.ResponseHandler(c, http.StatusBadRequest, msg, nil)
			return
		}
		utils.ResponseHandler(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	// 获取插入的用户ID
	userID, err := res.LastInsertId()
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// 返回成功响应
	utils.ResponseHandler(c, http.StatusOK, "user created successfully", CreateUserResponse{
		ID:       userID,
		Username: userParam.Username,
		Email:    userParam.Email,
		Phone:    userParam.Phone,
	})

}

type LoginRequest struct {
	Username string `json:"username" binding:"required,min=4,max=20"`
	Password string `json:"password" binding:"required,min=4,max=20"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
}

// LoginHandler 登录用户
func LoginHandler(c *gin.Context) {
	var loginParam LoginRequest
	if err := c.ShouldBindJSON(&loginParam); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 验证用户密码
	user, err := db.StoreInstance.GetUserByUsername(c, loginParam.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 用户不存在
			utils.ResponseHandler(c, http.StatusUnauthorized, "invalid username or password", nil)
			return
		}
		// 其他错误
		utils.ResponseHandler(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// 验证密码
	if !utils.CheckPasswordHash(loginParam.Password, user.Password) {
		utils.ResponseHandler(c, http.StatusUnauthorized, "invalid username or password", nil)
		return
	}

	// 生成登录令牌
	token, payload, err := token.DefaultMaker.CreateToken(loginParam.Username, user.ID, token.DefaultDuration)
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// 返回成功响应
	utils.ResponseHandler(c, http.StatusOK, "login success", LoginResponse{
		Token:     token,
		ExpiredAt: payload.ExpiresAt.Time,
	})
}
