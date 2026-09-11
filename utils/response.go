package utils

import "github.com/gin-gonic/gin"

func ResponseHandler(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, gin.H{
		"message": message,
		"data":    data,
	})
}

func ErrorResponseHandler(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}
