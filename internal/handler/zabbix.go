package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	model "zbxtable/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// ZabbixTenantSafeResponse 安全的租户响应结构，隐藏敏感信息
type ZabbixTenantSafeResponse struct {
	ID               int    `json:"id"`
	TenantID         string `json:"tenant_id"`
	Name             string `json:"name"`
	WebURL           string `json:"web_url"`
	User             string `json:"user"`
	Enabled          bool   `json:"enabled"`
	Version          string `json:"version"`
	LastTestOk       bool   `json:"last_test_ok"`
	LastTestMessage  string `json:"last_test_message"`
	NotifyMethod     string `json:"notify_method"`
	MSAgentInstalled bool   `json:"ms_agent_installed"`
	MSAgentVersion   string `json:"ms_agent_version"`
	WebhookInstalled bool   `json:"webhook_installed"`
	WebhookURL       string `json:"webhook_url"`
	// 不包含 Pass, Token, WebhookToken 字段
}

// toTenantSafeResponse 将租户转换为安全响应
func toTenantSafeResponse(tenant *model.ZabbixTenant) ZabbixTenantSafeResponse {
	if tenant == nil {
		return ZabbixTenantSafeResponse{}
	}
	return ZabbixTenantSafeResponse{
		ID:               tenant.ID,
		TenantID:         tenant.TenantID,
		Name:             tenant.Name,
		WebURL:           tenant.WebURL,
		User:             tenant.User,
		Enabled:          tenant.Enabled,
		Version:          tenant.Version,
		LastTestOk:       tenant.LastTestOk,
		LastTestMessage:  tenant.LastTestMessage,
		NotifyMethod:     tenant.NotifyMethod,
		MSAgentInstalled: tenant.MSAgentInstalled,
		MSAgentVersion:   tenant.MSAgentVersion,
		WebhookInstalled: tenant.WebhookInstalled,
		WebhookURL:       tenant.WebhookURL,
	}
}

// ListZabbixTenantsGin 列出所有租户
func ListZabbixTenantsGin(c *gin.Context) {
	list, err := model.ListZabbixTenants()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 转换为安全的响应结构
	safeList := make([]ZabbixTenantSafeResponse, 0, len(list))
	for _, tenant := range list {
		safeList = append(safeList, toTenantSafeResponse(&tenant))
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok", "data": safeList})
}

// GetZabbixTenantGin 获取单个租户
func GetZabbixTenantGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	tenant, err := model.GetZabbixTenantByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok", "data": toTenantSafeResponse(tenant)})
}

// CreateZabbixTenantGin 创建租户
func CreateZabbixTenantGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}

	m := &model.ZabbixTenant{
		TenantID:     gjson.Get(string(body), "tenant_id").String(),
		Name:         gjson.Get(string(body), "name").String(),
		WebURL:       gjson.Get(string(body), "web_url").String(),
		User:         gjson.Get(string(body), "user").String(),
		Pass:         gjson.Get(string(body), "pass").String(),
		Token:        gjson.Get(string(body), "token").String(),
		Enabled:      gjson.Get(string(body), "enabled").Bool(),
		NotifyMethod: gjson.Get(string(body), "notify_method").String(),
	}

	// 默认通知方式
	if m.NotifyMethod == "" {
		m.NotifyMethod = "webhook"
	}

	// 创建租户
	if err := model.CreateZabbixTenant(m); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 创建成功后，自动测试连接并更新版本信息
	// 查询刚创建的租户（获取ID）
	tenant, err := model.GetZabbixTenantByTenantID(m.TenantID)
	if err == nil && tenant != nil {
		// 测试连接并更新版本
		_, _, _ = model.TestAndUpdateZabbixTenant(int64(tenant.ID))
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok"})
}

// TestZabbixTenantConfigGin 测试配置（创建前）
func TestZabbixTenantConfigGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}

	webURL := gjson.Get(string(body), "web_url").String()
	user := gjson.Get(string(body), "user").String()
	pass := gjson.Get(string(body), "pass").String()
	token := gjson.Get(string(body), "token").String()

	ver, err := model.TestZabbixTenantConfig(webURL, user, pass, token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "连接成功", "data": gin.H{"version": ver}})
}

// TestZabbixTenantGin 测试租户连接
func TestZabbixTenantGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	tenant, ver, err := model.TestAndUpdateZabbixTenant(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "连接成功",
		"data": gin.H{
			"version":           ver,
			"last_test_ok":      tenant.LastTestOk,
			"last_test_message": tenant.LastTestMessage,
		},
	})
}

// UpdateZabbixTenantGin 更新租户
func UpdateZabbixTenantGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}

	patch := &model.ZabbixTenant{
		TenantID:     gjson.Get(string(body), "tenant_id").String(),
		Name:         gjson.Get(string(body), "name").String(),
		WebURL:       gjson.Get(string(body), "web_url").String(),
		User:         gjson.Get(string(body), "user").String(),
		Pass:         gjson.Get(string(body), "pass").String(),
		Token:        gjson.Get(string(body), "token").String(),
		Enabled:      gjson.Get(string(body), "enabled").Bool(),
		NotifyMethod: gjson.Get(string(body), "notify_method").String(),
	}

	tenant, err := model.UpdateZabbixTenant(id, patch)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 更新成功后，自动测试连接并更新版本信息
	// 注意：UpdateZabbixTenant 内部已经处理了连接变更时重置版本
	// 这里再次测试以获取最新版本
	tenant, _, _ = model.TestAndUpdateZabbixTenant(id)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok", "data": toTenantSafeResponse(tenant)})
}

// DeleteZabbixTenantGin 删除租户
func DeleteZabbixTenantGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	if err := model.DeleteZabbixTenant(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok"})
}

// EnableZabbixTenantGin 启用/禁用租户
func EnableZabbixTenantGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}

	enabled := gjson.Get(string(body), "enabled").Bool()

	tenant, err := model.SetZabbixTenantEnabled(id, enabled)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok", "data": toTenantSafeResponse(tenant)})
}

// InstallMSAgentGin 在 Zabbix 中安装 MS-Agent 配置
func InstallMSAgentGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// 获取租户信息
	tenant, err := model.GetZabbixTenantByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "实例不存在"})
		return
	}

	// 在 Zabbix 中安装 MS-Agent 配置
	if err := model.InstallMSAgentToZabbix(tenant.TenantID); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("安装失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "MS-Agent 配置安装成功"})
}

// InstallWebhookGin 在 Zabbix 中安装 Webhook 配置
func InstallWebhookGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// 获取租户信息
	tenant, err := model.GetZabbixTenantByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "租户不存在"})
		return
	}

	// 获取 ZbxTable 服务地址
	zbxtableURL := model.GetConfigValueByKey("webhook_url", "")
	if zbxtableURL == "" {
		zbxtableURL = os.Getenv("ZBXTABLE_URL")
	}
	if zbxtableURL == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		zbxtableURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}

	// 在 Zabbix 中安装 Webhook 配置
	if err := model.InstallWebhookToZabbix(tenant.TenantID, zbxtableURL); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("安装失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Webhook 配置安装成功"})
}

// GenerateMSAgentInstallScriptGin 生成 MS-Agent 安装脚本
func GenerateMSAgentInstallScriptGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// 获取租户信息
	tenant, err := model.GetZabbixTenantByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "租户不存在"})
		return
	}

	// 检查是否已安装
	if !tenant.MSAgentInstalled {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请先在 Zabbix 中安装 MS-Agent 配置"})
		return
	}

	// 获取 ZbxTable 服务地址
	zbxtableURL := model.GetConfigValueByKey("webhook_url", "")
	if zbxtableURL == "" {
		zbxtableURL = os.Getenv("ZBXTABLE_URL")
	}
	if zbxtableURL == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		zbxtableURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}

	// 生成安装脚本
	config := &model.MSAgentConfig{
		ZbxTableURL:  zbxtableURL,
		TenantID:     tenant.TenantID,
		WebhookToken: tenant.WebhookToken,
	}

	script, err := model.GenerateMSAgentInstallScript(config)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("生成脚本失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "ok",
		"data":    script,
	})
}

// GetWebhookInfoGin 获取 Webhook 配置信息
func GetWebhookInfoGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// 获取租户信息
	tenant, err := model.GetZabbixTenantByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "租户不存在"})
		return
	}

	// 检查是否已安装
	if !tenant.WebhookInstalled {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请先在 Zabbix 中安装 Webhook 配置"})
		return
	}

	// 获取 ZbxTable 服务地址
	zbxtableURL := model.GetConfigValueByKey("webhook_url", "")
	if zbxtableURL == "" {
		zbxtableURL = os.Getenv("ZBXTABLE_URL")
	}
	if zbxtableURL == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		zbxtableURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}

	// 获取 Webhook 信息
	info, err := model.GetWebhookInfo(tenant.TenantID, zbxtableURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("获取信息失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "ok",
		"data":    info,
	})
}

// UninstallMSAgentGin 卸载 MS-Agent 配置
func UninstallMSAgentGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// 从 Zabbix 中卸载 MS-Agent 配置
	if err := model.UninstallMSAgentFromZabbixTenant(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("卸载失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "MS-Agent 配置卸载成功"})
}

// UninstallWebhookGin 卸载 Webhook 配置
func UninstallWebhookGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// 从 Zabbix 中卸载 Webhook 配置
	if err := model.UninstallWebhookFromZabbixTenant(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("卸载失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Webhook 配置卸载成功"})
}
