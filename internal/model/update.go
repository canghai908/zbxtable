package model

import (
	"time"
	"zbxtable/pkg/logger"

	"github.com/sanbornm/go-selfupdate/selfupdate"
)

const UpdateURL = "http://dl.cactifans.com/stable/"

var (
	updateChecker *selfupdate.Updater
	updateTicker  *time.Ticker
	stopUpdateCheck chan bool
)

// InitUpdateChecker 初始化更新检查器
func InitUpdateChecker() {
	updateChecker = &selfupdate.Updater{
		ApiURL:         UpdateURL,
		BinURL:         UpdateURL,
		DiffURL:        UpdateURL,
		Dir:            "update/",
		CmdName:        "zbxtable",
		ForceCheck:     true,
		CurrentVersion: Version,
	}

	// 启动定期检查更新的任务（每24小时检查一次）
	stopUpdateCheck = make(chan bool)
	updateTicker = time.NewTicker(24 * time.Hour)

	go func() {
		// 启动后延迟5分钟进行首次检查
		time.Sleep(5 * time.Minute)
		checkForUpdates()

		for {
			select {
			case <-updateTicker.C:
				checkForUpdates()
			case <-stopUpdateCheck:
				logger.Log.Info("停止更新检查任务")
				return
			}
		}
	}()

	logger.Log.Info("更新检查器已启动，将每24小时检查一次更新")
}

// checkForUpdates 检查更新
func checkForUpdates() {
	logger.Log.Info("定期检查更新...")

	updateChecker.CurrentVersion = Version
	updateVersion, err := updateChecker.UpdateAvailable()
	if err != nil {
		logger.Log.Error("检查更新失败:", err)
		return
	}

	if updateVersion != "" {
		logger.Log.Infof("发现新版本: %s -> %s", Version, updateChecker.Info.Version)
		logger.Log.Info("请访问系统设置页面进行更新")
	} else {
		logger.Log.Info("当前已是最新版本")
	}
}

// StopUpdateChecker 停止更新检查器
func StopUpdateChecker() {
	if updateTicker != nil {
		updateTicker.Stop()
	}
	if stopUpdateCheck != nil {
		close(stopUpdateCheck)
	}
	logger.Log.Info("更新检查器已停止")
}

// GetUpdateChecker 获取更新检查器实例
func GetUpdateChecker() *selfupdate.Updater {
	if updateChecker == nil {
		InitUpdateChecker()
	}
	return updateChecker
}
