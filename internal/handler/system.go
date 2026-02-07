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
func UpdateSystem(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 获取实例ID
	zidStr := gjson.Get(string(body), "zid").String()
	if zidStr == "" {
		response.BadRequest(c, "请选择 Zabbix 实例")
		return
	}
	zid, _ := strconv.ParseInt(zidStr, 10, 64)
	cpuCore := gjson.Get(string(body), "cpu_core").String()
	uptimeId := gjson.Get(string(body), "uptime_id").String()
	cpuUtilizationId := gjson.Get(string(body), "cpu_utilization_id").String()
	groupId := gjson.Get(string(body), "group_id").String()
	memoryTotalId := gjson.Get(string(body), "memory_total_id").String()
	memoryUsedId := gjson.Get(string(body), "memory_used_id").String()
	memoryUtilizationId := gjson.Get(string(body), "memory_utilization_id").String()
	mode := gjson.Get(string(body), "model").String()
	pingTemplateId := gjson.Get(string(body), "ping_template_id").String()

	var SystemRes model.SystemList
	v := model.System{
		ID:                  int64(id),
		ZID:                 int(zid),
		CPUCore:             cpuCore,
		CPUUtilizationID:    cpuUtilizationId,
		GroupID:             groupId,
		MemoryTotalID:       memoryTotalId,
		UptimeID:            uptimeId,
		Model:               mode,
		MemoryUsedID:        memoryUsedId,
		MemoryUtilizationID: memoryUtilizationId,
		PingTemplateID:      pingTemplateId,
	}
	err = model.CreateOrUpdateSystem(&v)
	if err != nil {
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
		// 定义敏感字段列表
		sensitiveKeys := map[string]bool{
			"email_secret":     true, // SMTP 密码/授权码
			"wechat_secret":    true, // 企业微信 Secret
			"deepseek_api_key": true, // Deepseek API Key
			"encryption_key":   true, // 加密密钥（完全隐藏）
		}
		
		// 过滤和脱敏处理
		filteredConfigs := []model.Config{}
		for _, config := range val {
			// 完全隐藏加密密钥配置项
			if config.Key == "encryption_key" {
				continue
			}
			
			// 对敏感字段进行脱敏处理
			if sensitiveKeys[config.Key] && config.Value != "" {
				config.Value = "********" // 替换为星号
			}
			
			filteredConfigs = append(filteredConfigs, config)
		}
		
		SystemRes.Code = 200
		SystemRes.Message = "获取成功"
		SystemRes.Data.Items = filteredConfigs
		SystemRes.Data.Total = int64(len(filteredConfigs))
	}
	c.JSON(http.StatusOK, SystemRes)
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

	value := gjson.Get(string(body), "value").String()
	
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
	v := model.Config{ID: int64(id), Value: value}
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
	// 获取系统名称和Logo配置
	systemName := model.GetConfigValueByKey("system_name", "ZbxTable")
	systemLogo := model.GetConfigValueByKey("system_logo", "/static/img/logo.png")

	response.Success(c, gin.H{
		"system_name": systemName,
		"system_logo": systemLogo,
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
		if config.Key == "initial_setup_completed" {
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
	v := model.Config{ID: configID, Value: "1"}
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
