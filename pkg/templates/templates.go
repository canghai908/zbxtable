package templates

import (
	"embed"
	"os"
	"path/filepath"
	"zbxtable/pkg/utils"
)

//go:embed all:files
var templatesFS embed.FS

// RestoreTemplates 释放模板文件到指定目录
// 在程序启动时调用，将嵌入的模板文件释放到 ./template 目录
func RestoreTemplates() error {
	targetDir := "./template"

	// 检查目录是否存在
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		// 目录不存在，创建目录
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return err
		}
		utils.Log.Info("Created template directory:", targetDir)
	}

	// 读取嵌入的文件列表
	entries, err := templatesFS.ReadDir("files")
	if err != nil {
		return err
	}

	// 遍历并释放每个文件
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		sourcePath := "files/" + fileName
		targetPath := filepath.Join(targetDir, fileName)

		// 检查文件是否已存在
		if _, err := os.Stat(targetPath); err == nil {
			// 文件已存在，跳过
			utils.Log.Infof("Template file already exists, skipping: %s", targetPath)
			continue
		}

		// 读取嵌入的文件内容
		content, err := templatesFS.ReadFile(sourcePath)
		if err != nil {
			utils.Log.Errorf("Failed to read embedded template %s: %v", fileName, err)
			continue
		}

		// 写入到目标文件
		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			utils.Log.Errorf("Failed to write template file %s: %v", targetPath, err)
			continue
		}

		utils.Log.Infof("Restored template file: %s", targetPath)
	}

	utils.Log.Info("Template files restored successfully")
	return nil
}

// GetTemplateContent 获取模板内容（从嵌入的文件系统）
// 如果需要直接从内存读取模板，可以使用此函数
func GetTemplateContent(filename string) ([]byte, error) {
	return templatesFS.ReadFile("files/" + filename)
}
