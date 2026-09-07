package meta

import (
	"errors"
	"time"
)

// FileMeta 文件元信息
type FileMeta struct {
	FileName string `json:"filename"`
	Location string `json:"location"`
	UploadAt string `json:"upload_at"`
	FileSize int64  `json:"file_size"`
	ID       string `json:"id"`
}

// fileMetas 文件元信息存储
var fileMetas map[string]FileMeta

// init 初始化函数，初始化 fileMetas
func init() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta 更新文件元信息
func UpdateFileMeta(meta *FileMeta) {
	fileMetas[meta.ID] = *meta
}

// GetFileMeta 获取文件元信息
func GetFileMeta(id string) (*FileMeta, error) {
	if meta, ok := fileMetas[id]; ok {
		return &meta, nil
	}
	return nil, errors.New("file not found")
}

func AddFileMeta(meta *FileMeta) {
	fileMetas[meta.ID] = *meta
}

func RemoveFileMeta(id string) {
	delete(fileMetas, id)
}

func GenerateFileMeta(fileName, location string, fileSize int64, id string) *FileMeta {
	return &FileMeta{
		FileName: fileName,
		Location: location,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		FileSize: fileSize,
		ID:       id,
	}
}
