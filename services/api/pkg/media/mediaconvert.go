package media

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
)

// MediaConvertConfig MediaConvert配置
type MediaConvertConfig struct {
	Region       string
	Endpoint     string
	AccessKey    string
	SecretKey    string
	RoleArn      string
	OutputBucket string
}

// MediaConvertService MediaConvert服務
type MediaConvertService struct {
	client       *mediaconvert.MediaConvert
	roleArn      string
	outputBucket string
}

// TranscodeJob 轉碼任務配置
type TranscodeJob struct {
	JobID        string
	InputKey     string
	OutputPrefix string
	UserID       uint
	VideoID      uint
}

// QualityPreset 影片品質預設
type QualityPreset struct {
	Name    string
	Width   int64
	Height  int64
	Bitrate int64
}

// NewMediaConvertService 創建MediaConvert服務
func NewMediaConvertService(config MediaConvertConfig) (*MediaConvertService, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials(
			config.AccessKey,
			config.SecretKey,
			"",
		),
		Endpoint: aws.String(config.Endpoint),
	})
	if err != nil {
		return nil, fmt.Errorf("建立 AWS session 失敗: %w", err)
	}

	return &MediaConvertService{
		client:       mediaconvert.New(sess),
		roleArn:      config.RoleArn,
		outputBucket: config.OutputBucket,
	}, nil
}

// CreateHLSTranscodeJob 創建HLS轉碼任務
func (mc *MediaConvertService) CreateHLSTranscodeJob(inputKey string, userID, videoID uint) (*TranscodeJob, error) {
	// 生成輸出路徑前綴
	outputPrefix := fmt.Sprintf("videos/processed/%d/%d", userID, videoID)

	// 構建輸入設定
	input := &mediaconvert.Input{
		FileInput: aws.String(fmt.Sprintf("s3://%s/%s", mc.outputBucket, inputKey)),
	}

	// 構建HLS輸出群組
	hlsOutputGroup := &mediaconvert.OutputGroup{
		Name: aws.String("HLS"),
		OutputGroupSettings: &mediaconvert.OutputGroupSettings{
			Type: aws.String("HLS_GROUP_SETTINGS"),
			HlsGroupSettings: &mediaconvert.HlsGroupSettings{
				Destination: aws.String(fmt.Sprintf("s3://%s/%s/", mc.outputBucket, outputPrefix)),

				ManifestDurationFormat: aws.String("INTEGER"),
				OutputSelection:        aws.String("MANIFESTS_AND_SEGMENTS"),
				ProgramDateTimePeriod:  aws.Int64(600),
				SegmentLength:          aws.Int64(10),
				MinSegmentLength:       aws.Int64(0),
			},
		},
	}

	// 定義不同品質的輸出
	qualities := []QualityPreset{
		{Name: "720p", Width: 1280, Height: 720, Bitrate: 2500000},
		{Name: "480p", Width: 854, Height: 480, Bitrate: 1200000},
		{Name: "360p", Width: 640, Height: 360, Bitrate: 800000},
	}

	// 為每個品質創建輸出
	var outputs []*mediaconvert.Output
	for _, quality := range qualities {
		output := &mediaconvert.Output{
			NameModifier: aws.String(fmt.Sprintf("_%s", quality.Name)),
			VideoDescription: &mediaconvert.VideoDescription{
				Width:  aws.Int64(quality.Width),
				Height: aws.Int64(quality.Height),
				CodecSettings: &mediaconvert.VideoCodecSettings{
					Codec: aws.String("H_264"),
					H264Settings: &mediaconvert.H264Settings{
						RateControlMode:  aws.String("CBR"),
						Bitrate:          aws.Int64(quality.Bitrate / 1000), // 轉換為 kbps
						FramerateControl: aws.String("INITIALIZE_FROM_SOURCE"),
					},
				},
			},
			AudioDescriptions: []*mediaconvert.AudioDescription{
				{
					CodecSettings: &mediaconvert.AudioCodecSettings{
						Codec: aws.String("AAC"),
						AacSettings: &mediaconvert.AacSettings{
							Bitrate:    aws.Int64(128000),
							SampleRate: aws.Int64(48000),
							CodingMode: aws.String("CODING_MODE_2_0"),
						},
					},
				},
			},
			ContainerSettings: &mediaconvert.ContainerSettings{
				Container: aws.String("M3U8"),
			},
		}
		outputs = append(outputs, output)
	}

	hlsOutputGroup.Outputs = outputs

	// 創建縮圖輸出群組
	thumbnailOutputGroup := &mediaconvert.OutputGroup{
		Name: aws.String("Thumbnails"),
		OutputGroupSettings: &mediaconvert.OutputGroupSettings{
			Type: aws.String("FILE_GROUP_SETTINGS"),
			FileGroupSettings: &mediaconvert.FileGroupSettings{
				Destination: aws.String(fmt.Sprintf("s3://%s/%s/thumbnails/", mc.outputBucket, outputPrefix)),
			},
		},
		Outputs: []*mediaconvert.Output{
			{
				NameModifier: aws.String("_thumb"),
				VideoDescription: &mediaconvert.VideoDescription{
					Width:  aws.Int64(320),
					Height: aws.Int64(240),
					CodecSettings: &mediaconvert.VideoCodecSettings{
						Codec: aws.String("FRAME_CAPTURE"),
						FrameCaptureSettings: &mediaconvert.FrameCaptureSettings{
							FramerateNumerator:   aws.Int64(1),
							FramerateDenominator: aws.Int64(10), // 每10秒一張
						},
					},
				},
				ContainerSettings: &mediaconvert.ContainerSettings{
					Container: aws.String("RAW"),
				},
			},
		},
	}

	// 構建轉碼任務
	jobInput := &mediaconvert.CreateJobInput{
		Role: aws.String(mc.roleArn),
		Settings: &mediaconvert.JobSettings{
			Inputs:       []*mediaconvert.Input{input},
			OutputGroups: []*mediaconvert.OutputGroup{hlsOutputGroup, thumbnailOutputGroup},
		},
		Queue: aws.String("Default"),
		UserMetadata: map[string]*string{
			"user_id":  aws.String(fmt.Sprintf("%d", userID)),
			"video_id": aws.String(fmt.Sprintf("%d", videoID)),
		},
	}

	// 提交任務
	result, err := mc.client.CreateJob(jobInput)
	if err != nil {
		return nil, fmt.Errorf("創建轉碼任務失敗: %w", err)
	}

	return &TranscodeJob{
		JobID:        *result.Job.Id,
		InputKey:     inputKey,
		OutputPrefix: outputPrefix,
		UserID:       userID,
		VideoID:      videoID,
	}, nil
}

// GetJobStatus 獲取任務狀態
func (mc *MediaConvertService) GetJobStatus(jobID string) (*mediaconvert.Job, error) {
	input := &mediaconvert.GetJobInput{
		Id: aws.String(jobID),
	}

	result, err := mc.client.GetJob(input)
	if err != nil {
		return nil, fmt.Errorf("獲取任務狀態失敗: %w", err)
	}

	return result.Job, nil
}

// ListJobs 列出任務
func (mc *MediaConvertService) ListJobs(status string, maxResults int64) ([]*mediaconvert.Job, error) {
	input := &mediaconvert.ListJobsInput{
		MaxResults: aws.Int64(maxResults),
		Order:      aws.String("DESCENDING"),
	}

	if status != "" {
		input.Status = aws.String(status)
	}

	result, err := mc.client.ListJobs(input)
	if err != nil {
		return nil, fmt.Errorf("列出任務失敗: %w", err)
	}

	return result.Jobs, nil
}
