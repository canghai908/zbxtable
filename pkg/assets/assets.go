package assets

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"zbxtable/pkg/logger"
)

//go:embed all:files
var AssetsFS embed.FS

const (
	EchartsMinJS = "echarts.min.js"
	ShineJS      = "shine.js"
)

// RestoreAssets 释放静态资源文件到指定目录
// 在程序启动时调用，将嵌入的JS文件释放到 ./assets 目录
// 支持递归处理子目录
func RestoreAssets() error {
	targetDir := "./assets"

	// 检查目录是否存在
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		// 目录不存在，创建目录
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return err
		}
		logger.Log.Info("Created assets directory:", targetDir)
	}

	// 递归遍历嵌入的文件系统
	err := fs.WalkDir(AssetsFS, "files", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过 .DS_Store 文件
		if d.Name() == ".DS_Store" {
			return nil
		}

		// 计算相对路径（去掉 "files/" 前缀）
		relPath, err := filepath.Rel("files", path)
		if err != nil {
			return err
		}

		// 如果是根目录，跳过
		if relPath == "." {
			return nil
		}

		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			// 创建目录
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				logger.Log.Errorf("Failed to create directory %s: %v", targetPath, err)
				return err
			}
			logger.Log.Debugf("Created directory: %s", targetPath)
		} else {
			// 检查文件是否已存在
			if _, err := os.Stat(targetPath); err == nil {
				// 文件已存在，跳过
				logger.Log.Debugf("Asset file already exists, skipping: %s", targetPath)
				return nil
			}

			// 读取嵌入的文件内容
			content, err := AssetsFS.ReadFile(path)
			if err != nil {
				logger.Log.Errorf("Failed to read embedded asset %s: %v", path, err)
				return err
			}

			// 写入到目标文件
			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				logger.Log.Errorf("Failed to write asset file %s: %v", targetPath, err)
				return err
			}

			logger.Log.Infof("Restored asset file: %s", targetPath)
		}

		return nil
	})

	if err != nil {
		return err
	}

	logger.Log.Info("Asset files restored successfully")
	return nil
}

// CopyAssetsToDir 将静态资源文件从 ./assets 复制到指定目录
// 返回复制的文件路径列表
// 会在目标目录下创建 assets/js/ 和 assets/js/themes/ 子目录结构
func CopyAssetsToDir(targetDir string) ([]string, error) {
	// 确保目标目录存在
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, err
	}

	// 确保静态资源已释放
	if err := RestoreAssets(); err != nil {
		return nil, err
	}

	var copiedFiles []string

	// 创建 assets/js 目录
	jsDir := targetDir + "assets/js/"
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		return nil, err
	}

	// 创建 assets/js/themes 目录
	themesDir := targetDir + "assets/js/themes/"
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return nil, err
	}

	// 复制 echarts.min.js 到 assets/js/ 目录
	echartsSource := filepath.Join("./assets", "js", EchartsMinJS)
	echartsTarget := jsDir + EchartsMinJS
	if err := copyFile(echartsSource, echartsTarget); err != nil {
		logger.Log.Warnf("Failed to copy echarts.min.js: %v", err)
	} else {
		copiedFiles = append(copiedFiles, echartsTarget)
	}

	// 复制 shine.js 到 assets/js/themes/ 目录
	shineSource := filepath.Join("./assets", "js", "themes", ShineJS)
	shineTarget := themesDir + ShineJS
	if err := copyFile(shineSource, shineTarget); err != nil {
		logger.Log.Warnf("Failed to copy shine.js: %v", err)
	} else {
		copiedFiles = append(copiedFiles, shineTarget)
	}

	return copiedFiles, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}

// GetLocalAssetsHost 返回本地相对路径（用于HTML中的引用）
func GetLocalAssetsHost() string {
	return "./assets/js/"
}
