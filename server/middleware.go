package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Panyu920/cloud-disk/token"
	"github.com/Panyu920/cloud-disk/utils"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey    = "authorization"
	AuthorizationHeaderPrefix = "bearer"
	AuthorizationPayloadKey   = "authorization_payload"
)

var (
	ErrAuthorizationHeaderMissing = errors.New("Authorization header is missing")
	ErrAuthorizationFormatInvalid = errors.New("Authorization header format is invalid")
	ErrAuthorizationTypeInvalid   = errors.New("Authorization type is invalid")
	ErrAuthorizationTokenInvalid  = errors.New("Authorization token is invalid")
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查请求头是否包含 Authorization 字段
		authHeader := c.GetHeader(AuthorizationHeaderKey)
		if len(authHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponseHandler(ErrAuthorizationHeaderMissing))
			return
		}
		fields := strings.Fields(authHeader)
		if len(fields) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponseHandler(ErrAuthorizationFormatInvalid))
			return
		}
		// 验证 Authorization 字段是否有效
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationHeaderPrefix {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponseHandler(ErrAuthorizationTypeInvalid))
			return
		}

		// 验证token
		payload, err := token.DefaultMaker.VerifyToken(fields[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponseHandler(err))
			return
		}

		// 存储 payload 到上下文
		c.Set(AuthorizationPayloadKey, payload)
		c.Next()
	}
}
