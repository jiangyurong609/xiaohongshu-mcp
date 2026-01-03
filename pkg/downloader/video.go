package downloader

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/h2non/filetype"
	"github.com/pkg/errors"
)

// VideoDownloader handles video downloads and file management
type VideoDownloader struct {
	savePath   string
	httpClient *http.Client
}

// NewVideoDownloader creates a video downloader
func NewVideoDownloader(savePath string) *VideoDownloader {
	// Ensure save directory exists
	if err := os.MkdirAll(savePath, 0755); err != nil {
		panic(fmt.Sprintf("failed to create video save path: %v", err))
	}

	return &VideoDownloader{
		savePath: savePath,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute, // Videos can be large, longer timeout
		},
	}
}

// DownloadVideo downloads a video from URL and returns local path
func (d *VideoDownloader) DownloadVideo(videoURL string) (string, error) {
	// Validate URL format
	if !d.isValidVideoURL(videoURL) {
		return "", errors.New("invalid video URL format")
	}

	// Download video data
	resp, err := d.httpClient.Get(videoURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to download video")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Generate unique filename
	fileName := d.generateFileName(videoURL, "mp4") // Default to mp4
	filePath := filepath.Join(d.savePath, fileName)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", errors.Wrap(err, "failed to create video file")
	}
	defer file.Close()

	// Stream copy to file (memory efficient for large videos)
	written, err := io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(filePath) // Clean up partial file
		return "", errors.Wrap(err, "failed to save video")
	}

	// Validate video file
	if written == 0 {
		os.Remove(filePath)
		return "", errors.New("downloaded video is empty")
	}

	// Detect actual file type and rename if needed
	actualPath, err := d.validateAndRename(filePath)
	if err != nil {
		os.Remove(filePath)
		return "", err
	}

	return actualPath, nil
}

// SaveUploadedVideo saves an uploaded video from reader to local storage
func (d *VideoDownloader) SaveUploadedVideo(reader io.Reader, originalName string) (string, error) {
	// Generate unique filename
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".mp4"
	}

	timestamp := time.Now().UnixNano()
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s_%d", originalName, timestamp)))
	hashStr := fmt.Sprintf("%x", hash)[:16]
	fileName := fmt.Sprintf("video_%s_%d%s", hashStr, timestamp, ext)
	filePath := filepath.Join(d.savePath, fileName)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return "", errors.Wrap(err, "failed to create video file")
	}
	defer file.Close()

	// Copy data to file
	written, err := io.Copy(file, reader)
	if err != nil {
		os.Remove(filePath)
		return "", errors.Wrap(err, "failed to save video")
	}

	if written == 0 {
		os.Remove(filePath)
		return "", errors.New("uploaded video is empty")
	}

	return filePath, nil
}

// isValidVideoURL checks if URL is valid
func (d *VideoDownloader) isValidVideoURL(rawURL string) bool {
	if !strings.HasPrefix(strings.ToLower(rawURL), "http://") &&
		!strings.HasPrefix(strings.ToLower(rawURL), "https://") {
		return false
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	return parsedURL.Scheme != "" && parsedURL.Host != ""
}

// generateFileName generates a unique filename for video
func (d *VideoDownloader) generateFileName(videoURL, extension string) string {
	hash := sha256.Sum256([]byte(videoURL))
	hashStr := fmt.Sprintf("%x", hash)
	shortHash := hashStr[:16]
	timestamp := time.Now().Unix()
	return fmt.Sprintf("video_%s_%d.%s", shortHash, timestamp, extension)
}

// validateAndRename validates video type and renames with correct extension
func (d *VideoDownloader) validateAndRename(filePath string) (string, error) {
	// Read first 8KB to detect file type
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	head := make([]byte, 8192)
	n, err := file.Read(head)
	file.Close()
	if err != nil && err != io.EOF {
		return "", err
	}

	kind, err := filetype.Match(head[:n])
	if err != nil {
		return "", errors.Wrap(err, "failed to detect video type")
	}

	// Check if it's a valid video
	if !filetype.IsVideo(head[:n]) {
		return "", errors.New("uploaded file is not a valid video")
	}

	// Rename with correct extension if needed
	if kind.Extension != "" {
		dir := filepath.Dir(filePath)
		base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		newPath := filepath.Join(dir, base+"."+kind.Extension)

		if newPath != filePath {
			if err := os.Rename(filePath, newPath); err != nil {
				return filePath, nil // Keep original if rename fails
			}
			return newPath, nil
		}
	}

	return filePath, nil
}

// CleanupOldVideos removes videos older than specified duration
func (d *VideoDownloader) CleanupOldVideos(maxAge time.Duration) error {
	entries, err := os.ReadDir(d.savePath)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-maxAge)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(d.savePath, entry.Name()))
		}
	}

	return nil
}

// IsVideoURL checks if string is a video URL
func IsVideoURL(path string) bool {
	return strings.HasPrefix(strings.ToLower(path), "http://") ||
		strings.HasPrefix(strings.ToLower(path), "https://")
}

// GetSavePath returns the video save directory
func (d *VideoDownloader) GetSavePath() string {
	return d.savePath
}
