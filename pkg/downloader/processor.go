package downloader

import (
	"fmt"
	"io"

	"github.com/xpzouying/xiaohongshu-mcp/configs"
)

// ImageProcessor 图片处理器
type ImageProcessor struct {
	downloader *ImageDownloader
}

// NewImageProcessor 创建图片处理器
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		downloader: NewImageDownloader(configs.GetImagesPath()),
	}
}

// ProcessImages 处理图片列表，返回本地文件路径
// 支持两种输入格式：
// 1. URL格式 (http/https开头) - 自动下载到本地
// 2. 本地文件路径 - 直接使用
// 保持原始图片顺序，如果下载失败直接返回错误
func (p *ImageProcessor) ProcessImages(images []string) ([]string, error) {
	localPaths := make([]string, 0, len(images))

	// 按顺序处理每张图片
	for _, image := range images {
		if IsImageURL(image) {
			// URL图片：立即下载，失败直接返回错误
			localPath, err := p.downloader.DownloadImage(image)
			if err != nil {
				return nil, fmt.Errorf("下载图片失败 %s: %w", image, err)
			}
			localPaths = append(localPaths, localPath)
		} else {
			// 本地路径直接使用
			localPaths = append(localPaths, image)
		}
	}

	if len(localPaths) == 0 {
		return nil, fmt.Errorf("no valid images found")
	}

	return localPaths, nil
}

// VideoProcessor 视频处理器
type VideoProcessor struct {
	downloader *VideoDownloader
}

// NewVideoProcessor 创建视频处理器
func NewVideoProcessor() *VideoProcessor {
	return &VideoProcessor{
		downloader: NewVideoDownloader(configs.GetVideosPath()),
	}
}

// ProcessVideo 处理视频，返回本地文件路径
// 支持两种输入格式：
// 1. URL格式 (http/https开头) - 自动下载到本地
// 2. 本地文件路径 - 直接使用
func (p *VideoProcessor) ProcessVideo(video string) (string, error) {
	if IsVideoURL(video) {
		// URL视频：下载到本地
		localPath, err := p.downloader.DownloadVideo(video)
		if err != nil {
			return "", fmt.Errorf("下载视频失败 %s: %w", video, err)
		}
		return localPath, nil
	}
	// 本地路径直接使用
	return video, nil
}

// SaveUploadedVideo 保存上传的视频文件
func (p *VideoProcessor) SaveUploadedVideo(reader io.Reader, filename string) (string, error) {
	return p.downloader.SaveUploadedVideo(reader, filename)
}

// GetVideosPath 获取视频存储路径
func (p *VideoProcessor) GetVideosPath() string {
	return p.downloader.GetSavePath()
}
