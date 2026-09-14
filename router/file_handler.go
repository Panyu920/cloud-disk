package router

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	db "github.com/Panyu920/cloud-disk/db/sqlc"
	"github.com/Panyu920/cloud-disk/meta"
	"github.com/Panyu920/cloud-disk/token"
	"github.com/Panyu920/cloud-disk/utils"
	"github.com/gin-gonic/gin"
)

func HandleIndex(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Welcome to the Index Page",
	})
}

type UploadRequest struct {
	Username string                  `form:"username" binding:"required"`
	Email    string                  `form:"email" binding:"required,email"`
	Files    []*multipart.FileHeader `form:"file" binding:"required"`
}
type UploadResponse struct {
	FileMetas []meta.FileMeta `json:"file_metas"`
	Username  string          `json:"username"`
	Email     string          `json:"email"`
}

func HandleUpload(c *gin.Context) {
	var form UploadRequest
	if err := c.ShouldBind(&form); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid form data", nil)
		return
	}
	// 从上下文获取 payload
	payload, ok := c.Get(AuthorizationPayloadKey)
	if !ok {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// 校验 payload 中的 username 是否与 form 中的 username 一致
	p := payload.(*token.Payload)
	if p.Username != form.Username {
		println(p.Username, form.Username)
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	fileMetas := make([]meta.FileMeta, len(form.Files))

	// 保存上传的文件到指定路径
	var err error
	for i, file := range form.Files {
		// 生成文件保存路径
		dst := filepath.Join("./uploads/", filepath.Base(file.Filename))

		fileMetas[i], err = saveUploadedFile(file, dst)
		if err != nil {
			utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to save file", nil)
			return
		}
		// meta.AddFileMeta(filemeta)
		result, err := db.StoreInstance.CreateFile(c, db.CreateFileParams{
			FileSha1: fileMetas[i].FileSha1,
			// FileName: fileMetas[i].FileName,
			FileSize: fileMetas[i].FileSize,
			FileAddr: fileMetas[i].Location,
		})
		if err != nil {
			log.Println("err:", err)
			utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to create file record", nil)
			return
		}
		fileID, err := result.LastInsertId()
		if err != nil {
			log.Println("err:", err)
			utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to create file record", nil)
			return
		}
		fileMetas[i].ID = fileID
	}
	// 获取上传的文件名列表
	// var fileNames []string
	// for _, fileHeader := range form.Files {
	// 	fileNames = append(fileNames, fileHeader.Filename)
	// }

	// 创建file_user 记录
	for i := range fileMetas {
		_, err := db.StoreInstance.CreateFileUser(c, db.CreateFileUserParams{
			Username: p.Username,
			FileSha1: fileMetas[i].FileSha1,
			FileSize: fileMetas[i].FileSize,
			Filename: fileMetas[i].FileName,
		})
		if err != nil {
			log.Println("err:", err)
			utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to create file_user record", nil)
			return
		}
	}

	utils.ResponseHandler(c, http.StatusOK, "File uploaded successfully", UploadResponse{
		FileMetas: fileMetas,
		Username:  form.Username,
		Email:     form.Email,
	})
}

func saveUploadedFile(file *multipart.FileHeader, dst string, perm ...fs.FileMode) (meta.FileMeta, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return meta.FileMeta{}, err
	}
	defer src.Close()

	//	 创建目标文件夹
	var mode os.FileMode = 0o750
	if len(perm) > 0 {
		mode = perm[0]
	}
	dir := filepath.Dir(dst)
	if err = os.MkdirAll(dir, mode); err != nil {
		return meta.FileMeta{}, err
	}
	if err = os.Chmod(dir, mode); err != nil {
		return meta.FileMeta{}, err
	}

	out, err := os.Create(dst)
	if err != nil {
		return meta.FileMeta{}, err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	if err != nil {
		return meta.FileMeta{}, err
	}
	fileSha1, err := utils.Sha256FileFromReader(out)

	if err != nil {
		return meta.FileMeta{}, err
	}
	metaData := meta.GenerateFileMeta(file.Filename, dst, file.Size, fileSha1)
	// log.Printf("%s : %s\n", file.Filename, id)
	return metaData, nil
}

// GetFileMeta 获取文件元信息
func GetFileMeta(c *gin.Context) {
	fileIDStr := c.Query("file_id")
	// meta, err := meta.GetFileMeta(fileID)
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	if err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, "file_id is invalid", nil)
		return
	}
	meta, err := db.StoreInstance.GetFileById(context.Background(), fileID)
	if err != nil {
		utils.ResponseHandler(c, http.StatusNotFound, "File not found", nil)
		return
	}
	utils.ResponseHandler(c, http.StatusOK, "File meta retrieved successfully", meta)
}

// HandleDownload 处理文件下载请求
func HandleDownload(c *gin.Context) {
	fileIDStr := c.Query("file_id")
	// meta, err := meta.GetFileMeta(fileID)
	fileID, err := strconv.ParseInt(fileIDStr, 10, 64)
	meta, err := db.StoreInstance.GetFileById(context.Background(), fileID)
	if err != nil {
		utils.ResponseHandler(c, http.StatusNotFound, "File not found", nil)
		return
	}

	c.FileAttachment(meta.FileAddr, meta.FileName)
}

type UpdateFileMetaRequest struct {
	FileID   int64  `json:"file_id" binding:"required"`
	FileName string `json:"file_name" binding:"required"`
}

// UpdateFileMeta 更新文件元信息
func UpdateFileMeta(c *gin.Context) {
	var req UpdateFileMetaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid request data", nil)
		return
	}

	// metaData, err := meta.GetFileMeta(req.FileID)
	// if err != nil {
	// 	utils.ResponseHandler(c, http.StatusNotFound, "File not found", nil)
	// 	return
	// }

	updateFileParams := db.UpdateFileParams{
		ID:       req.FileID,
		FileName: sql.NullString{Valid: true, String: req.FileName},
	}
	res, err := db.StoreInstance.UpdateFile(c, updateFileParams)
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to update file record", nil)
		return
	}

	_, err = res.RowsAffected()
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to update file record", nil)
		return
	}

	// metaData.FileName = req.FileName
	// meta.UpdateFileMeta(metaData)

	utils.ResponseHandler(c, http.StatusOK, "File meta updated successfully", nil)
}

// DeleteFile 删除文件及其元信息
func DeleteFile(c *gin.Context) {
	fileID := c.Query("file_id")
	metaData, err := meta.GetFileMeta(fileID)
	if err != nil {
		utils.ResponseHandler(c, http.StatusNotFound, "File not found", nil)
		return
	}
	err = os.Remove(metaData.Location)
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to delete file", nil)
		return
	}
	meta.RemoveFileMeta(fileID)
	utils.ResponseHandler(c, http.StatusOK, "File deleted successfully", nil)
}

type FastUploadRequest struct {
	Username string `json:"username" binding:"required"`
	Filename string `json:"filename" binding:"required"`
	FileSize int64  `json:"file_size" binding:"required"`
	FileSha1 string `json:"file_sha1" binding:"required"`
}

type FastUploadResponse struct {
	FileID int64 `json:"file_id"`
}

func FastUploadHandler(c *gin.Context) {
	var req FastUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid request data", nil)
		return
	}

	payloadS, ok := c.Get(AuthorizationPayloadKey)
	if !ok {
		utils.ResponseHandler(c, http.StatusForbidden, "forbiddend upload", nil)
		return
	}
	payload := payloadS.(*token.Payload)
	if payload.Username != req.Username {
		utils.ResponseHandler(c, http.StatusForbidden, "forbiddend upload", nil)
		return
	}

	file, err := db.StoreInstance.GetFileBySha1AndSize(c, db.GetFileBySha1AndSizeParams{
		FileSha1: req.FileSha1,
		FileSize: req.FileSize,
	})
	if err != nil {
		// 文件不存在
		if errors.Is(err, sql.ErrNoRows) {
			utils.ResponseHandler(c, http.StatusNotFound, "File not found", nil)
			return
		}
		utils.ResponseHandler(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// 文件已存在

	// 创建文件用户关联
	_, err = db.StoreInstance.CreateFileUser(c, db.CreateFileUserParams{
		Username: req.Username,
		FileSha1: req.FileSha1,
		FileSize: req.FileSize,
		Filename: req.Filename,
	})

	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.ResponseHandler(c, http.StatusOK, "File uploaded successfully", FastUploadResponse{
		FileID: file.ID,
	})
}
