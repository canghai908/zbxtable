package handler

import (
	"io"
	"strconv"
	"strings"
	"zbxtable/internal/model"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetMetricMappings 获取指标映射列表
func GetMetricMappings(c *gin.Context) {
	instanceIDStr := c.Query("zid")

	var mappings []model.MetricMapping
	var err error

	if instanceIDStr != "" {
		instanceID, _ := strconv.Atoi(instanceIDStr)
		mappings, err = model.GetMetricMappingsByInstance(instanceID)
	} else {
		mappings, err = model.GetAllMetricMappings()
	}

	if err != nil {
		response.DatabaseError(c, "获取映射配置失败: "+err.Error())
		return
	}

	response.SuccessWithPage(c, mappings, int64(len(mappings)))
}

// GetMetricMappingByID 获取单个指标映射配置
func GetMetricMappingByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	mapping, err := model.GetMetricMappingByID(id)
	if err != nil {
		response.NotFound(c, "映射配置不存在")
		return
	}

	response.Success(c, mapping)
}

// CreateOrUpdateMetricMapping 创建或更新指标映射
func CreateOrUpdateMetricMapping(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	var mapping model.MetricMapping

	// 解析基本字段
	idStr := gjson.Get(string(body), "id").String()
	if idStr != "" {
		mapping.ID, _ = strconv.ParseInt(idStr, 10, 64)
	}

	mapping.ZID, _ = strconv.Atoi(gjson.Get(string(body), "zid").String())
	mapping.SystemType = gjson.Get(string(body), "system_type").String()
	mapping.HostGroupIDs = gjson.Get(string(body), "host_group_ids").String()
	mapping.MetricConfig = gjson.Get(string(body), "metric_config").String()

	// 解析自动化配置
	mapping.AutoInit, _ = strconv.Atoi(gjson.Get(string(body), "auto_init").String())
	mapping.InitCron = gjson.Get(string(body), "init_cron").String()
	if mapping.InitCron == "" {
		mapping.InitCron = "0 0 2 * * *"
	}
	mapping.InitOnNewHost, _ = strconv.Atoi(gjson.Get(string(body), "init_on_new_host").String())

	// 解析最大重试次数
	maxRetryStr := gjson.Get(string(body), "max_retry").String()
	if maxRetryStr != "" {
		mapping.MaxRetry, _ = strconv.Atoi(maxRetryStr)
	} else {
		mapping.MaxRetry = 3
	}

	// 验证必填字段
	if mapping.ZID == 0 {
		response.ValidationError(c, "请选择实例")
		return
	}
	if mapping.SystemType == "" {
		response.ValidationError(c, "请选择系统类型")
		return
	}

	err = model.CreateOrUpdateMetricMapping(&mapping)
	if err != nil {
		response.DatabaseError(c, "保存失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "保存成功", mapping)
}

// DeleteMetricMapping 删除指标映射
func DeleteMetricMapping(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	err = model.DeleteMetricMapping(id)
	if err != nil {
		response.DatabaseError(c, "删除失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// ExecuteMetricMappingManual 手动执行指标映射
func ExecuteMetricMappingManual(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	mapping, err := model.GetMetricMappingByID(id)
	if err != nil {
		response.NotFound(c, "映射配置不存在")
		return
	}

	// 异步执行
	go func() {
		err := model.ExecuteMetricMapping(mapping, "manual")
		if err != nil {
			logger.Log.Errorf("手动执行指标映射失败: %v", err)
		}
	}()

	response.SuccessWithMessage(c, "初始化任务已提交，请稍后查看执行历史", nil)
}

// GetMappingHistory 获取执行历史
func GetMappingHistory(c *gin.Context) {
	mappingIDStr := c.Query("mapping_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var history []model.MetricMappingHistory
	var total int64
	var err error

	if mappingIDStr != "" {
		mappingID, _ := strconv.ParseInt(mappingIDStr, 10, 64)
		history, total, err = model.GetMappingHistory(mappingID, page, limit)
	} else {
		history, total, err = model.GetAllMappingHistory(page, limit)
	}

	if err != nil {
		response.DatabaseError(c, "获取历史记录失败: "+err.Error())
		return
	}

	response.SuccessWithPage(c, history, total)
}

// GetMappingRules 获取指标匹配规则列表
func GetMappingRules(c *gin.Context) {
	enabledOnly := c.Query("enabled_only") == "1"
	rules, err := model.GetMappingRules(enabledOnly)
	if err != nil {
		response.DatabaseError(c, "获取规则失败: "+err.Error())
		return
	}
	response.Success(c, rules)
}

// CreateOrUpdateMappingRule 创建或更新匹配规则
func CreateOrUpdateMappingRule(c *gin.Context) {
	var rule model.MappingRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		response.BadRequest(c, "参数解析失败")
		return
	}

	if rule.TargetField == "" || rule.MatchValue == "" {
		response.ValidationError(c, "目标字段和匹配值不能为空")
		return
	}

	err := model.CreateOrUpdateMappingRule(&rule)
	if err != nil {
		response.DatabaseError(c, "保存规则失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "保存成功", rule)
}

// DeleteMappingRule 删除匹配规则
func DeleteMappingRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	err = model.DeleteMappingRule(id)
	if err != nil {
		response.DatabaseError(c, "删除失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// SyncMetricMappingOnTemplates 手动触发模板级指标同步
func SyncMetricMappingOnTemplates(c *gin.Context) {
	instanceIDStr := c.Query("zid")
	if instanceIDStr == "" {
		response.BadRequest(c, "请指定 Zabbix 实例 ID")
		return
	}
	zid, _ := strconv.Atoi(instanceIDStr)

	// 获取所有启用的规则
	allRules, err := model.GetMappingRules(true)
	if err != nil {
		response.DatabaseError(c, "获取规则失败")
		return
	}

	// 过滤适用于该实例的规则 (全局规则或匹配当前 ZID)
	var filteredRules []model.MappingRule
	zidStr := strconv.Itoa(zid)
	for _, rule := range allRules {
		isMatch := false
		if rule.ZIDs == "" {
			isMatch = true // 全局规则
		} else {
			zids := strings.Split(rule.ZIDs, ",")
			for _, z := range zids {
				if strings.TrimSpace(z) == zidStr {
					isMatch = true
					break
				}
			}
		}
		if isMatch {
			filteredRules = append(filteredRules, rule)
		}
	}

	if len(filteredRules) == 0 {
		response.SuccessWithMessage(c, "该实例没有适用的指标映射规则", nil)
		return
	}

	// 获取 API 实例
	apiInstance, err := model.GetAPIByZID(zid)
	if err != nil {
		response.DatabaseError(c, "获取 Zabbix API 失败")
		return
	}

	// 异步执行同步
	go func() {
		count, err := model.ExecuteMappingOnTemplates(apiInstance, filteredRules)
		if err != nil {
			logger.Log.Errorf("模板同步失败 [zid=%d, rules=%d]: %v", zid, len(filteredRules), err)
		} else {
			logger.Log.Infof("模板同步完成 [zid=%d, rules=%d], 影响模板数: %d", zid, len(filteredRules), count)
		}
	}()

	response.SuccessWithMessage(c, "同步任务已在后台启动", nil)
}

// GetTemplateItemsDebug Debug：回查模板 items（包含 key_ / inventory_link）
func GetTemplateItemsDebug(c *gin.Context) {
	instanceIDStr := c.Query("zid")
	templateID := c.Query("templateid")
	if instanceIDStr == "" || templateID == "" {
		response.BadRequest(c, "请指定 zid 和 templateid")
		return
	}
	zid, _ := strconv.Atoi(instanceIDStr)

	apiInstance, err := model.GetAPIByZID(zid)
	if err != nil {
		response.DatabaseError(c, "获取 Zabbix API 失败")
		return
	}

	items, err := model.GetTemplateItemsDebug(apiInstance, templateID)
	if err != nil {
		response.DatabaseError(c, "获取模板 items 失败: "+err.Error())
		return
	}
	response.Success(c, items)
}
