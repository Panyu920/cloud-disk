package meta

import (
	"errors"
	"time"
)

// FileMeta 文件元信息
type FileMeta struct {
	FileName string
	Location string
	UploadAt string
	FileSize int64
	Md5      string
}

// fileMetas 文件元信息存储
var fileMetas map[string]FileMeta

// init 初始化函数，初始化 fileMetas
func init() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta 更新文件元信息
func UpdateFileMeta(meta *FileMeta) {
	fileMetas[meta.Md5] = *meta
}

// GetFileMeta 获取文件元信息
func GetFileMeta(md5 string) (*FileMeta, error) {
	if meta, ok := fileMetas[md5]; ok {
		return &meta, nil
	}
	return nil, errors.New("file not found")
}

func AddFileMeta(meta *FileMeta) {
	fileMetas[meta.Md5] = *meta
}

func RemoveFileMeta(md5 string) {
	delete(fileMetas, md5)
}

func GenerateFileMeta(fileName, location string, fileSize int64, md5 string) *FileMeta {
	return &FileMeta{
		FileName: fileName,
		Location: location,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		FileSize: fileSize,
		Md5:      md5,
	}
}
