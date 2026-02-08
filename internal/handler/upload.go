package handler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// 上传目录
	UploadDir = "./upload"
	// 背景图片上传目录
	BackgroundImageDir = "./upload/background"
	// 允许的图片格式
	AllowedImageExts = ".jpg,.jpeg,.png,.gif,.bmp,.webp,.svg"
	// 最大文件大小 (10MB)
	MaxFileSize = 10 * 1024 * 1024
)

// init 初始化上传目录
func init() {
	// 确保上传目录存在
	dirs := []string{UploadDir, BackgroundImageDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Log.Error("Failed to create upload directory:", dir, err)
		}
	}
}

// UploadBackgroundImage 上传拓扑背景图片
func UploadBackgroundImage(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	// 检查文件大小
	if file.Size > MaxFileSize {
		response.BadRequest(c, fmt.Sprintf("文件大小超过限制，最大允许 %dMB", MaxFileSize/(1024*1024)))
		return
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !strings.Contains(AllowedImageExts, ext) {
		response.BadRequest(c, "不支持的文件格式，仅支持: "+AllowedImageExts)
		return
	}

	// 生成唯一文件名
	filename := generateUniqueFilename(ext)
	savePath := filepath.Join(BackgroundImageDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		logger.Log.Error("Failed to save uploaded file:", err)
		response.InternalError(c, "文件保存失败")
		return
	}

	// 返回文件访问路径（相对路径）
	relativePath := "/upload/background/" + filename

	logger.Log.Info("Background image uploaded successfully:", relativePath)

	response.SuccessWithMessage(c, "上传成功", gin.H{
		"path":     relativePath,
		"filename": filename,
		"size":     file.Size,
	})
}

// DeleteBackgroundImage 删除拓扑背景图片
func DeleteBackgroundImage(c *gin.Context) {
	// 从请求中获取文件路径
	path := c.Query("path")
	if path == "" {
		response.BadRequest(c, "文件路径不能为空")
		return
	}

	// 安全检查：确保路径在允许的目录内
	if !strings.HasPrefix(path, "/upload/background/") {
		response.BadRequest(c, "无效的文件路径")
		return
	}

	// 转换为实际文件路径
	filename := strings.TrimPrefix(path, "/upload/background/")
	filePath := filepath.Join(BackgroundImageDir, filename)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		response.BadRequest(c, "文件不存在")
		return
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		logger.Log.Error("Failed to delete file:", err)
		response.InternalError(c, "文件删除失败")
		return
	}

	logger.Log.Info("Background image deleted successfully:", path)

	response.SuccessWithMessage(c, "删除成功", nil)
}

// generateUniqueFilename 生成唯一的文件名
func generateUniqueFilename(ext string) string {
	// 使用时间戳 + UUID 生成唯一文件名
	timestamp := time.Now().Format("20060102150405")
	uniqueID := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%s%s", timestamp, uniqueID, ext)
}

// UploadImage 通用图片上传接口（可用于其他场景）
func UploadImage(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	// 检查文件大小
	if file.Size > MaxFileSize {
		response.BadRequest(c, fmt.Sprintf("文件大小超过限制，最大允许 %dMB", MaxFileSize/(1024*1024)))
		return
	}

	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !strings.Contains(AllowedImageExts, ext) {
		response.BadRequest(c, "不支持的文件格式，仅支持: "+AllowedImageExts)
		return
	}

	// 生成唯一文件名
	filename := generateUniqueFilename(ext)
	savePath := filepath.Join(UploadDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		logger.Log.Error("Failed to save uploaded file:", err)
		response.InternalError(c, "文件保存失败")
		return
	}

	// 返回文件访问路径（相对路径）
	relativePath := "/upload/" + filename

	logger.Log.Info("Image uploaded successfully:", relativePath)

	response.SuccessWithMessage(c, "上传成功", gin.H{
		"path":     relativePath,
		"filename": filename,
		"size":     file.Size,
	})
}
