package handler

import (
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// UploadBackgroundImage 上传拓扑背景图片
func UploadBackgroundImage(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	// 调用 model 层保存文件
	result, err := model.SaveBackgroundImage(file, "")
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "上传成功", gin.H{
		"path":     result.Path,
		"filename": result.Filename,
		"size":     result.Size,
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

	// 调用 model 层删除文件
	if err := model.DeleteBackgroundImage(path); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// UploadImage 通用图片上传接口（可用于其他场景）
func UploadImage(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	// 调用 model 层保存文件
	result, err := model.SaveImage(file)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "上传成功", gin.H{
		"path":     result.Path,
		"filename": result.Filename,
		"size":     result.Size,
	})
}
