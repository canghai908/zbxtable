package controllers

import (
	"io"
	"net/http"
	"strconv"
	"zbxtable/models"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetAllSystem 获取系统配置列表
func GetAllSystem(c *gin.Context) {
	var SystemRes models.SystemList
	cnt, val, err := models.GetALlSystem()
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
	var SystemRes models.SystemList
	val, err := models.GetSystemByID(int64(id))
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

// UpdateSystem 更新系统配置
func UpdateSystem(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	cpuCore := gjson.Get(string(body), "cpu_core").String()
	uptimeId := gjson.Get(string(body), "uptime_id").String()
	cpuUtilizationId := gjson.Get(string(body), "cpu_utilization_id").String()
	groupId := gjson.Get(string(body), "group_id").String()
	memoryTotalId := gjson.Get(string(body), "memory_total_id").String()
	memoryUsedId := gjson.Get(string(body), "memory_used_id").String()
	memoryUtilizationId := gjson.Get(string(body), "memory_utilization_id").String()
	model := gjson.Get(string(body), "model").String()
	pingTemplateId := gjson.Get(string(body), "ping_template_id").String()
	
	var SystemRes models.SystemList
	v := models.System{ID: int64(id), CPUCore: cpuCore, CPUUtilizationID: cpuUtilizationId,
		GroupID: groupId, MemoryTotalID: memoryTotalId, UptimeID: uptimeId, Model: model,
		MemoryUsedID: memoryUsedId, MemoryUtilizationID: memoryUtilizationId,
		PingTemplateID: pingTemplateId}
	err = models.UpdateSystem(&v)
	if err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
	} else {
		SystemRes.Code = 200
		SystemRes.Message = "更新成功"
	}
	c.JSON(http.StatusOK, SystemRes)
}

// SystemInit 初始化系统
func SystemInit(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var SystemRes models.SystemList
	err := models.SystemInit(int64(id))
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
	var SystemRes models.SystemList
	val, err := models.GetEgress()
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
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	nameone := gjson.Get(string(body), "name_one").String()
	in_one := gjson.Get(string(body), "in_one").String()
	out_one := gjson.Get(string(body), "out_one").String()
	nametwo := gjson.Get(string(body), "name_two").String()
	in_two := gjson.Get(string(body), "in_two").String()
	out_two := gjson.Get(string(body), "out_two").String()
	var SystemRes models.SystemList
	v := models.Egress{ID: 1,
		NameOne: nameone, InOne: in_one, OutOne: out_one,
		NameTwo: nametwo, InTwo: in_two, OutTwo: out_two}
	err = models.UpdateEgress(&v)
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
	var SystemRes models.SystemList
	val, err := models.GetConfigList()
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
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	value := gjson.Get(string(body), "value").String()
	var SystemRes models.SystemList
	v := models.Config{ID: int64(id), Value: value}
	err = models.UpdateConfig(&v)
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
