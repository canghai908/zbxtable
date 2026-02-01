package handler

import (
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

// ============ 新的出口配置 API ============

// GetAllEgressConfigs 获取所有出口配置
func GetAllEgressConfigs(c *gin.Context) {
	configs, err := model.GetAllEgressConfigs()
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
	tenantID := gjson.Get(string(body), "tenant_id").String()
	hostID := gjson.Get(string(body), "host_id").String()
	inItemID := gjson.Get(string(body), "in_item_id").String()
	outItemID := gjson.Get(string(body), "out_item_id").String()
	sortOrder := int(gjson.Get(string(body), "sort_order").Int())

	if name == "" || tenantID == "" || hostID == "" || inItemID == "" || outItemID == "" {
		response.BadRequest(c, "缺少必填字段")
		return
	}

	config := &model.EgressConfig{
		Name:       name,
		InstanceID: tenantID,
		HostID:     hostID,
		InItemID:   inItemID,
		OutItemID:  outItemID,
		Status:     1,
		SortOrder:  sortOrder,
	}

	err = model.AddEgressConfig(config)
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
	tenantID := gjson.Get(string(body), "tenant_id").String()
	hostID := gjson.Get(string(body), "host_id").String()
	inItemID := gjson.Get(string(body), "in_item_id").String()
	outItemID := gjson.Get(string(body), "out_item_id").String()
	status := int(gjson.Get(string(body), "status").Int())
	sortOrder := int(gjson.Get(string(body), "sort_order").Int())

	if name == "" || tenantID == "" || hostID == "" || inItemID == "" || outItemID == "" {
		response.BadRequest(c, "缺少必填字段")
		return
	}

	config := &model.EgressConfig{
		ID:         id,
		Name:       name,
		InstanceID: tenantID,
		HostID:     hostID,
		InItemID:   inItemID,
		OutItemID:  outItemID,
		Status:     status,
		SortOrder:  sortOrder,
	}

	err = model.UpdateEgressConfig(config)
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
