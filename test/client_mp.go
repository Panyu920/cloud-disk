package test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// 服务端接口路径（按你的路由实际前缀改）
const (
	pathInit     = "/upload/init"
	pathPart     = "/upload/part"
	pathComplete = "/upload/complete"
	pathCancel   = "/upload/cancel"
	pathInfo     = "/upload/info"
	pathContinue = "/upload/continue"
)

// Client 分块上传客户端
type Client struct {
	BaseURL     string
	Username    string
	Token       string
	HTTPClient  *http.Client
	Concurrency int
	MaxRetries  int
	RetryDelay  time.Duration
	OnProgress  func(uploaded, total int64)
}

// NewClient 创建客户端
func NewClient(baseURL, username, token string) *Client {
	return &Client{
		BaseURL:     baseURL,
		Username:    username,
		Token:       token,
		HTTPClient:  &http.Client{Timeout: 5 * time.Minute},
		Concurrency: 4,
		MaxRetries:  3,
		RetryDelay:  500 * time.Millisecond,
	}
}

// ---------- 统一响应 ----------

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// ---------- 请求/响应结构 ----------

type InitRequest struct {
	Username string `json:"username"`
	FileSize int64  `json:"file_size"`
}

type InitResponse struct {
	UploadID    string `json:"upload_id"`
	ChunkSize   int64  `json:"chunk_size"`
	TotalChunks int64  `json:"total_chunks"`
}

type CompleteRequest struct {
	Username string `json:"username"`
	UploadID string `json:"upload_id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	FileHash string `json:"file_hash"`
}

type CancelRequest struct {
	Username string `json:"username"`
	UploadID string `json:"upload_id"`
}

type InfoRequest struct {
	Username string `json:"username"`
	UploadID string `json:"upload_id"`
}

type InfoResponse struct {
	UploadID       string `json:"upload_id"`
	FileSize       int64  `json:"file_size"`
	ChunkSize      int64  `json:"chunk_size"`
	TotalChunks    int64  `json:"total_chunks"`
	UploadedChunks int64  `json:"uploaded_chunks"`
}

type ContinueRequest struct {
	Username string `json:"username"`
	UploadID string `json:"upload_id"`
}

type ContinueResponse struct {
	UploadID       string  `json:"upload_id"`
	UploadedChunks int64   `json:"uploaded_chunks"`
	ChunkSize      int64   `json:"chunk_size"`
	UploadedIndex  []int64 `json:"uploaded_index"`
}

// ---------- 主流程 ----------

// UploadResult 上传结果
type UploadResult struct {
	UploadID  string
	FileHash  string
	FileSize  int64
	Uploaded  int64 // 本次实际上传的字节数
	Skipped   int64 // 断点续传跳过的字节数
	TotalPart int64
}

// Upload 完整上传：init -> （可选 continue）-> 并发补传分块 -> complete
func (c *Client) Upload(ctx context.Context, filePath string) (*UploadResult, error) {
	return c.UploadWithResume(ctx, filePath, "")
}

// UploadWithResume 支持断点续传：
// 如果 resumeUploadID 非空，则先调用 Continue 查询已上传分块，只补传缺失的；
// 否则调用 Init 新建会话。
func (c *Client) UploadWithResume(ctx context.Context, filePath, resumeUploadID string) (*UploadResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}
	fileSize := info.Size()
	if fileSize <= 0 {
		return nil, errors.New("文件为空")
	}

	// 1. 计算整文件 SHA-256
	fileHash, err := hashReader(f)
	if err != nil {
		return nil, fmt.Errorf("计算文件哈希失败: %w", err)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("重置文件偏移失败: %w", err)
	}

	var (
		uploadID    string
		chunkSize   int64
		totalChunks int64
		uploadedSet map[int64]bool
	)

	// 2. 决定是新建还是续传
	if resumeUploadID != "" {
		cont, err := c.Continue(ctx, resumeUploadID)
		if err != nil {
			return nil, fmt.Errorf("查询续传信息失败: %w", err)
		}
		uploadID = cont.UploadID
		chunkSize = cont.ChunkSize
		if chunkSize <= 0 {
			return nil, errors.New("服务端返回 chunk_size 非法")
		}
		totalChunks = (fileSize + chunkSize - 1) / chunkSize
		uploadedSet = make(map[int64]bool, len(cont.UploadedIndex))
		for _, idx := range cont.UploadedIndex {
			uploadedSet[idx] = true
		}
	} else {
		initResp, err := c.Init(ctx, fileSize)
		if err != nil {
			return nil, fmt.Errorf("初始化上传失败: %w", err)
		}
		uploadID = initResp.UploadID
		chunkSize = initResp.ChunkSize
		totalChunks = initResp.TotalChunks
		if chunkSize <= 0 || totalChunks <= 0 {
			return nil, errors.New("服务端返回的 chunk_size / total_chunks 非法")
		}
		uploadedSet = map[int64]bool{}
	}

	// 3. 并发补传缺失分块
	res, err := c.uploadMissingParts(ctx, f, fileSize, uploadID, chunkSize, totalChunks, uploadedSet)
	if err != nil {
		return nil, err
	}
	res.UploadID = uploadID
	res.FileHash = fileHash
	res.FileSize = fileSize

	// 4. 合并
	if err := c.Complete(ctx, uploadID, filepath.Base(filePath), fileSize, fileHash); err != nil {
		return nil, fmt.Errorf("合并分块失败: %w", err)
	}
	return res, nil
}

// uploadMissingParts 并发上传所有未上传的分块
func (c *Client) uploadMissingParts(
	ctx context.Context,
	f *os.File,
	fileSize int64,
	uploadID string,
	chunkSize int64,
	totalChunks int64,
	uploadedSet map[int64]bool,
) (*UploadResult, error) {
	concurrency := c.Concurrency
	if concurrency <= 0 {
		concurrency = 4
	}

	var (
		wg           sync.WaitGroup
		uploadedByte int64
		skippedByte  int64
		firstErr     error
		errOnce      sync.Once
		sem          = make(chan struct{}, concurrency)
	)
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 统计跳过的字节数
	for idx := int64(0); idx < totalChunks; idx++ {
		if uploadedSet[idx] {
			offset := idx * chunkSize
			size := chunkSize
			if offset+size > fileSize {
				size = fileSize - offset
			}
			if size > 0 {
				skippedByte += size
			}
		}
	}
	if c.OnProgress != nil && skippedByte > 0 {
		c.OnProgress(skippedByte, fileSize)
	}

	for i := int64(0); i < totalChunks; i++ {
		if uploadedSet[i] {
			continue
		}

		wg.Add(1)
		go func(idx int64) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-cctx.Done():
				return
			}

			offset := idx * chunkSize
			size := chunkSize
			if offset+size > fileSize {
				size = fileSize - offset
			}
			if size <= 0 {
				return
			}

			data := make([]byte, size)
			if _, err := f.ReadAt(data, offset); err != nil && err != io.EOF {
				errOnce.Do(func() {
					firstErr = fmt.Errorf("读取分块 %d 失败: %w", idx, err)
					cancel()
				})
				return
			}

			sum := sha256.Sum256(data)
			hash := hex.EncodeToString(sum[:])

			if err := c.uploadPartWithRetry(cctx, uploadID, idx, size, hash, data); err != nil {
				errOnce.Do(func() {
					firstErr = fmt.Errorf("上传分块 %d 失败: %w", idx, err)
					cancel()
				})
				return
			}

			cur := atomic.AddInt64(&uploadedByte, size) + skippedByte
			if c.OnProgress != nil {
				c.OnProgress(cur, fileSize)
			}
		}(i)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return &UploadResult{
		Uploaded:  uploadedByte,
		Skipped:   skippedByte,
		TotalPart: totalChunks,
	}, nil
}

// ---------- 各接口方法 ----------

func (c *Client) Init(ctx context.Context, fileSize int64) (*InitResponse, error) {
	var out InitResponse
	err := c.doJSON(ctx, http.MethodPost, pathInit, InitRequest{
		Username: c.Username,
		FileSize: fileSize,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Complete(ctx context.Context, uploadID, fileName string, fileSize int64, fileHash string) error {
	return c.doJSON(ctx, http.MethodPost, pathComplete, CompleteRequest{
		Username: c.Username,
		UploadID: uploadID,
		FileName: fileName,
		FileSize: fileSize,
		FileHash: fileHash,
	}, nil)
}

func (c *Client) Cancel(ctx context.Context, uploadID string) error {
	return c.doJSON(ctx, http.MethodPost, pathCancel, CancelRequest{
		Username: c.Username,
		UploadID: uploadID,
	}, nil)
}

func (c *Client) Info(ctx context.Context, uploadID string) (*InfoResponse, error) {
	var out InfoResponse
	err := c.doJSON(ctx, http.MethodPost, pathInfo, InfoRequest{
		Username: c.Username,
		UploadID: uploadID,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) Continue(ctx context.Context, uploadID string) (*ContinueResponse, error) {
	var out ContinueResponse
	err := c.doJSON(ctx, http.MethodPost, pathContinue, ContinueRequest{
		Username: c.Username,
		UploadID: uploadID,
	}, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- 分块上传 ----------

func (c *Client) uploadPartWithRetry(ctx context.Context, uploadID string, index, chunkSize int64, hash string, data []byte) error {
	maxRetries := c.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	delay := c.RetryDelay
	if delay <= 0 {
		delay = 500 * time.Millisecond
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * delay):
			}
		}
		if err := c.uploadPartOnce(ctx, uploadID, index, chunkSize, hash, data); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	return lastErr
}

func (c *Client) uploadPartOnce(ctx context.Context, uploadID string, index, chunkSize int64, hash string, data []byte) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 表单字段与服务端 UploadPartRequest 对齐
	_ = writer.WriteField("username", c.Username)
	_ = writer.WriteField("upload_id", uploadID)
	_ = writer.WriteField("chunk", strconv.FormatInt(chunkSize, 10))
	_ = writer.WriteField("chunk_index", strconv.FormatInt(index, 10))
	_ = writer.WriteField("chunk_hash", hash)

	part, err := writer.CreateFormFile("file", fmt.Sprintf("chunk_%d", index))
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+pathPart, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	c.setAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("解析响应失败: %w, body=%s", err, string(body))
	}
	if apiResp.Code != http.StatusOK {
		return fmt.Errorf("code=%d msg=%s", apiResp.Code, apiResp.Message)
	}
	return nil
}

// ---------- 工具方法 ----------

func (c *Client) doJSON(ctx context.Context, method, path string, reqBody any, out any) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp apiResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("解析响应失败: %w, body=%s", err, string(respBody))
	}
	if apiResp.Code != http.StatusOK {
		return fmt.Errorf("接口 %s 返回: code=%d msg=%s", path, apiResp.Code, apiResp.Message)
	}
	if out != nil && len(apiResp.Data) > 0 && string(apiResp.Data) != "null" {
		if err := json.Unmarshal(apiResp.Data, out); err != nil {
			return fmt.Errorf("解析 data 失败: %w", err)
		}
	}
	return nil
}

func (c *Client) setAuth(req *http.Request) {
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
}

func hashReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func main1() {
	ctx := context.Background()
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InBhbnl1MyIsInVzZXJfaWQiOjIsImV4cCI6MTc4OTQ0NjQ1MCwiaWF0IjoxNzg5MzYwMDUwLCJqdGkiOiIwMWEwOWUyYi05MWZkLTdlMmYtOTYyMi03YWY3ZGM1NTg4Y2EifQ.GkE6HkVZ1wvjEARq20NJ288JYgz2KI5PQ9QXQq5EF_E"
	client := NewClient("http://localhost:8080", "panyu3", token)
	client.Concurrency = 4
	client.MaxRetries = 3
	client.OnProgress = func(uploaded, total int64) {
		pct := float64(uploaded) / float64(total) * 100
		fmt.Printf("\r进度: %.2f%% (%d/%d)", pct, uploaded, total)
	}

	res, err := client.Upload(ctx, "./uploads/cowork.exe")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n上传成功，uploadId=%s，本次上传 %d 字节，跳过 %d 字节\n",
		res.UploadID, res.Uploaded, res.Skipped)

}
