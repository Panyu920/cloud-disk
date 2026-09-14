package router

import (
	"context"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Panyu920/cloud-disk/cache"
	db "github.com/Panyu920/cloud-disk/db/sqlc"
	"github.com/Panyu920/cloud-disk/token"
	"github.com/Panyu920/cloud-disk/utils"
	"github.com/gin-gonic/gin"
)

const (
	chunkSize = 1024 * 1024 * 5
)

type InitMultiPartUploadRequest struct {
	Username string `json:"username" binding:"required"`
	FileSize int64  `json:"file_size" binding:"required"`
}

type InitMultiPartUploadResponse struct {
	UploadID    string `json:"upload_id"`
	ChunkSize   int64  `json:"chunk_size"`
	TotalChunks int64  `json:"total_chunks"`
}

func InitMultiPartUploadHandler(c *gin.Context) {
	// 1. json 解析
	var req InitMultiPartUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid json data", nil)
		return
	}
	// 2. 校验 form 中的 username 是否与 payload 中的 username 一致
	payload, ok := c.Get(AuthorizationPayloadKey)
	if !ok {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	p := payload.(*token.Payload)
	if p.Username != req.Username {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	// 3. 设置uploadID
	uploadID := utils.GenerateUUID()
	// 4. 文件进行分块
	totalChunks := math.Ceil(float64(req.FileSize) / float64(chunkSize))
	// 5. 存储到redis
	_, err := cache.RedisClient.HSet(context.Background(), "mpID"+uploadID, "username", req.Username, "fileSize", req.FileSize, "chunkSize", chunkSize, "totalChunks", totalChunks).Result()
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to store upload info", nil)
		return
	}
	// 6. 返回uploadID
	utils.ResponseHandler(c, http.StatusOK, "Success", InitMultiPartUploadResponse{
		UploadID:    uploadID,
		ChunkSize:   int64(chunkSize),
		TotalChunks: int64(totalChunks),
	})
}

type UploadPartRequest struct {
	Username   string                `form:"username" binding:"required"`
	UploadID   string                `form:"upload_id" binding:"required"`
	ChunkSize  int64                 `form:"chunk" binding:"required"`
	ChunkIndex int64                 `form:"chunk_index" binding:"required"`
	ChunkHash  string                `form:"chunk_hash" binding:"required"`
	FileHeader *multipart.FileHeader `form:"file"`
}

func UploadPartHandler(c *gin.Context) {
	// 1. form 解析
	var req UploadPartRequest
	if err := c.ShouldBind(&req); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid form data", nil)
		return
	}

	// 2. 校验 form 中的 username 是否与 payload 中的 username 一致
	payload, ok := c.Get(AuthorizationPayloadKey)
	if !ok {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	p := payload.(*token.Payload)
	if p.Username != req.Username {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// 3. 校验 form 中的 uploadID 是否与 redis 中的 uploadID 一致
	uploadID := req.UploadID
	_, err := cache.RedisClient.HGetAll(context.Background(), "mpID"+uploadID).Result()
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to get upload info", nil)
		return
	}

	// 4. 创建文件目录
	path := "./data/" + uploadID
	err = os.MkdirAll(path, 0755)
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to create directory", nil)
		return
	}

	// 5. 创建文件
	file, err := os.Create(path + "/" + strconv.FormatInt(req.ChunkIndex, 10))
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to create file", nil)
		return
	}
	defer file.Close()

	// 6. 写入文件
	src, err := req.FileHeader.Open()
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to open file", nil)
		return
	}
	defer src.Close()
	if _, err := io.Copy(file, src); err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to write file", nil)
		return
	}

	// 7. 校验文件哈希值是否一致
	getChunkHash, err := utils.Sha256FileFromReader(file)
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to get chunk hash", nil)
		return
	}
	if getChunkHash != req.ChunkHash {
		// 删除文件
		if err := os.Remove(file.Name()); err != nil {
			utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to upload chunk", nil)
			return
		}
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid chunk hash", nil)
		return
	}

	// 8. 存储到redis
	_, err = cache.RedisClient.HSet(context.Background(), "mpID"+uploadID, "chunkIndex"+strconv.FormatInt(req.ChunkIndex, 10), req.ChunkIndex).Result()
	if err != nil {
		if err := os.Remove(file.Name()); err != nil {
			utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to upload chunk", nil)
			return
		}
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to store chunk hash", nil)
		return
	}

	// 9. 返回成功
	utils.ResponseHandler(c, http.StatusOK, "Success", nil)
}

type CompleteMultiPartUploadRequest struct {
	Username string `json:"username" binding:"required"`
	UploadID string `json:"upload_id" binding:"required"`
	FileName string `json:"file_name" binding:"required"`
	FileSize int64  `json:"file_size" binding:"required"`
	FileHash string `json:"file_hash" binding:"required"`
}

func CompleteMultiPartUploadHandler(c *gin.Context) {
	// 1. json 解析
	var req CompleteMultiPartUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid json data", nil)
		return
	}

	// 2. 校验 json 中的 username 是否与 payload 中的 username 一致
	payload, ok := c.Get(AuthorizationPayloadKey)
	if !ok {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	p := payload.(*token.Payload)
	if p.Username != req.Username {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// 3. redis中的totalChunks
	record, err := cache.RedisClient.HGetAll(context.Background(), "mpID"+req.UploadID).Result()
	rec_chunkCount := 0
	temp_chunkCount := 0
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to get upload info", nil)
		return
	}
	for k, v := range record {
		if k == "totalChunks" {
			rec_chunkCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chunkIndex") {
			temp_chunkCount++
		}
	}
	if rec_chunkCount != temp_chunkCount {
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid chunk count", nil)
		return
	}

	// 4. 合并文件
	path := "./data/" + req.UploadID
	dstPath := "./uploads/" + req.UploadID
	fileSize, err := MergeFiles(path, dstPath)
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to merge files", nil)
		return
	}
	// 5.保存到files表
	fileSha1, err := utils.Sha256File(dstPath)
	if err != nil {
		// 删除文件
		if err := os.Remove(dstPath); err != nil {
			utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to delete file", nil)
			return
		}
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to get file hash", nil)
		return
	}
	_, err = db.StoreInstance.CreateFile(context.Background(), db.CreateFileParams{
		FileSha1: fileSha1,
		FileSize: fileSize,
		FileAddr: dstPath,
	})
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to create file", nil)
		return
	}
	// 6.保存记录到file_users表
	_, err = db.StoreInstance.CreateFileUser(context.Background(), db.CreateFileUserParams{
		Username: req.Username,
		FileSha1: fileSha1,
		FileSize: fileSize,
		Filename: req.FileName,
	})
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to create file user", nil)
		return
	}

	// 7.删除redis中的uploadID
	_, err = cache.RedisClient.Del(context.Background(), "mpID"+req.UploadID).Result()
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to delete upload info", nil)
		return
	}

	// 8. 返回成功
	utils.ResponseHandler(c, http.StatusOK, "Success", nil)
}

// MergeFiles 合并文件
func MergeFiles(srcDir, dstPath string) (int64, error) {
	var fileSize int64
	// 1. 读取文件目录
	files, err := os.ReadDir(srcDir)
	if err != nil {
		return fileSize, err
	}
	// 2. 排序文件名
	fileNames := make([]string, 0, len(files))
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fileNames = append(fileNames, file.Name())
	}
	sort.Slice(fileNames, func(i, j int) bool {
		a, _ := strconv.Atoi(fileNames[i])
		b, _ := strconv.Atoi(fileNames[j])
		return a < b
	})
	// 3. 打开要写入的目标文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fileSize, err
	}
	defer dstFile.Close()

	// 4. 合并文件
	for _, fileName := range fileNames {
		srcFilePath := srcDir + "/" + fileName
		srcFile, err := os.Open(srcFilePath)
		if err != nil {
			return fileSize, err
		}
		fileInfo, err := srcFile.Stat()
		if err != nil {
			return fileSize, err
		}
		fileSize += fileInfo.Size()
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			srcFile.Close()
			return fileSize, err
		}
		srcFile.Close()
	}

	return fileSize, nil
}

type CancelMultiPartUploadRequest struct {
	Username string `json:"username" binding:"required"`
	UploadID string `json:"upload_id" binding:"required"`
}

type CancelMultiPartUploadResponse struct {
	Success bool `json:"success"`
}

// CancelMultiPartUploadHandler 取消多部分上传
func CancelMultiPartUploadHandler(c *gin.Context) {
	// 1. json 解析
	var req CancelMultiPartUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid json data", nil)
		return
	}
	// 2. 校验 json 中的 username 是否与 payload 中的 username 一致
	payload, ok := c.Get(AuthorizationPayloadKey)
	if !ok {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	p := payload.(*token.Payload)
	if p.Username != req.Username {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// 3. 校验 uploadID 是否存在
	record, err := cache.RedisClient.HGetAll(context.Background(), "mpID"+req.UploadID).Result()
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to get upload info", nil)
		return
	}
	if len(record) == 0 {
		utils.ResponseHandler(c, http.StatusBadRequest, "Upload ID not found", nil)
		return
	}

	// 4. 校验 uploadID 是否属于当前用户
	if record["username"] != req.Username {
		utils.ResponseHandler(c, http.StatusBadRequest, "Upload ID not found", nil)
		return
	}

	// 5. 删除redis中的uploadID
	_, err = cache.RedisClient.Del(context.Background(), "mpID"+req.UploadID).Result()
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to delete upload info", nil)
		return
	}

	// 6. 删除已经上传的分块内容
	if err := os.RemoveAll("./data/" + req.UploadID); err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to delete files", nil)
		return
	}

	// 7. 返回成功
	utils.ResponseHandler(c, http.StatusOK, "Success", nil)
}

type GetMultiPartUploadInfoRequest struct {
	Username string `json:"username" binding:"required"`
	UploadID string `json:"upload_id" binding:"required"`
}
type GetMultiPartUploadInfoResponse struct {
	UploadID       string `json:"upload_id"`
	FileSize       int64  `json:"file_size"`
	ChunkSize      int64  `json:"chunk_size"`
	TotalChunks    int64  `json:"total_chunks"`
	UploadedChunks int64  `json:"uploaded_chunks"`
}

func GetMultiPartUploadInfoHandler(c *gin.Context) {
	// 1. json 解析
	var req GetMultiPartUploadInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid json data", nil)
		return
	}

	// 2. 校验 json 中的 username 是否与 payload 中的 username 一致
	payload, ok := c.Get(AuthorizationPayloadKey)
	if !ok {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	p := payload.(*token.Payload)
	if p.Username != req.Username {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// 3. 校验 uploadID 是否存在
	record, err := cache.RedisClient.HGetAll(context.Background(), "mpID"+req.UploadID).Result()
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to get upload info", nil)
		return
	}
	if len(record) == 0 {
		utils.ResponseHandler(c, http.StatusBadRequest, "Upload ID not found", nil)
		return
	}
	// 4. 校验 uploadID 是否属于当前用户
	if record["username"] != req.Username {
		utils.ResponseHandler(c, http.StatusBadRequest, "Upload ID not found", nil)
		return
	}

	// 5. 解析上传信息
	var uploadedChunks int64
	for k, _ := range record {
		if strings.HasPrefix(k, "chunkIndex") {
			uploadedChunks++
		}
	}
	totalChunks, _ := strconv.ParseInt(record["totalChunks"], 10, 64)
	fileSize, _ := strconv.ParseInt(record["fileSize"], 10, 64)
	chunkSize, _ := strconv.ParseInt(record["chunkSize"], 10, 64)

	// 6. 返回成功
	utils.ResponseHandler(c, http.StatusOK, "Success", GetMultiPartUploadInfoResponse{
		UploadID:       req.UploadID,
		FileSize:       fileSize,
		ChunkSize:      chunkSize,
		TotalChunks:    totalChunks,
		UploadedChunks: uploadedChunks,
	})
}

type ContinueMultiPartUploadRequest struct {
	Username string `json:"username" binding:"required"`
	UploadID string `json:"upload_id" binding:"required"`
}

type ContinueMultiPartUploadResponse struct {
	UploadID       string  `json:"upload_id"`
	UploadedChunks int64   `json:"uploaded_chunks"`
	ChunkSize      int64   `json:"chunk_size"`
	UploadedIndex  []int64 `json:"uploaded_index"`
}

func ContinueMultiPartUploadHandler(c *gin.Context) {
	// 1. json 解析
	var req ContinueMultiPartUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseHandler(c, http.StatusBadRequest, "Invalid json data", nil)
		return
	}

	// 2. 校验 json 中的 username 是否与 payload 中的 username 一致
	payload, ok := c.Get(AuthorizationPayloadKey)
	if !ok {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}
	p := payload.(*token.Payload)
	if p.Username != req.Username {
		utils.ResponseHandler(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// 3. 校验 uploadID 是否存在
	record, err := cache.RedisClient.HGetAll(context.Background(), "mpID"+req.UploadID).Result()
	if err != nil {
		utils.ResponseHandler(c, http.StatusInternalServerError, "Failed to get upload info", nil)
		return
	}
	if len(record) == 0 {
		utils.ResponseHandler(c, http.StatusBadRequest, "Upload ID not found", nil)
		return
	}
	// 4. 校验 uploadID 是否属于当前用户
	if record["username"] != req.Username {
		utils.ResponseHandler(c, http.StatusBadRequest, "Upload ID not found", nil)
		return
	}

	// 5. 解析上传信息
	var uploadedChunks int64
	var uploadedIndex []int64
	for k, v := range record {
		if strings.HasPrefix(k, "chunkIndex") {
			uploadedChunks++
			index, _ := strconv.ParseInt(v, 10, 64)
			uploadedIndex = append(uploadedIndex, index)
		}
	}
	chunkSize, _ := strconv.ParseInt(record["chunkSize"], 10, 64)

	// 6. 返回成功
	utils.ResponseHandler(c, http.StatusOK, "Success", ContinueMultiPartUploadResponse{
		UploadID:       req.UploadID,
		UploadedChunks: uploadedChunks,
		ChunkSize:      chunkSize,
		UploadedIndex:  uploadedIndex,
	})
}
