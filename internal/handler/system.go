package handler

import (
	"encoding/base64"
	"io"
	"net/http"
	"strconv"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetAllSystem 获取系统配置列表
func GetAllSystem(c *gin.Context) {
	var SystemRes model.SystemList
	cnt, val, err := model.GetALlSystem()
	if err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
	} else {
		SystemRes.Code = 200
		SystemRes.Message = "ok"
		SystemRes.Data.Items = val
		SystemRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, SystemRes)
}

// GetSystemByID 获取单个系统配置
func GetSystemByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var SystemRes model.SystemList
	val, err := model.GetSystemByID(int64(id))
	if err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
	} else {
		SystemRes.Code = 200
		SystemRes.Message = "ok"
		SystemRes.Data.Items = val
		SystemRes.Data.Total = 1
	}
	c.JSON(http.StatusOK, SystemRes)
}

// UpdateSystem 更新系统配置（支持多实例）
func parseSystemBody(body string) model.System {
	zid, _ := strconv.Atoi(gjson.Get(body, "zid").String())
	return model.System{
		ZID:                 zid,
		Name:                gjson.Get(body, "name").String(),
		TypeCode:            gjson.Get(body, "type_code").String(),
		GroupID:             gjson.Get(body, "group_id").String(),
		CPUCore:             gjson.Get(body, "cpu_core").String(),
		CPUUtilizationID:    gjson.Get(body, "cpu_utilization_id").String(),
		MemoryTotalID:       gjson.Get(body, "memory_total_id").String(),
		MemoryUsedID:        gjson.Get(body, "memory_used_id").String(),
		MemoryUtilizationID: gjson.Get(body, "memory_utilization_id").String(),
		UptimeID:            gjson.Get(body, "uptime_id").String(),
		Model:               gjson.Get(body, "model").String(),
		PingTemplateID:      gjson.Get(body, "ping_template_id").String(),
		AutoInit:            int(gjson.Get(body, "auto_init").Int()),
		InitCron:            gjson.Get(body, "init_cron").String(),
		InitOnNewHost:       int(gjson.Get(body, "init_on_new_host").Int()),
		MaxRetry:            int(gjson.Get(body, "max_retry").Int()),
	}
}

func UpdateSystem(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	if gjson.Get(string(body), "zid").String() == "" {
		response.BadRequest(c, "请选择 Zabbix 实例")
		return
	}

	v := parseSystemBody(string(body))
	v.ID = int64(id)

	var SystemRes model.SystemList
	if err = model.CreateOrUpdateSystem(&v); err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
	} else {
		SystemRes.Code = 200
		SystemRes.Message = "更新成功"
	}
	c.JSON(http.StatusOK, SystemRes)
}

// SystemInit 初始化系统（支持多实例）
func SystemInit(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 读取请求体获取实例ID
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	zidStr := gjson.Get(string(body), "zid").String()
	if zidStr == "" {
		response.BadRequest(c, "请选择 Zabbix 实例")
		return
	}
	zid, _ := strconv.Atoi(zidStr)

	var SystemRes model.SystemList
	err = model.SystemInitWithInstance(int64(id), int(zid))
	if err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
		SystemRes.Data.Total = 0
		SystemRes.Data.Items = ""
	} else {
		SystemRes.Code = 200
		SystemRes.Message = "初始化完成"
		SystemRes.Data.Total = 0
		SystemRes.Data.Items = ""
	}
	c.JSON(http.StatusOK, SystemRes)
}

// CreateSystem 创建新的资产绑定配置
func CreateSystem(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}
	if gjson.Get(string(body), "zid").String() == "" {
		response.BadRequest(c, "请选择 Zabbix 实例")
		return
	}
	if gjson.Get(string(body), "type_code").String() == "" {
		response.BadRequest(c, "请选择资产类型")
		return
	}
	v := parseSystemBody(string(body))
	if err := model.CreateSystem(&v); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", v)
}

// GetSystemHistory 获取资产绑定初始化历史
func GetSystemHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	list, total, err := model.GetSystemHistory(id, page, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"items": list, "total": total})
}

// DeleteSystem 删除资产绑定配置
func DeleteSystem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := model.DeleteSystem(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetEgress 获取出口带宽配置
func GetEgress(c *gin.Context) {
	var SystemRes model.SystemList
	val, err := model.GetEgress()
	if err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
	} else {
		SystemRes.Code = 200
		SystemRes.Message = "获取成功"
		SystemRes.Data.Items = val
		SystemRes.Data.Total = 1
	}
	c.JSON(http.StatusOK, SystemRes)
}

// GetAllConfig 获取系统参数配置
func GetAllConfig(c *gin.Context) {
	var SystemRes model.SystemList
	val, err := model.GetConfigList()
	if err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
	} else {
		filteredConfigs := sanitizeConfigsForResponse(val)

		SystemRes.Code = 200
		SystemRes.Message = "获取成功"
		SystemRes.Data.Items = filteredConfigs
		SystemRes.Data.Total = int64(len(filteredConfigs))
	}
	c.JSON(http.StatusOK, SystemRes)
}

func sanitizeConfigsForResponse(configs []model.Config) []model.Config {
	filteredConfigs := make([]model.Config, 0, len(configs))
	for _, config := range configs {
		if config.ConfigKey == "encryption_key" {
			continue
		}

		if modelSensitiveConfigKey(config.ConfigKey) && config.ConfigValue != "" {
			config.ConfigValue = "********"
		}

		filteredConfigs = append(filteredConfigs, config)
	}
	return filteredConfigs
}

func modelSensitiveConfigKey(key string) bool {
	switch key {
	case "email_secret", "wechat_secret", "deepseek_api_key", "custom_api_key":
		return true
	default:
		return false
	}
}

// UpdateConfig 更新系统参数配置
func UpdateConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	value := gjson.Get(string(body), "config_value").String()

	// 如果值是星号（脱敏标记），则不更新
	// 这表示前端没有修改该敏感字段
	if value == "********" {
		var SystemRes model.SystemList
		SystemRes.Code = 200
		SystemRes.Message = "未修改敏感字段，保持原值"
		SystemRes.Data.Items = ""
		SystemRes.Data.Total = 1
		c.JSON(http.StatusOK, SystemRes)
		return
	}

	var SystemRes model.SystemList
	v := model.Config{ID: int64(id), ConfigValue: value}
	err = model.UpdateConfig(&v)
	if err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
	} else {
		SystemRes.Code = 200
		SystemRes.Message = "更新成功"
		SystemRes.Data.Items = ""
		SystemRes.Data.Total = 1
	}
	c.JSON(http.StatusOK, SystemRes)
}

// ============ 新的出口配置 API ============

// GetAllEgressConfigs 获取所有出口配置
func GetAllEgressConfigs(c *gin.Context) {
	configs, err := model.GetAllEgress()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, configs)
}

// GetEgressConfigByID 根据ID获取出口配置
func GetEgressConfigByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	config, err := model.GetEgressConfigByID(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, config)
}

// AddEgressConfig 添加出口配置
func AddEgressConfig(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	name := gjson.Get(string(body), "name").String()
	zidStr := gjson.Get(string(body), "zid").String()
	zid, _ := strconv.Atoi(zidStr)
	hostID := gjson.Get(string(body), "host_id").String()
	inItemID := gjson.Get(string(body), "in_item_id").String()
	outItemID := gjson.Get(string(body), "out_item_id").String()
	sortOrder := int(gjson.Get(string(body), "sort_order").Int())

	if name == "" || zidStr == "" || hostID == "" || inItemID == "" || outItemID == "" {
		response.BadRequest(c, "缺少必填字段")
		return
	}

	config := &model.Egress{
		Name:      name,
		ZID:       zid,
		HostID:    hostID,
		InItemID:  inItemID,
		OutItemID: outItemID,
		Status:    1,
		SortOrder: sortOrder,
	}

	err = model.AddEgress(config)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "添加成功", config)
}

// UpdateEgressConfigHandler 更新出口配置
func UpdateEgressConfigHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	name := gjson.Get(string(body), "name").String()
	zidStr := gjson.Get(string(body), "zid").String()
	hostID := gjson.Get(string(body), "host_id").String()
	inItemID := gjson.Get(string(body), "in_item_id").String()
	outItemID := gjson.Get(string(body), "out_item_id").String()
	status := int(gjson.Get(string(body), "status").Int())
	sortOrder := int(gjson.Get(string(body), "sort_order").Int())

	if name == "" || zidStr == "" || hostID == "" || inItemID == "" || outItemID == "" {
		response.BadRequest(c, "缺少必填字段")
		return
	}
	zid, _ := strconv.Atoi(zidStr)

	config := &model.Egress{
		ID:        id,
		Name:      name,
		ZID:       zid,
		HostID:    hostID,
		InItemID:  inItemID,
		OutItemID: outItemID,
		Status:    status,
		SortOrder: sortOrder,
	}

	err = model.UpdateEgress(config)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", config)
}

// DeleteEgressConfigHandler 删除出口配置
func DeleteEgressConfigHandler(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	err = model.DeleteEgressConfig(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// UploadLogo 上传系统Logo（存储为base64到数据库）
func UploadLogo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "文件上传失败: "+err.Error())
		return
	}

	// 检查文件类型
	contentType := file.Header.Get("Content-Type")
	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/jpg" && contentType != "image/svg+xml" {
		response.BadRequest(c, "只支持 PNG、JPG、JPEG 或 SVG 格式的图片")
		return
	}

	// 检查文件大小（限制为2MB）
	if file.Size > 2*1024*1024 {
		response.BadRequest(c, "文件大小不能超过2MB")
		return
	}

	// 打开文件并读取内容
	fileContent, err := file.Open()
	if err != nil {
		response.InternalError(c, "读取文件失败: "+err.Error())
		return
	}
	defer fileContent.Close()

	// 读取文件字节
	fileBytes, err := io.ReadAll(fileContent)
	if err != nil {
		response.InternalError(c, "读取文件内容失败: "+err.Error())
		return
	}

	// 转换为base64
	base64String := base64.StdEncoding.EncodeToString(fileBytes)

	// 构建完整的Data URL格式（包含MIME类型）
	var dataURL string
	switch contentType {
	case "image/svg+xml":
		dataURL = "data:image/svg+xml;base64," + base64String
	case "image/png":
		dataURL = "data:image/png;base64," + base64String
	case "image/jpeg", "image/jpg":
		dataURL = "data:image/jpeg;base64," + base64String
	default:
		dataURL = "data:image/png;base64," + base64String
	}

	// 返回base64编码的图片数据
	response.SuccessWithMessage(c, "上传成功", gin.H{
		"url": dataURL,
	})
}

// GetPublicSystemInfo 获取系统公开信息（无需认证）
func GetPublicSystemInfo(c *gin.Context) {
	// 默认值
	systemName := "ZbxTable"
	systemLogo := "/logo.png"
	demoMode := "false"

	// 尝试从数据库获取配置（如果数据库已初始化）
	if model.DB != nil {
		systemName = model.GetConfigValueByKey("system_name", "ZbxTable")
		systemLogo = model.GetConfigValueByKey("system_logo", "/logo.png")
		// 增加 demo_mode 返回
		demoMode = model.GetConfKey("demo_mode")
		if demoMode == "" {
			demoMode = model.GetConfigValueByKey("demo_mode", "false")
		}
	}

	response.Success(c, gin.H{
		"system_name": systemName,
		"system_logo": systemLogo,
		"demo_mode":   demoMode == "true",
	})
}

// GetInitialSetupStatus 获取初始配置状态
func GetInitialSetupStatus(c *gin.Context) {
	// 检查是否完成初始配置
	setupCompleted := model.GetConfigValueByKey("initial_setup_completed", "0")

	// 检查webhook_url是否已配置
	webhookURL := model.GetConfigValueByKey("webhook_url", "")
	webhookConfigured := webhookURL != ""

	// 检查是否有Zabbix实例
	instances, err := model.GetAllZabbixInstances()
	zabbixConfigured := err == nil && len(instances) > 0

	response.Success(c, gin.H{
		"setup_completed":    setupCompleted == "1",
		"webhook_configured": webhookConfigured,
		"zabbix_configured":  zabbixConfigured,
	})
}

// CompleteInitialSetup 标记初始配置已完成
func CompleteInitialSetup(c *gin.Context) {
	// 查找initial_setup_completed配置项
	configs, err := model.GetConfigList()
	if err != nil {
		response.InternalError(c, "获取配置失败")
		return
	}

	var configID int64
	found := false
	for _, config := range configs {
		if config.ConfigKey == "initial_setup_completed" {
			configID = config.ID
			found = true
			break
		}
	}

	if !found {
		response.InternalError(c, "配置项不存在")
		return
	}

	// 更新配置
	v := model.Config{ID: configID, ConfigValue: "1"}
	err = model.UpdateConfig(&v)
	if err != nil {
		response.InternalError(c, "更新配置失败")
		return
	}

	response.SuccessWithMessage(c, "初始配置已完成", nil)
}

// TestEmailConfig 测试邮件配置
func TestEmailConfig(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	testEmail := gjson.Get(string(body), "test_email").String()
	if testEmail == "" {
		response.BadRequest(c, "请输入测试邮箱地址")
		return
	}

	// 发送测试邮件
	err = model.SendTestEmail(testEmail)
	if err != nil {
		response.InternalError(c, "邮件发送失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "测试邮件发送成功，请检查邮箱", nil)
}

// TestWechatConfig 测试企业微信配置
func TestWechatConfig(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	testUserID := gjson.Get(string(body), "test_user_id").String()
	if testUserID == "" {
		response.BadRequest(c, "请输入测试用户ID")
		return
	}

	// 发送测试企业微信消息
	err = model.SendTestWechat(testUserID)
	if err != nil {
		response.InternalError(c, "企业微信消息发送失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "测试消息发送成功，请检查企业微信", nil)
}
