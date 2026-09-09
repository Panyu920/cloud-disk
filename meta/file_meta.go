package meta

import (
	"errors"
	"sync"
	"time"
)

// FileMeta 文件元信息
type FileMeta struct {
	FileName string `json:"filename"`
	Location string `json:"location"`
	UploadAt string `json:"upload_at"`
	FileSize int64  `json:"file_size"`
	FileSha1 string `json:"file_sha1"`
	ID       int64  `json:"id"`
}

// fileMetas 文件元信息存储
var fileMetas map[string]FileMeta
var rwlock = &sync.RWMutex{}

// init 初始化函数，初始化 fileMetas
func init() {
	rwlock.Lock()
	defer rwlock.Unlock()
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta 更新文件元信息
func UpdateFileMeta(meta *FileMeta) {
	rwlock.Lock()
	defer rwlock.Unlock()
	fileMetas[meta.FileSha1] = *meta
}

// GetFileMeta 获取文件元信息
func GetFileMeta(fileSha1 string) (*FileMeta, error) {
	rwlock.RLock()
	defer rwlock.RUnlock()
	if meta, ok := fileMetas[fileSha1]; ok {
		return &meta, nil
	}
	return nil, errors.New("file not found")
}

func AddFileMeta(meta *FileMeta) {
	rwlock.Lock()
	defer rwlock.Unlock()
	fileMetas[meta.FileSha1] = *meta
}

func RemoveFileMeta(id string) {
	rwlock.Lock()
	defer rwlock.Unlock()
	delete(fileMetas, id)
}

func GenerateFileMeta(fileName, location string, fileSize int64, fileSha1 string) FileMeta {
	return FileMeta{
		FileName: fileName,
		Location: location,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		FileSize: fileSize,
		FileSha1: fileSha1,
	}
}
