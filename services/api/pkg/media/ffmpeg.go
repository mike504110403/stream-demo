package media

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// FFmpegConfig FFmpeg 轉碼配置
type FFmpegConfig struct {
	ContainerName string // Docker 容器名稱
	Enabled       bool   // 是否啟用 FFmpeg 轉碼
}

// FFmpegService 本地 FFmpeg 轉碼服務
type FFmpegService struct {
	containerName string
	enabled       bool
	jobs          map[string]*FFmpegTranscodeJob // 任務狀態管理
	mutex         sync.RWMutex
}

// FFmpegTranscodeJob FFmpeg 轉碼任務
type FFmpegTranscodeJob struct {
	JobID        string    `json:"job_id"`
	InputKey     string    `json:"input_key"`
	OutputPrefix string    `json:"output_prefix"`
	UserID       uint      `json:"user_id"`
	VideoID      uint      `json:"video_id"`
	Status       string    `json:"status"`
	StartedAt    time.Time `json:"started_at"`
	CompletedAt  time.Time `json:"completed_at,omitempty"`
	Error        string    `json:"error,omitempty"`
}

// TranscodeReport 轉碼報告
type TranscodeReport struct {
	Status       string                 `json:"status"`
	InputFile    string                 `json:"input_file"`
	OutputPrefix string                 `json:"output_prefix"`
	OriginalInfo map[string]interface{} `json:"original_info"`
	Outputs      map[string]string      `json:"outputs"`
	Qualities    []string               `json:"qualities"`
	CompletedAt  string                 `json:"completed_at"`
}

// NewFFmpegService 創建 FFmpeg 轉碼服務
func NewFFmpegService(config FFmpegConfig) *FFmpegService {
	return &FFmpegService{
		containerName: config.ContainerName,
		enabled:       config.Enabled,
		jobs:          make(map[string]*FFmpegTranscodeJob),
	}
}

// CreateHLSTranscodeJob 創建 HLS 轉碼任務
func (fs *FFmpegService) CreateHLSTranscodeJob(inputKey string, userID, videoID uint) (*FFmpegTranscodeJob, error) {
	if !fs.enabled {
		return nil, fmt.Errorf("FFmpeg 轉碼服務未啟用")
	}

	// 生成任務 ID
	jobID := fmt.Sprintf("ffmpeg_%d_%d_%d", userID, videoID, time.Now().Unix())

	// 生成輸出路徑前綴
	outputPrefix := fmt.Sprintf("videos/processed/%d/%d", userID, videoID)

	fmt.Printf("🎬 創建 FFmpeg 轉碼任務 - JobID: %s, InputKey: %s, OutputPrefix: %s\n", jobID, inputKey, outputPrefix)

	// 創建轉碼任務
	job := &FFmpegTranscodeJob{
		JobID:        jobID,
		InputKey:     inputKey,
		OutputPrefix: outputPrefix,
		UserID:       userID,
		VideoID:      videoID,
		Status:       "SUBMITTED",
		StartedAt:    time.Now(),
	}

	// 註冊任務
	fs.mutex.Lock()
	fs.jobs[jobID] = job
	fs.mutex.Unlock()

	// 異步執行轉碼
	go fs.executeTranscode(job)

	return job, nil
}

// executeTranscode 執行轉碼任務
func (fs *FFmpegService) executeTranscode(job *FFmpegTranscodeJob) {
	fmt.Printf("🚀 開始執行 FFmpeg 轉碼 - JobID: %s\n", job.JobID)

	// 更新狀態為處理中
	fs.mutex.Lock()
	job.Status = "PROGRESSING"
	fs.mutex.Unlock()

	// 執行 Docker 命令調用轉碼腳本
	cmd := exec.Command("docker", "exec", fs.containerName, "/scripts/transcode.sh",
		job.InputKey,
		job.OutputPrefix,
		fmt.Sprintf("%d", job.UserID),
		fmt.Sprintf("%d", job.VideoID),
	)

	fmt.Printf("🔧 執行轉碼命令: docker exec %s /scripts/transcode.sh %s %s %d %d\n",
		fs.containerName, job.InputKey, job.OutputPrefix, job.UserID, job.VideoID)

	// 執行命令
	output, err := cmd.CombinedOutput()

	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if err != nil {
		job.Status = "ERROR"
		job.Error = fmt.Sprintf("轉碼失敗: %v, 輸出: %s", err, string(output))
		job.CompletedAt = time.Now()
		fmt.Printf("❌ 轉碼任務失敗 [%s]: %s\n", job.JobID, job.Error)
		return
	}

	// 轉碼成功
	job.Status = "COMPLETE"
	job.CompletedAt = time.Now()
	fmt.Printf("✅ 轉碼任務完成 [%s]: %s\n", job.JobID, string(output))
}

// GetJobStatus 獲取任務狀態
func (fs *FFmpegService) GetJobStatus(jobID string) (*FFmpegTranscodeJob, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	job, exists := fs.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("任務不存在: %s", jobID)
	}

	return job, nil
}

// GetTranscodeReport 獲取轉碼報告
func (fs *FFmpegService) GetTranscodeReport(outputPrefix string) (*TranscodeReport, error) {
	if !fs.enabled {
		return nil, fmt.Errorf("FFmpeg 轉碼服務未啟用")
	}

	// 從 MinIO 下載轉碼報告（從處理後桶）
	cmd := exec.Command("docker", "exec", fs.containerName, "mc", "cp",
		fmt.Sprintf("s3/stream-demo-processed/%s/transcode_report.json", outputPrefix),
		"/tmp/report.json",
	)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("無法獲取轉碼報告: %v", err)
	}

	// 讀取報告內容
	catCmd := exec.Command("docker", "exec", fs.containerName, "cat", "/tmp/report.json")
	output, err := catCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("無法讀取轉碼報告: %v", err)
	}

	// 解析 JSON
	var report TranscodeReport
	if err := json.Unmarshal(output, &report); err != nil {
		return nil, fmt.Errorf("無法解析轉碼報告: %v", err)
	}

	return &report, nil
}

// IsEnabled 檢查 FFmpeg 服務是否啟用
func (fs *FFmpegService) IsEnabled() bool {
	return fs.enabled
}

// TestConnection 測試與 FFmpeg 容器的連接
func (fs *FFmpegService) TestConnection() error {
	if !fs.enabled {
		return fmt.Errorf("FFmpeg 轉碼服務未啟用")
	}

	// 測試容器是否可訪問
	cmd := exec.Command("docker", "exec", fs.containerName, "ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("無法連接到 FFmpeg 容器: %v", err)
	}

	return nil
}

// CleanupCompletedJobs 清理已完成的任務（可選）
func (fs *FFmpegService) CleanupCompletedJobs() {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour) // 保留24小時內的任務

	for jobID, job := range fs.jobs {
		if job.Status == "COMPLETE" || job.Status == "ERROR" {
			if job.CompletedAt.Before(cutoff) {
				delete(fs.jobs, jobID)
			}
		}
	}
}
