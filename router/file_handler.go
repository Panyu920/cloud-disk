package router

import (
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Panyu920/cloud-disk/meta"
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
		// 生成文件保存路径
		dst := filepath.Join("./uploads/", filepath.Base(file.Filename))

		filemeta, err := saveUploadedFile(file, dst)
		if err != nil {
			utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to save file", nil)
			return
		}
		meta.AddFileMeta(filemeta)

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

func saveUploadedFile(file *multipart.FileHeader, dst string, perm ...fs.FileMode) (*meta.FileMeta, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	//	 创建目标文件夹
	var mode os.FileMode = 0o750
	if len(perm) > 0 {
		mode = perm[0]
	}
	dir := filepath.Dir(dst)
	if err = os.MkdirAll(dir, mode); err != nil {
		return nil, err
	}
	if err = os.Chmod(dir, mode); err != nil {
		return nil, err
	}

	out, err := os.Create(dst)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	if err != nil {
		return nil, err
	}
	id, err := utils.Sha256FileFromReader(out)

	if err != nil {
		return nil, err
	}
	metaData := meta.GenerateFileMeta(file.Filename, dst, file.Size, id)
	log.Println("%s : %s", file.Filename, id)
	return metaData, nil
}

// GetFileMeta 获取文件元信息
func GetFileMeta(c *gin.Context) {
	fileID := c.Query("file_id")
	meta, err := meta.GetFileMeta(fileID)
	if err != nil {
		utils.ResponseHandler(c, http.StatusNotFound, "File not found", nil)
		return
	}
	utils.ResponseHandler(c, http.StatusOK, "File meta retrieved successfully", meta)
}
