package handler

import (
	"zbxtable/internal/model"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/sanbornm/go-selfupdate/selfupdate"
)

var updater = &selfupdate.Updater{
	ApiURL:         model.UpdateURL,
	BinURL:         model.UpdateURL,
	DiffURL:        model.UpdateURL,
	Dir:            "update/",
	CmdName:        "zbxtable",
	ForceCheck:     true,
	CurrentVersion: model.Version,
}

// UpdateInfo 更新信息响应结构
type UpdateInfo struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	HasUpdate      bool   `json:"hasUpdate"`
	UpdateURL      string `json:"updateUrl"`
	ReleaseNotes   string `json:"releaseNotes,omitempty"`
}

// UpdateResult 更新结果响应结构
type UpdateResult struct {
	Success      bool   `json:"success"`
	OldVersion   string `json:"oldVersion"`
	NewVersion   string `json:"newVersion"`
	NeedRestart  bool   `json:"needRestart"`
	RestartDelay int    `json:"restartDelay"`
}

// SystemVersion 系统版本信息响应结构
type SystemVersion struct {
	Version   string `json:"version"`
	GitHash   string `json:"gitHash"`
	BuildTime string `json:"buildTime"`
}

// CheckUpdate 检查更新
func CheckUpdate(c *gin.Context) {
	logger.Log.Info("检查更新...")

	// 更新当前版本信息
	updater.CurrentVersion = model.Version

	// 检查是否有更新
	updateVersion, err := updater.UpdateAvailable()
	if err != nil {
		logger.Log.Error("检查更新失败:", err)
		response.InternalError(c, "检查更新失败: "+err.Error())
		return
	}

	hasUpdate := updateVersion != ""
	latestVersion := updater.Info.Version
	if latestVersion == "" {
		latestVersion = model.Version
	}

	updateInfo := UpdateInfo{
		CurrentVersion: model.Version,
		LatestVersion:  latestVersion,
		HasUpdate:      hasUpdate,
		UpdateURL:      model.UpdateURL,
	}

	if hasUpdate {
		logger.Log.Infof("发现新版本: %s -> %s", model.Version, latestVersion)
	} else {
		logger.Log.Info("当前已是最新版本")
	}

	response.Success(c, updateInfo)
}

// DoUpdate 执行更新
func DoUpdate(c *gin.Context) {
	logger.Log.Info("开始执行更新...")

	// 更新当前版本信息
	updater.CurrentVersion = model.Version

	// 检查是否有更新
	updateVersion, err := updater.UpdateAvailable()
	if err != nil {
		logger.Log.Error("检查更新失败:", err)
		response.InternalError(c, "检查更新失败: "+err.Error())
		return
	}

	if updateVersion == "" {
		logger.Log.Info("当前已是最新版本，无需更新")
		response.BadRequest(c, "当前已是最新版本，无需更新")
		return
	}

	logger.Log.Infof("发现新版本: %s -> %s", model.Version, updater.Info.Version)

	// 执行更新
	logger.Log.Info("下载并安装新版本...")
	err = updater.Update()
	if err != nil {
		logger.Log.Error("更新失败:", err)
		response.InternalError(c, "更新失败: "+err.Error())
		return
	}

	logger.Log.Infof("更新成功: %s -> %s", model.Version, updater.Info.Version)

	// 返回成功响应，提示用户手动重启服务
	updateResult := UpdateResult{
		Success:      true,
		OldVersion:   model.Version,
		NewVersion:   updater.Info.Version,
		NeedRestart:  true,
		RestartDelay: 0, // 不自动重启
	}
	response.SuccessWithMessage(c, "更新成功，请手动重启服务以应用更新", updateResult)
}

// GetCurrentVersion 获取当前版本信息
func GetCurrentVersion(c *gin.Context) {
	versionInfo := SystemVersion{
		Version:   model.Version,
		GitHash:   model.GitHash,
		BuildTime: model.BuildTime,
	}
	response.Success(c, versionInfo)
}
