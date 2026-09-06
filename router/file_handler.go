package router

import (
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Panyu920/cloud-disk/utils"
	"github.com/gin-gonic/gin"
)

func HandleIndex(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Welcome to the Index Page",
	})
}

type UploadRequest struct {
	Name  string                  `form:"name" binding:"required"`
	Email string                  `form:"email" binding:"required,email"`
	Files []*multipart.FileHeader `form:"file" binding:"required"`
}
type UploadResponse struct {
	Filename string `json:"filename"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

func HandleUpload(c *gin.Context) {
	var form UploadRequest
	if err := c.ShouldBind(&form); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid form data", nil)
		return
	}

	// 保存上传的文件到指定路径
	for _, file := range form.Files {
		dst := filepath.Join("./uploads/", filepath.Base(file.Filename))

		if err := c.SaveUploadedFile(file, dst); err != nil {
			utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to save file", nil)
			return
		}
	}
	// 获取上传的文件名列表
	var fileNames []string
	for _, fileHeader := range form.Files {
		fileNames = append(fileNames, fileHeader.Filename)
	}

	utils.ResponseHandler(c, http.StatusOK, "File uploaded successfully", UploadResponse{
		Filename: strings.Join(fileNames, ","),
		Name:     form.Name,
		Email:    form.Email,
	})
}
