package handler

import (
	"os"
	"os/exec"
	"syscall"
	"time"
	"zbxtable/internal/model"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/sanbornm/go-selfupdate/selfupdate"
)

const UpdateURL = "http://dl.cactifans.com/stable/"

var updater = &selfupdate.Updater{
	ApiURL:         UpdateURL,
	BinURL:         UpdateURL,
	DiffURL:        UpdateURL,
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
		UpdateURL:      UpdateURL,
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

	// 备份旧的 web 目录
	logger.Log.Info("备份旧的 web 目录...")
	_ = os.RemoveAll("./.web")
	_ = os.Rename("./web", "./.web")

	// 执行更新
	logger.Log.Info("下载并安装新版本...")
	err = updater.Update()
	if err != nil {
		logger.Log.Error("更新失败:", err)
		// 恢复 web 目录
		_ = os.Rename("./.web", "./web")
		response.InternalError(c, "更新失败: "+err.Error())
		return
	}

	logger.Log.Infof("更新成功: %s -> %s", model.Version, updater.Info.Version)

	// 返回成功响应
	updateResult := UpdateResult{
		Success:      true,
		OldVersion:   model.Version,
		NewVersion:   updater.Info.Version,
		NeedRestart:  true,
		RestartDelay: 3,
	}
	response.SuccessWithMessage(c, "更新成功，系统将在 3 秒后自动重启", updateResult)

	// 延迟重启，让响应先返回给客户端
	go func() {
		time.Sleep(3 * time.Second)
		triggerUpdateRestart()
	}()
}

// triggerUpdateRestart 触发更新后重启
func triggerUpdateRestart() {
	logger.Log.Info("更新完成，触发系统重启...")

	// 获取当前进程的可执行文件路径
	executable, err := os.Executable()
	if err != nil {
		logger.Log.Error("获取可执行文件路径失败:", err)
		// 尝试使用 SIGTERM 信号退出，让外部服务管理器重启
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		return
	}

	// 获取当前进程的参数
	args := os.Args

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		logger.Log.Error("获取工作目录失败:", err)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		return
	}

	// 创建新进程
	cmd := exec.Command(executable, args[1:]...)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// 启动新进程
	err = cmd.Start()
	if err != nil {
		logger.Log.Error("启动新进程失败:", err)
		// 如果启动失败，尝试发送信号让服务管理器重启
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		return
	}

	logger.Log.Info("新进程已启动，PID:", cmd.Process.Pid)

	// 等待一小段时间确保新进程启动成功
	time.Sleep(1 * time.Second)

	// 退出当前进程
	logger.Log.Info("当前进程即将退出...")
	os.Exit(0)
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
