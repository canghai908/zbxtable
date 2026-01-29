package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	models "zbxtable/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// ZabbixTenantBindingSafeResponse 安全的租户绑定响应结构，隐藏敏感信息
type ZabbixTenantBindingSafeResponse struct {
	ID               int    `json:"id"`
	TenantID         string `json:"tenant_id"`
	ZabbixInstanceID int    `json:"zabbix_instance_id"`
	Enabled          bool   `json:"enabled"`
	NotifyMethod     string `json:"notify_method"`
	MSAgentInstalled bool   `json:"ms_agent_installed"`
	MSAgentVersion   string `json:"ms_agent_version"`
	WebhookInstalled bool   `json:"webhook_installed"`
	WebhookURL       string `json:"webhook_url"`
	// 不包含 Token 字段
}

// toTenantSafeResponse 将租户绑定转换为安全响应
func toTenantSafeResponse(binding *models.ZabbixTenantBinding) ZabbixTenantBindingSafeResponse {
	if binding == nil {
		return ZabbixTenantBindingSafeResponse{}
	}
	return ZabbixTenantBindingSafeResponse{
		ID:               binding.ID,
		TenantID:         binding.TenantID,
		ZabbixInstanceID: binding.ZabbixInstanceID,
		Enabled:          binding.Enabled,
		NotifyMethod:     binding.NotifyMethod,
		MSAgentInstalled: binding.MSAgentInstalled,
		MSAgentVersion:   binding.MSAgentVersion,
		WebhookInstalled: binding.WebhookInstalled,
		WebhookURL:       binding.WebhookURL,
	}
}

// ListZabbixTenantBindingsGin 列出租户绑定
func ListZabbixTenantBindingsGin(c *gin.Context) {
	list, err := models.ListZabbixTenantBindings()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 转换为安全的响应结构，隐藏 token
	safeList := make([]ZabbixTenantBindingSafeResponse, 0, len(list))
	for _, binding := range list {
		safeList = append(safeList, toTenantSafeResponse(&binding))
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok", "data": safeList})
}

// UpsertZabbixTenantBindingGin 新增/更新租户绑定（按 tenant_id upsert）
func UpsertZabbixTenantBindingGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	tenantID := gjson.Get(string(body), "tenant_id").String()
	token := gjson.Get(string(body), "token").String()
	zid := int(gjson.Get(string(body), "zabbix_instance_id").Int())
	enabled := gjson.Get(string(body), "enabled").Bool()
	notifyMethod := gjson.Get(string(body), "notify_method").String()

	// 如果未指定通知方式，默认使用 msagent
	if notifyMethod == "" {
		notifyMethod = "msagent"
	}

	m := &models.ZabbixTenantBinding{
		TenantID:         tenantID,
		Token:            token,
		ZabbixInstanceID: zid,
		Enabled:          enabled,
		NotifyMethod:     notifyMethod,
	}
	if err := models.UpsertZabbixTenantBinding(m); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok"})
}

// DeleteZabbixTenantBindingGin 删除租户绑定
func DeleteZabbixTenantBindingGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	if err := models.DeleteZabbixTenantBinding(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok"})
}

// InstallMSAgentGin 在 Zabbix 中安装 MS-Agent 配置
func InstallMSAgentGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 获取租户绑定信息
	var binding models.ZabbixTenantBinding
	if err := models.DB.First(&binding, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "租户绑定不存在"})
		return
	}

	// 在 Zabbix 中安装 MS-Agent 配置
	if err := models.InstallMSAgentToZabbix(binding.ZabbixInstanceID, binding.TenantID); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("安装失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "MS-Agent 配置安装成功"})
}

// InstallWebhookGin 在 Zabbix 中安装 Webhook 配置
func InstallWebhookGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 获取租户绑定信息
	var binding models.ZabbixTenantBinding
	if err := models.DB.First(&binding, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "租户绑定不存在"})
		return
	}

	// 获取 ZbxTable 服务地址
	// 优先使用配置的 webhook_url
	zbxtableURL := models.GetConfigValueByKey("webhook_url", "")
	if zbxtableURL == "" {
		// 如果配置为空，尝试从环境变量获取
		zbxtableURL = os.Getenv("ZBXTABLE_URL")
	}
	if zbxtableURL == "" {
		// 最后从请求中获取
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		zbxtableURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}

	// 在 Zabbix 中安装 Webhook 配置
	if err := models.InstallWebhookToZabbix(binding.ZabbixInstanceID, binding.TenantID, zbxtableURL); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("安装失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Webhook 配置安装成功"})
}

// GenerateMSAgentInstallScriptGin 生成 MS-Agent 安装脚本
func GenerateMSAgentInstallScriptGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 获取租户绑定信息
	var binding models.ZabbixTenantBinding
	if err := models.DB.First(&binding, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "租户绑定不存在"})
		return
	}

	// 检查是否已安装
	if !binding.MSAgentInstalled {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请先在 Zabbix 中安装 MS-Agent 配置"})
		return
	}

	// 获取 ZbxTable 服务地址
	// 优先使用配置的 webhook_url
	zbxtableURL := models.GetConfigValueByKey("webhook_url", "")
	if zbxtableURL == "" {
		// 如果配置为空，尝试从环境变量获取
		zbxtableURL = os.Getenv("ZBXTABLE_URL")
	}
	if zbxtableURL == "" {
		// 最后从请求中获取
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		zbxtableURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}

	// 生成安装脚本
	config := &models.MSAgentConfig{
		ZbxTableURL: zbxtableURL,
		TenantID:    binding.TenantID,
		Token:       binding.Token,
	}

	script, err := models.GenerateMSAgentInstallScript(config)
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
	id, _ := strconv.Atoi(idStr)

	// 获取租户绑定信息
	var binding models.ZabbixTenantBinding
	if err := models.DB.First(&binding, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "租户绑定不存在"})
		return
	}

	// 检查是否已安装
	if !binding.WebhookInstalled {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请先在 Zabbix 中安装 Webhook 配置"})
		return
	}

	// 获取 ZbxTable 服务地址
	// 优先使用配置的 webhook_url
	zbxtableURL := models.GetConfigValueByKey("webhook_url", "")
	if zbxtableURL == "" {
		// 如果配置为空，尝试从环境变量获取
		zbxtableURL = os.Getenv("ZBXTABLE_URL")
	}
	if zbxtableURL == "" {
		// 最后从请求中获取
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		zbxtableURL = fmt.Sprintf("%s://%s", scheme, c.Request.Host)
	}

	// 获取 Webhook 信息
	info, err := models.GetWebhookInfo(binding.TenantID, zbxtableURL)
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
	id, _ := strconv.Atoi(idStr)

	// 获取租户绑定信息
	var binding models.ZabbixTenantBinding
	if err := models.DB.First(&binding, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "租户绑定不存在"})
		return
	}

	// 从 Zabbix 中卸载 MS-Agent 配置
	if err := models.UninstallMSAgentFromZabbix(binding.ZabbixInstanceID, binding.TenantID); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("卸载失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "MS-Agent 配置卸载成功"})
}

// UninstallWebhookGin 卸载 Webhook 配置
func UninstallWebhookGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 获取租户绑定信息
	var binding models.ZabbixTenantBinding
	if err := models.DB.First(&binding, id).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "租户绑定不存在"})
		return
	}

	// 从 Zabbix 中卸载 Webhook 配置
	if err := models.UninstallWebhookFromZabbix(binding.ZabbixInstanceID, binding.TenantID); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("卸载失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Webhook 配置卸载成功"})
}
