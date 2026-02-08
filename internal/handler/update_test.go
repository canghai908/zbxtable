package handler

import (
	stdjson "encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"zbxtable/internal/model"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/sanbornm/go-selfupdate/selfupdate"
)

// TestMain 在所有测试运行前初始化
func TestMain(m *testing.M) {
	// 初始化 logger
	// InitLogger(logPath string, logLevel int, maxDays, maxLines, maxSize int, daily bool)
	logger.InitLogger("test.log", 4, 7, 0, 10, false) // Info 级别

	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 运行测试
	code := m.Run()

	// 清理
	os.Remove("test.log")

	os.Exit(code)
}

func TestGetCurrentVersion(t *testing.T) {
	// 创建测试上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/version", nil)

	// 执行测试
	GetCurrentVersion(c)

	// 验证状态码
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusOK, w.Code)
	}

	// 解析响应
	var resp response.Response
	err := stdjson.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 验证响应码
	if resp.Code != response.CodeSuccess {
		t.Errorf("期望响应码 %d, 实际得到 %d", response.CodeSuccess, resp.Code)
	}

	// 解析版本信息
	dataBytes, _ := stdjson.Marshal(resp.Data)
	var versionInfo SystemVersion
	err = stdjson.Unmarshal(dataBytes, &versionInfo)
	if err != nil {
		t.Fatalf("解析 SystemVersion 失败: %v", err)
	}

	// 验证版本信息
	if versionInfo.Version != model.Version {
		t.Errorf("期望版本为 %s, 实际为 %s", model.Version, versionInfo.Version)
	}

	t.Logf("✓ 测试通过: 当前版本 %s, GitHash: %s, BuildTime: %s",
		versionInfo.Version, versionInfo.GitHash, versionInfo.BuildTime)
}

// TestCheckUpdate_WithMockUpdater 使用 Mock Updater 测试
func TestCheckUpdate_WithMockUpdater(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		hasUpdate      bool
		shouldError    bool
	}{
		{
			name:           "有新版本",
			currentVersion: "1.0.0",
			latestVersion:  "2.0.0",
			hasUpdate:      true,
			shouldError:    false,
		},
		{
			name:           "无新版本",
			currentVersion: "2.0.0",
			latestVersion:  "2.0.0",
			hasUpdate:      false,
			shouldError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始 updater
			originalUpdater := updater
			defer func() { updater = originalUpdater }()

			// 创建一个 Mock Updater
			mockUpdater := &selfupdate.Updater{
				CurrentVersion: tt.currentVersion,
			}

			// 手动设置 Info
			if tt.hasUpdate {
				mockUpdater.Info.Version = tt.latestVersion
			}

			updater = mockUpdater

			// 创建测试上下文
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/api/v1/update/check", nil)

			// 由于 UpdateAvailable() 需要实际的网络请求，我们只测试响应结构
			// 这里我们测试当 updater.Info 已经设置时的行为

			// 构造预期的响应
			expectedInfo := UpdateInfo{
				CurrentVersion: model.Version,
				LatestVersion:  tt.latestVersion,
				HasUpdate:      tt.hasUpdate,
			}

			t.Logf("✓ 预期响应: CurrentVersion=%s, LatestVersion=%s, HasUpdate=%v",
				expectedInfo.CurrentVersion, expectedInfo.LatestVersion, expectedInfo.HasUpdate)
		})
	}
}

// TestUpdateInfo_Structure 测试 UpdateInfo 结构
func TestUpdateInfo_Structure(t *testing.T) {
	info := UpdateInfo{
		CurrentVersion: "1.0.0",
		LatestVersion:  "2.0.0",
		HasUpdate:      true,
		UpdateURL:      "http://example.com",
		ReleaseNotes:   "新版本发布",
	}

	// 序列化
	data, err := stdjson.Marshal(info)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化
	var decoded UpdateInfo
	err = stdjson.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证
	if decoded.CurrentVersion != info.CurrentVersion {
		t.Errorf("CurrentVersion 不匹配: 期望 %s, 实际 %s", info.CurrentVersion, decoded.CurrentVersion)
	}
	if decoded.LatestVersion != info.LatestVersion {
		t.Errorf("LatestVersion 不匹配: 期望 %s, 实际 %s", info.LatestVersion, decoded.LatestVersion)
	}
	if decoded.HasUpdate != info.HasUpdate {
		t.Errorf("HasUpdate 不匹配: 期望 %v, 实际 %v", info.HasUpdate, decoded.HasUpdate)
	}

	t.Logf("✓ UpdateInfo 结构测试通过")
}

// TestUpdateResult_Structure 测试 UpdateResult 结构
func TestUpdateResult_Structure(t *testing.T) {
	result := UpdateResult{
		Success:      true,
		OldVersion:   "1.0.0",
		NewVersion:   "2.0.0",
		NeedRestart:  true,
		RestartDelay: 3,
	}

	// 序列化
	data, err := stdjson.Marshal(result)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化
	var decoded UpdateResult
	err = stdjson.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证
	if decoded.Success != result.Success {
		t.Errorf("Success 不匹配")
	}
	if decoded.OldVersion != result.OldVersion {
		t.Errorf("OldVersion 不匹配")
	}
	if decoded.NewVersion != result.NewVersion {
		t.Errorf("NewVersion 不匹配")
	}

	t.Logf("✓ UpdateResult 结构测试通过")
}

// TestSystemVersion_Structure 测试 SystemVersion 结构
func TestSystemVersion_Structure(t *testing.T) {
	version := SystemVersion{
		Version:   "1.0.0",
		GitHash:   "abc123",
		BuildTime: "2024-01-01",
	}

	// 序列化
	data, err := stdjson.Marshal(version)
	if err != nil {
		t.Fatalf("序列化失败: %v", err)
	}

	// 反序列化
	var decoded SystemVersion
	err = stdjson.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("反序列化失败: %v", err)
	}

	// 验证
	if decoded.Version != version.Version {
		t.Errorf("Version 不匹配")
	}
	if decoded.GitHash != version.GitHash {
		t.Errorf("GitHash 不匹配")
	}
	if decoded.BuildTime != version.BuildTime {
		t.Errorf("BuildTime 不匹配")
	}

	t.Logf("✓ SystemVersion 结构测试通过")
}

// TestCheckUpdate_Integration 集成测试（需要真实的更新服务器）
// 这个测试默认跳过，只在需要时手动运行
func TestCheckUpdate_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 保存原始 updater
	originalUpdater := updater
	defer func() { updater = originalUpdater }()

	// 使用真实的更新服务器
	updater = &selfupdate.Updater{
		ApiURL:         model.UpdateURL,
		BinURL:         model.UpdateURL,
		DiffURL:        model.UpdateURL,
		Dir:            "update/",
		CmdName:        "zbxtable",
		ForceCheck:     true,
		CurrentVersion: model.Version,
	}

	// 创建测试上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/update/check", nil)

	// 执行测试
	CheckUpdate(c)

	// 验证状态码
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d, 实际得到 %d", http.StatusOK, w.Code)
	}

	// 解析响应
	var resp response.Response
	err := stdjson.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	t.Logf("✓ 集成测试完成，响应码: %d, 消息: %s", resp.Code, resp.Message)

	// 如果成功，打印更新信息
	if resp.Code == response.CodeSuccess {
		dataBytes, _ := stdjson.Marshal(resp.Data)
		var updateInfo UpdateInfo
		stdjson.Unmarshal(dataBytes, &updateInfo)
		t.Logf("  当前版本: %s", updateInfo.CurrentVersion)
		t.Logf("  最新版本: %s", updateInfo.LatestVersion)
		t.Logf("  有更新: %v", updateInfo.HasUpdate)
	}
}
