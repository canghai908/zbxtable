package model

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
	"zbxtable/pkg/logger"

	"github.com/google/uuid"
)

// UploadResult 上传结果
type UploadResult struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// ensureDir 确保目录存在，不存在则创建
func ensureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Log.Error("Failed to create directory:", dir, err)
			return fmt.Errorf("目录创建失败")
		}
		logger.Log.Info("Created directory:", dir)
	}
	return nil
}

// ValidateImageFile 验证图片文件
func ValidateImageFile(file *multipart.FileHeader) error {
	// 检查文件大小
	if file.Size > MaxFileSize {
		return fmt.Errorf("文件大小超过限制，最大允许 %dMB", MaxFileSize/(1024*1024))
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !strings.Contains(AllowedImageExts, ext) {
		return fmt.Errorf("不支持的文件格式，仅支持: %s", AllowedImageExts)
	}

	return nil
}

// GenerateUniqueFilename 生成唯一的文件名
func GenerateUniqueFilename(ext string) string {
	// 使用时间戳 + UUID 生成唯一文件名
	timestamp := time.Now().Format("20060102150405")
	uniqueID := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%s%s", timestamp, uniqueID, ext)
}

func saveImageToDir(file *multipart.FileHeader, diskDir string, urlPrefix string) (*UploadResult, error) {
	// 验证文件
	if err := ValidateImageFile(file); err != nil {
		return nil, err
	}

	// 确保目录存在
	if err := ensureDir(diskDir); err != nil {
		return nil, err
	}

	// 生成唯一文件名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	filename := GenerateUniqueFilename(ext)
	fullPath := filepath.Join(diskDir, filename)

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		logger.Log.Error("Failed to open uploaded file:", err)
		return nil, fmt.Errorf("文件打开失败")
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		logger.Log.Error("Failed to create destination file:", err)
		return nil, fmt.Errorf("文件创建失败")
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := dst.ReadFrom(src); err != nil {
		logger.Log.Error("Failed to save file:", err)
		return nil, fmt.Errorf("文件保存失败")
	}

	relativePath := urlPrefix + "/" + filename
	logger.Log.Info("Image uploaded successfully:", relativePath)

	return &UploadResult{
		Path:     relativePath,
		Filename: filename,
		Size:     file.Size,
	}, nil
}

// SaveBackgroundImage 保存背景图片
func SaveBackgroundImage(file *multipart.FileHeader, savePath string) (*UploadResult, error) {
	return saveImageToDir(file, BackgroundImageDir, "/upload/background")
}

// DeleteBackgroundImage 删除背景图片
func DeleteBackgroundImage(path string) error {
	// 安全检查：确保路径在允许的目录内
	if !strings.HasPrefix(path, "/upload/background/") {
		return fmt.Errorf("无效的文件路径")
	}

	// 转换为实际文件路径
	filename := strings.TrimPrefix(path, "/upload/background/")
	filePath := filepath.Join(BackgroundImageDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在")
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		logger.Log.Error("Failed to delete file:", err)
		return fmt.Errorf("文件删除失败")
	}

	logger.Log.Info("Background image deleted successfully:", path)

	return nil
}

// SaveImage 保存通用图片
func SaveImage(file *multipart.FileHeader) (*UploadResult, error) {
	return saveImageToDir(file, UploadDir, "/upload")
}

// SaveLogoImage 保存系统 Logo
func SaveLogoImage(file *multipart.FileHeader) (*UploadResult, error) {
	return saveImageToDir(file, LogoUploadDir, "/upload/logo")
}
