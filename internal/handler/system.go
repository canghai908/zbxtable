package handler

import (
	"zbxtable/pkg/response"
	"io"
	"net/http"
	"strconv"
	"zbxtable/internal/model"

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
	instanceIDStr := gjson.Get(string(body), "instance_id").String()
	if instanceIDStr == "" {
		response.BadRequest(c, "请选择 Zabbix 实例")
		return
	}
	
	// 查找实例获取数字ID
	tenant, err := model.GetZabbixTenantByTenantID(instanceIDStr)
	if err != nil {
		response.InternalError(c, "实例不存在")
		return
	}
	
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
		InstanceID:          tenant.ID,
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
	
	instanceIDStr := gjson.Get(string(body), "instance_id").String()
	if instanceIDStr == "" {
		response.BadRequest(c, "请选择 Zabbix 实例")
		return
	}
	
	// 查找实例获取数字ID
	tenant, err := model.GetZabbixTenantByTenantID(instanceIDStr)
	if err != nil {
		response.InternalError(c, "实例不存在")
		return
	}
	
	var SystemRes model.SystemList
	err = model.SystemInitWithInstance(int64(id), tenant.ID)
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

// UpdateEgress 更新带宽配置
func UpdateEgress(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	nameone := gjson.Get(string(body), "name_one").String()
	in_one := gjson.Get(string(body), "in_one").String()
	out_one := gjson.Get(string(body), "out_one").String()
	nametwo := gjson.Get(string(body), "name_two").String()
	in_two := gjson.Get(string(body), "in_two").String()
	out_two := gjson.Get(string(body), "out_two").String()
	var SystemRes model.SystemList
	v := model.Egress{ID: 1,
		NameOne: nameone, InOne: in_one, OutOne: out_one,
		NameTwo: nametwo, InTwo: in_two, OutTwo: out_two}
	err = model.UpdateEgress(&v)
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

// GetAllConfig 获取系统参数配置
func GetAllConfig(c *gin.Context) {
	var SystemRes model.SystemList
	val, err := model.GetConfigList()
	if err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
	} else {
		SystemRes.Code = 200
		SystemRes.Message = "获取成功"
		SystemRes.Data.Items = val
		SystemRes.Data.Total = int64(len(val))
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
