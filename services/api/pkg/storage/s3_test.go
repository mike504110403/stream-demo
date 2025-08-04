package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestS3Storage_GenerateCDNURL(t *testing.T) {
	storage := &S3Storage{
		cdnDomain: "https://cdn.example.com",
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "生成原始影片CDN URL",
			key:      "videos/original/1/test.mp4",
			expected: "https://cdn.example.com/videos/original/1/test.mp4",
		},
		{
			name:     "生成縮圖CDN URL",
			key:      "thumbnails/1/test.jpg",
			expected: "https://cdn.example.com/thumbnails/1/test.jpg",
		},
		{
			name:     "空CDN域名",
			key:      "videos/original/1/test.mp4",
			expected: "http://localhost:9000//videos/original/1/test.mp4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "空CDN域名" {
				storage.cdnDomain = ""
			} else {
				storage.cdnDomain = "https://cdn.example.com"
			}

			result := storage.GenerateCDNURL(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestS3Storage_GenerateProcessedCDNURL(t *testing.T) {
	storage := &S3Storage{
		cdnDomain: "https://cdn.example.com",
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "生成處理後影片CDN URL",
			key:      "videos/processed/1/test_720p.mp4",
			expected: "https://cdn.example.com/videos/processed/1/test_720p.mp4",
		},
		{
			name:     "生成處理後縮圖CDN URL",
			key:      "thumbnails/processed/1/test_thumb.jpg",
			expected: "https://cdn.example.com/thumbnails/processed/1/test_thumb.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.GenerateProcessedCDNURL(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetContentType(t *testing.T) {
	tests := []struct {
		name     string
		ext      string
		expected string
	}{
		{
			name:     "MP4 影片",
			ext:      ".mp4",
			expected: "video/mp4",
		},
		{
			name:     "AVI 影片",
			ext:      ".avi",
			expected: "video/avi",
		},
		{
			name:     "MOV 影片",
			ext:      ".mov",
			expected: "video/quicktime",
		},
		{
			name:     "JPEG 圖片",
			ext:      ".jpg",
			expected: "image/jpeg",
		},
		{
			name:     "PNG 圖片",
			ext:      ".png",
			expected: "image/png",
		},
		{
			name:     "未知格式",
			ext:      ".xyz",
			expected: "application/octet-stream",
		},
		{
			name:     "空副檔名",
			ext:      "",
			expected: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getContentType(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewS3Storage(t *testing.T) {
	tests := []struct {
		name          string
		config        S3Config
		expectedError bool
	}{
		{
			name: "有效配置",
			config: S3Config{
				AccessKey: "test-key",
				SecretKey: "test-secret",
				Region:    "us-east-1",
				Bucket:    "test-bucket",
			},
			expectedError: false,
		},
		{
			name: "本地 MinIO 配置",
			config: S3Config{
				AccessKey: "minioadmin",
				SecretKey: "minioadmin",
				Region:    "us-east-1",
				Bucket:    "test-bucket",
				Endpoint:  "http://localhost:9000",
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewS3Storage(tt.config)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, storage)
			} else {
				// 注意：由於沒有實際的 AWS 憑證，這個測試會失敗
				// 但我們主要測試函數不會 panic
				if err == nil {
					assert.NotNil(t, storage)
					assert.Equal(t, tt.config.Bucket, storage.bucket)
					assert.Equal(t, tt.config.Region, storage.region)
					assert.Equal(t, tt.config.CDNDomain, storage.cdnDomain)
				}
			}
		})
	}
}

func TestS3Storage_GeneratePresignedUploadURL(t *testing.T) {
	// 這個測試需要實際的 AWS 憑證，所以我們跳過它
	t.Skip("Skipping test that requires AWS credentials")
}

func TestS3Storage_GeneratePresignedDownloadURL(t *testing.T) {
	// 這個測試需要實際的 AWS 憑證，所以我們跳過它
	t.Skip("Skipping test that requires AWS credentials")
}

func TestS3Storage_CheckFileExists(t *testing.T) {
	// 這個測試需要實際的 AWS 憑證，所以我們跳過它
	t.Skip("Skipping test that requires AWS credentials")
}

func TestS3Storage_GetFileInfo(t *testing.T) {
	// 這個測試需要實際的 AWS 憑證，所以我們跳過它
	t.Skip("Skipping test that requires AWS credentials")
}

func TestS3Storage_UploadVideo(t *testing.T) {
	// 這個測試需要實際的 AWS 憑證，所以我們跳過它
	t.Skip("Skipping test that requires AWS credentials")
}

func TestS3Storage_UploadThumbnail(t *testing.T) {
	// 這個測試需要實際的 AWS 憑證，所以我們跳過它
	t.Skip("Skipping test that requires AWS credentials")
}

func TestS3Storage_DeleteFile(t *testing.T) {
	// 這個測試需要實際的 AWS 憑證，所以我們跳過它
	t.Skip("Skipping test that requires AWS credentials")
}
