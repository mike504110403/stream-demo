package storage

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

// S3Config S3配置
type S3Config struct {
	AccessKey string
	SecretKey string
	Region    string
	Bucket    string
	Endpoint  string
	CDNDomain string
}

// S3Storage S3儲存服務
type S3Storage struct {
	client    *s3.S3
	bucket    string
	region    string
	cdnDomain string
}

// PresignedUploadURL 預簽名上傳URL響應
type PresignedUploadURL struct {
	UploadURL string            `json:"upload_url"`
	FormData  map[string]string `json:"form_data"`
	Key       string            `json:"key"`
	CDNUrl    string            `json:"cdn_url"`
}

// NewS3Storage 創建S3儲存服務
func NewS3Storage(config S3Config) (*S3Storage, error) {
	awsConfig := &aws.Config{
		Region: aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(
			config.AccessKey,
			config.SecretKey,
			"",
		),
	}

	// 如果有自定義endpoint，設置它（MinIO 需要特殊配置）
	if config.Endpoint != "" {
		awsConfig.Endpoint = aws.String(config.Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true) // MinIO 需要路徑樣式
		awsConfig.DisableSSL = aws.Bool(true)       // 本地開發不使用 SSL
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("建立 AWS session 失敗: %w", err)
	}

	return &S3Storage{
		client:    s3.New(sess),
		bucket:    config.Bucket,
		region:    config.Region,
		cdnDomain: config.CDNDomain,
	}, nil
}

// GeneratePresignedUploadURL 生成預簽名上傳URL
func (s *S3Storage) GeneratePresignedUploadURL(userID uint, fileExt string, fileSize int64) (*PresignedUploadURL, error) {
	// 生成唯一檔名
	key := fmt.Sprintf("videos/original/%d/%s%s", userID, uuid.New().String(), fileExt)

	// 創建預簽名請求（簡化版本，不預設複雜 headers）
	req, _ := s.client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		// 不預設 ContentType, ACL 等，避免簽名問題
	})

	// 生成預簽名URL，15分鐘有效
	urlStr, err := req.Presign(15 * time.Minute)
	if err != nil {
		return nil, fmt.Errorf("生成預簽名URL失敗: %w", err)
	}

	// 調試輸出
	fmt.Printf("🔧 生成預簽名URL - Key: %s\n", key)
	fmt.Printf("🔧 生成預簽名URL - URL: %s\n", urlStr)
	fmt.Printf("🔧 生成預簽名URL - ContentType: %s\n", getContentType(fileExt))

	// 生成CDN URL
	cdnURL := s.GenerateCDNURL(key)

	return &PresignedUploadURL{
		UploadURL: urlStr,
		Key:       key,
		CDNUrl:    cdnURL,
		FormData: map[string]string{
			"Content-Type": getContentType(fileExt),
		},
	}, nil
}

// GeneratePresignedDownloadURL 生成預簽名下載URL（用於私有檔案）
func (s *S3Storage) GeneratePresignedDownloadURL(key string, expiration time.Duration) (string, error) {
	req, _ := s.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(expiration)
	if err != nil {
		return "", fmt.Errorf("生成預簽名下載URL失敗: %w", err)
	}

	return urlStr, nil
}

// GenerateCDNURL 生成CDN URL
func (s *S3Storage) GenerateCDNURL(key string) string {
	if s.cdnDomain != "" {
		return fmt.Sprintf("%s/%s", s.cdnDomain, key)
	}
	// 如果沒有CDN，返回 MinIO URL（本地開發）
	return fmt.Sprintf("http://localhost:9000/%s/%s", s.bucket, key)
}

// GenerateProcessedCDNURL 生成處理後檔案的 CDN URL
func (s *S3Storage) GenerateProcessedCDNURL(key string) string {
	processedBucket := "stream-demo-processed"
	if s.cdnDomain != "" {
		return fmt.Sprintf("%s/%s", s.cdnDomain, key)
	}
	// 如果沒有CDN，返回處理後桶的 MinIO URL（本地開發）
	return fmt.Sprintf("http://localhost:9000/%s/%s", processedBucket, key)
}

// CheckFileExists 檢查檔案是否存在
func (s *S3Storage) CheckFileExists(key string) (bool, error) {
	_, err := s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return false, nil // 檔案不存在
	}

	return true, nil
}

// GetFileInfo 獲取檔案資訊
func (s *S3Storage) GetFileInfo(key string) (*s3.HeadObjectOutput, error) {
	return s.client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
}

// UploadVideo 上傳影片到S3（保留舊方法以兼容）
func (s *S3Storage) UploadVideo(file multipart.File, header *multipart.FileHeader, userID uint) (string, error) {
	// 生成唯一檔名
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("videos/original/%d/%s%s", userID, uuid.New().String(), ext)

	// 讀取檔案內容
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		return "", fmt.Errorf("讀取檔案失敗: %w", err)
	}

	// 上傳到S3
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(filename),
		Body:          bytes.NewReader(buffer.Bytes()),
		ContentLength: aws.Int64(int64(buffer.Len())),
		ContentType:   aws.String(getContentType(ext)),
		ACL:           aws.String("private"),
	})
	if err != nil {
		return "", fmt.Errorf("上傳到S3失敗: %w", err)
	}

	return s.GenerateCDNURL(filename), nil
}

// UploadThumbnail 上傳縮圖到S3
func (s *S3Storage) UploadThumbnail(file multipart.File, header *multipart.FileHeader, userID uint) (string, error) {
	// 生成唯一檔名
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("thumbnails/%d/%s%s", userID, uuid.New().String(), ext)

	// 讀取檔案內容
	buffer := bytes.NewBuffer(nil)
	if _, err := io.Copy(buffer, file); err != nil {
		return "", fmt.Errorf("讀取檔案失敗: %w", err)
	}

	// 上傳到S3
	_, err := s.client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(filename),
		Body:          bytes.NewReader(buffer.Bytes()),
		ContentLength: aws.Int64(int64(buffer.Len())),
		ContentType:   aws.String("image/jpeg"),
		ACL:           aws.String("public-read"),
	})
	if err != nil {
		return "", fmt.Errorf("上傳縮圖到S3失敗: %w", err)
	}

	return s.GenerateCDNURL(filename), nil
}

// DeleteFile 刪除S3檔案
func (s *S3Storage) DeleteFile(key string) error {
	_, err := s.client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// getContentType 根據檔案擴展名返回Content-Type
func getContentType(ext string) string {
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/avi"
	case ".mov":
		return "video/quicktime"
	case ".mkv":
		return "video/x-matroska"
	case ".webm":
		return "video/webm"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	default:
		return "application/octet-stream"
	}
}
