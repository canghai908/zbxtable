package handler

import (
	"io"
	"net/http"
	"strconv"
	"zbxtable/internal/model"
	"zbxtable/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetMetricMappings 获取指标映射列表
func GetMetricMappings(c *gin.Context) {
	instanceIDStr := c.Query("instance_id")
	
	var mappings []model.MetricMapping
	var err error
	
	if instanceIDStr != "" {
		instanceID, _ := strconv.Atoi(instanceIDStr)
		mappings, err = model.GetMetricMappingsByInstance(instanceID)
	} else {
		mappings, err = model.GetAllMetricMappings()
	}
	
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"items": mappings,
			"total": len(mappings),
		},
	})
}

// GetMetricMappingByID 获取单个指标映射配置
func GetMetricMappingByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	
	mapping, err := model.GetMetricMappingByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    mapping,
	})
}

// CreateOrUpdateMetricMapping 创建或更新指标映射
func CreateOrUpdateMetricMapping(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	var mapping model.MetricMapping
	
	// 解析基本字段
	idStr := gjson.Get(string(body), "id").String()
	if idStr != "" {
		mapping.ID, _ = strconv.ParseInt(idStr, 10, 64)
	}
	
	mapping.InstanceID, _ = strconv.Atoi(gjson.Get(string(body), "instance_id").String())
	mapping.SystemType = gjson.Get(string(body), "system_type").String()
	mapping.MappingName = gjson.Get(string(body), "mapping_name").String()
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
	if mapping.InstanceID == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请选择实例"})
		return
	}
	if mapping.SystemType == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请选择系统类型"})
		return
	}
	if mapping.MappingName == "" {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "请输入映射名称"})
		return
	}
	
	err = model.CreateOrUpdateMetricMapping(&mapping)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "保存成功", "data": mapping})
}

// DeleteMetricMapping 删除指标映射
func DeleteMetricMapping(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	
	err = model.DeleteMetricMapping(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ExecuteMetricMappingManual 手动执行指标映射
func ExecuteMetricMappingManual(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "无效的ID"})
		return
	}
	
	mapping, err := model.GetMetricMappingByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "配置不存在"})
		return
	}
	
	// 异步执行
	go func() {
		err := model.ExecuteMetricMapping(mapping, "manual")
		if err != nil {
			logger.Log.Errorf("手动执行指标映射失败: %v", err)
		}
	}()
	
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "初始化任务已提交，请稍后查看执行历史"})
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
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"items": history,
			"total": total,
		},
	})
}
