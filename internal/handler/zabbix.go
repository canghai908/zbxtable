package handler

import (
	"fmt"
	"io"
	"os"
	"strconv"
	model "zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// ZabbixInstanceafeResponse 安全的租户响应结构，隐藏敏感信息
type ZabbixInstanceafeResponse struct {
	ID               int      `json:"id"`
	Instance         string   `json:"instance"`
	Name             string   `json:"name"`
	URL              string   `json:"url"`
	User             string   `json:"user"`
	Enabled          bool     `json:"enabled"`
	Version          string   `json:"version"`
	LastTestOk       bool     `json:"last_test_ok"`
	LastTestMessage  string   `json:"last_test_message"`
	NotifyMethod     string   `json:"notify_method"`
	MSAgentInstalled bool     `json:"ms_agent_installed"`
	MSAgentVersion   string   `json:"ms_agent_version"`
	WebhookInstalled bool     `json:"webhook_installed"`
	WebhookURL       string   `json:"webhook_url"`
	AuthMethods      []string `json:"auth_methods"` // 认证方式：password, token
	// 不包含 Pass, Token, WebhookToken 字段
}

// toTenantSafeResponse 将租户转换为安全响应
func toTenantSafeResponse(instance *model.ZabbixInstance) ZabbixInstanceafeResponse {
	if instance == nil {
		return ZabbixInstanceafeResponse{}
	}

	// 判断认证方式
	authMethods := []string{}
	if instance.User != "" && instance.Pass != "" {
		authMethods = append(authMethods, "password")
	}
	if instance.Token != "" {
		authMethods = append(authMethods, "token")
	}

	return ZabbixInstanceafeResponse{
		ID:               instance.ID,
		Instance:         instance.Instance,
		Name:             instance.Name,
		URL:              instance.URL,
		Enabled:          instance.Enabled,
		Version:          instance.Version,
		LastTestOk:       instance.LastTestOk,
		LastTestMessage:  instance.LastTestMessage,
		NotifyMethod:     instance.NotifyMethod,
		MSAgentInstalled: instance.MSAgentInstalled,
		MSAgentVersion:   instance.MSAgentVersion,
		WebhookInstalled: instance.WebhookInstalled,
		WebhookURL:       instance.WebhookURL,
		AuthMethods:      authMethods,
	}
}

// ListZabbixInstanceGin 列出所有租户
func ListZabbixInstanceGin(c *gin.Context) {
	list, err := model.ListZabbixInstance()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 转换为安全的响应结构
	safeList := make([]ZabbixInstanceafeResponse, 0, len(list))
	for _, tenant := range list {
		safeList = append(safeList, toTenantSafeResponse(&tenant))
	}

	response.Success(c, safeList)
}

// GetZabbixInstanceGin 获取单个租户（用于编辑）
func GetZabbixInstanceGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	tenant, err := model.GetZabbixInstanceByZID(int(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 返回租户信息，但不包含密码和Token（安全考虑，编辑时不显示已配置的密码和Token）
	response.Success(c, gin.H{
		"id":                 tenant.ID,
		"instance":           tenant.Instance,
		"name":               tenant.Name,
		"url":                tenant.URL,
		"user":               tenant.User,
		"pass":               "", // 不返回密码
		"token":              "", // 不返回Token
		"enabled":            tenant.Enabled,
		"version":            tenant.Version,
		"last_test_ok":       tenant.LastTestOk,
		"last_test_message":  tenant.LastTestMessage,
		"notify_method":      tenant.NotifyMethod,
		"ms_agent_installed": tenant.MSAgentInstalled,
		"ms_agent_version":   tenant.MSAgentVersion,
		"webhook_installed":  tenant.WebhookInstalled,
		"webhook_url":        tenant.WebhookURL,
	})
}

// CreateZabbixInstanceGin 创建租户
func CreateZabbixInstanceGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}
	m := &model.ZabbixInstance{
		Instance:     gjson.Get(string(body), "instance").String(),
		Name:         gjson.Get(string(body), "name").String(),
		URL:          gjson.Get(string(body), "url").String(),
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
	if err := model.CreateZabbixInstance(m); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 创建成功后，自动测试连接并更新版本信息
	// 查询刚创建的租户（获取ID）
	_, _, _ = model.TestAndUpdateZabbixInstance(m.ID)

	response.Success(c, nil)
}

// TestZabbixInstanceConfigGin 测试配置（创建前）
func TestZabbixInstanceConfigGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}
	webURL := gjson.Get(string(body), "url").String()
	user := gjson.Get(string(body), "user").String()
	pass := gjson.Get(string(body), "pass").String()
	token := gjson.Get(string(body), "token").String()

	ver, err := model.TestZabbixInstanceConfig(webURL, user, pass, token)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "连接成功", gin.H{"version": ver})
}

// TestZabbixInstanceGin 测试租户连接
func TestZabbixInstanceGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	tenant, ver, err := model.TestAndUpdateZabbixInstance(int(id))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "连接成功", gin.H{
		"version":           ver,
		"last_test_ok":      tenant.LastTestOk,
		"last_test_message": tenant.LastTestMessage,
	})
}

// UpdateZabbixInstanceGin 更新租户
func UpdateZabbixInstanceGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	patch := &model.ZabbixInstance{
		Instance:     gjson.Get(string(body), "tenant_id").String(),
		Name:         gjson.Get(string(body), "name").String(),
		URL:          gjson.Get(string(body), "url").String(),
		User:         gjson.Get(string(body), "user").String(),
		Pass:         gjson.Get(string(body), "pass").String(),
		Token:        gjson.Get(string(body), "token").String(),
		Enabled:      gjson.Get(string(body), "enabled").Bool(),
		NotifyMethod: gjson.Get(string(body), "notify_method").String(),
	}

	instance, err := model.UpdateZabbixInstance(int(id), patch)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 更新成功后，自动测试连接并更新版本信息
	// 注意：UpdateZabbixInstance 内部已经处理了连接变更时重置版本
	// 这里再次测试以获取最新版本
	instance, _, _ = model.TestAndUpdateZabbixInstance(int(id))

	response.Success(c, toTenantSafeResponse(instance))
}

// DeleteZabbixInstanceGin 删除租户
func DeleteZabbixInstanceGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	if err := model.DeleteZabbixInstance(int(id)); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// EnableZabbixInstanceGin 启用/禁用租户
func EnableZabbixInstanceGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	enabled := gjson.Get(string(body), "enabled").Bool()

	tenant, err := model.SetZabbixInstanceEnabled(int(id), enabled)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, toTenantSafeResponse(tenant))
}

// InstallWebhookGin 在 Zabbix 中安装 Webhook 配置
func InstallWebhookGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// 获取租户信息
	instance, err := model.GetZabbixInstanceByZID(int(id))
	if err != nil {
		response.InternalError(c, "租户不存在")
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
	if err := model.InstallWebhookToZabbix(instance.ID, zbxtableURL); err != nil {
		response.InternalError(c, fmt.Sprintf("安装失败: %v", err))
		return
	}

	response.SuccessWithMessage(c, "Webhook 配置安装成功", nil)
}

// GetWebhookInfoGin 获取 Webhook 配置信息
func GetWebhookInfoGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 获取租户信息
	instance, err := model.GetZabbixInstanceByZID(id)
	if err != nil {
		response.InternalError(c, "实例不存在")
		return
	}

	// 检查是否已安装
	if !instance.WebhookInstalled {
		response.BadRequest(c, "请先在 Zabbix 中安装 Webhook 配置")
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
	info, err := model.GetWebhookInfo(instance.ID, zbxtableURL)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("获取信息失败: %v", err))
		return
	}

	response.Success(c, info)
}

// UninstallMSAgentGin 卸载 MS-Agent 配置
func UninstallMSAgentGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	// 从 Zabbix 中卸载 MS-Agent 配置
	if err := model.UninstallMSAgentFromZabbixInstance(int(id)); err != nil {
		response.InternalError(c, fmt.Sprintf("卸载失败: %v", err))
		return
	}

	response.SuccessWithMessage(c, "MS-Agent 配置卸载成功", nil)
}

// UninstallWebhookGin 卸载 Webhook 配置
func UninstallWebhookGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	// 从 Zabbix 中卸载 Webhook 配置
	if err := model.UninstallWebhookFromZabbixInstance(int(id)); err != nil {
		response.InternalError(c, fmt.Sprintf("卸载失败: %v", err))
		return
	}

	response.SuccessWithMessage(c, "Webhook 配置卸载成功", nil)
}
