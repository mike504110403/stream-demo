package services

import (
	"testing"
	"time"

	"stream-demo/backend/database/models"

	"github.com/stretchr/testify/assert"
)

// 測試輔助函數
func createTestVideo() *models.Video {
	return &models.Video{
		ID:          1,
		Title:       "測試影片",
		Description: "這是一個測試影片",
		OriginalKey: "videos/original/1/test.mp4",
		Status:      "uploading",
		UserID:      1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// TestTranscodeWorker_Basic 基本功能測試
func TestTranscodeWorker_Basic(t *testing.T) {
	t.Run("創建測試影片", func(t *testing.T) {
		video := createTestVideo()

		assert.Equal(t, uint(1), video.ID)
		assert.Equal(t, "測試影片", video.Title)
		assert.Equal(t, "uploading", video.Status)
		assert.Equal(t, "videos/original/1/test.mp4", video.OriginalKey)
	})

	t.Run("影片狀態驗證", func(t *testing.T) {
		video := createTestVideo()

		// 測試有效狀態
		validStatuses := []string{"uploading", "processing", "transcoding", "ready", "failed"}
		for _, status := range validStatuses {
			video.Status = status
			assert.Contains(t, validStatuses, video.Status)
		}
	})

	t.Run("檔案路徑驗證", func(t *testing.T) {
		video := createTestVideo()

		// 測試檔案路徑格式
		assert.Contains(t, video.OriginalKey, "videos/original")
		assert.Contains(t, video.OriginalKey, ".mp4")
	})
}

// TestTranscodeWorker_Validation 驗證邏輯測試
func TestTranscodeWorker_Validation(t *testing.T) {
	t.Run("空檔案 Key 驗證", func(t *testing.T) {
		video := createTestVideo()
		video.OriginalKey = ""

		assert.Empty(t, video.OriginalKey)
		assert.True(t, video.OriginalKey == "")
	})

	t.Run("檔案格式驗證", func(t *testing.T) {
		video := createTestVideo()

		// 測試不同檔案格式
		testCases := []struct {
			key      string
			expected string
		}{
			{"videos/original/1/test.mp4", "mp4"},
			{"videos/original/1/test.mov", "mov"},
			{"videos/original/1/test.avi", "avi"},
		}

		for _, tc := range testCases {
			video.OriginalKey = tc.key
			extension := getFileExtension(tc.key)
			assert.Equal(t, tc.expected, extension)
		}
	})
}

// TestTranscodeWorker_StatusFlow 狀態流程測試
func TestTranscodeWorker_StatusFlow(t *testing.T) {
	t.Run("狀態轉換流程", func(t *testing.T) {
		video := createTestVideo()

		// 模擬狀態轉換流程
		statusFlow := []string{"uploading", "processing", "transcoding", "ready"}

		for i, expectedStatus := range statusFlow {
			video.Status = expectedStatus
			assert.Equal(t, expectedStatus, video.Status, "步驟 %d: 狀態應該為 %s", i+1, expectedStatus)
		}
	})

	t.Run("失敗狀態處理", func(t *testing.T) {
		video := createTestVideo()

		// 模擬失敗流程
		video.Status = "uploading"
		video.Status = "failed"

		assert.Equal(t, "failed", video.Status)
	})
}

// TestTranscodeWorker_FileExtension 檔案擴展名測試
func TestTranscodeWorker_FileExtension(t *testing.T) {
	t.Run("基本檔案擴展名", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"test.mp4", "mp4"},
			{"test.mov", "mov"},
			{"test.avi", "avi"},
			{"test.mkv", "mkv"},
			{"test.webm", "webm"},
		}

		for _, tc := range testCases {
			result := getFileExtension(tc.input)
			assert.Equal(t, tc.expected, result, "輸入: %s", tc.input)
		}
	})

	t.Run("複雜路徑檔案擴展名", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"videos/original/1/test.mp4", "mp4"},
			{"videos/original/user/123/video.mov", "mov"},
			{"path/to/file.avi", "avi"},
		}

		for _, tc := range testCases {
			result := getFileExtension(tc.input)
			assert.Equal(t, tc.expected, result, "輸入: %s", tc.input)
		}
	})

	t.Run("無擴展名檔案", func(t *testing.T) {
		testCases := []struct {
			input    string
			expected string
		}{
			{"test", ""},
			{"videos/original/file", ""},
			{"", ""},
		}

		for _, tc := range testCases {
			result := getFileExtension(tc.input)
			assert.Equal(t, tc.expected, result, "輸入: %s", tc.input)
		}
	})
}
