package controllers

import (
	"github.com/tidwall/gjson"
	"strconv"
	"zbxtable/models"
)

type SystemController struct {
	BaseController
}

var SystemRes models.SystemList

func (c *SystemController) URLMapping() {
	c.Mapping("GetOne", c.GetOne)
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("PutOne", c.PutOne)
	c.Mapping("DeployInit", c.DeployInit)
	c.Mapping("GetEgress", c.GetEgress)
	c.Mapping("PutEgress", c.PutEgress)

}

// GetOne ...
// @Title Get One
// @Description get Alarm by id
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Alarm
// @Failure 403 :id is empty
// @router /:id [get]
func (c *SystemController) GetOne() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	val, err := models.GetSystemByID(int64(id))
	if err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
	} else {
		SystemRes.Message = "ok"
		SystemRes.Data.Items = val
		SystemRes.Data.Total = 1
	}
	c.Data["json"] = SystemRes
	c.ServeJSON()
}

// PutOne 更新初始化数据 ...
// @Title Get One
// @Description get Alarm by id
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	id			path 	string			true		"system id"
// @Param	body		body 	models.System	true		"body for System content"
// @Success 200 {object} models.System
// @Failure 403 :id is empty
// @router /:id [put]
func (c *SystemController) PutOne() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	cpuCore := gjson.Get(string(c.Ctx.Input.RequestBody), "cpu_core").String()
	uptimeId := gjson.Get(string(c.Ctx.Input.RequestBody), "uptime_id").String()
	cpuUtilizationId := gjson.Get(string(c.Ctx.Input.RequestBody), "cpu_utilization_id").String()
	groupId := gjson.Get(string(c.Ctx.Input.RequestBody), "group_id").String()
	memoryTotalId := gjson.Get(string(c.Ctx.Input.RequestBody), "memory_total_id").String()
	memoryUsedId := gjson.Get(string(c.Ctx.Input.RequestBody), "memory_used_id").String()
	memoryUtilizationId := gjson.Get(string(c.Ctx.Input.RequestBody), "memory_utilization_id").String()
	model := gjson.Get(string(c.Ctx.Input.RequestBody), "model").String()
	pingTemplateId := gjson.Get(string(c.Ctx.Input.RequestBody), "ping_template_id").String()
	v := models.System{ID: int64(id), CPUCore: cpuCore, CPUUtilizationID: cpuUtilizationId,
		GroupID: groupId, MemoryTotalID: memoryTotalId, UptimeID: uptimeId, Model: model,
		MemoryUsedID: memoryUsedId, MemoryUtilizationID: memoryUtilizationId,
		PingTemplateID: pingTemplateId}
	err := models.UpdateSystem(&v)
	if err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
	} else {
		SystemRes.Code = 200
		SystemRes.Message = "更新成功"
	}
	c.Data["json"] = SystemRes
	c.ServeJSON()
}

// GetAll ...
// @Title 查询初始化状态
// @Description 获取配置列表
// @Param	X-Token		header  string			true		"x-token in header"
// @Success 200 {object} models.System
// @Failure 403
// @router / [get]
func (c *SystemController) GetAll() {
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
	c.Data["json"] = SystemRes
	c.ServeJSON()
}

// GetAll ...
// @Title 初始化
// @Description 初始化
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	id			path 	string			true		"system id"
// @Param	body		body 	models.System	true		"body for System content"
// @Success 200 {object} models.System
// @Failure 403
// @router /init/:id [post]
func (c *SystemController) DeployInit() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
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
	c.Data["json"] = SystemRes
	c.ServeJSON()
}

// GetAll ...
// @Title 初始化
// @Description 初始化
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	id			path 	string			true		"system id"
// @Success 200 {object} models.System
// @Failure 403
// @router /egress/ [get]
func (c *SystemController) GetEgress() {
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
	c.Data["json"] = SystemRes
	c.ServeJSON()
}

// GetAll ...
// @Title 更新带宽配置
// @Description 更新带宽配置
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	body		body 	models.Egress	true		"body for System content"
// @Success 200 {object} models.Egress
// @Failure 403
// @router /egress/ [put]
func (c *SystemController) PutEgress() {
	nameone := gjson.Get(string(c.Ctx.Input.RequestBody), "name_one").String()
	in_one := gjson.Get(string(c.Ctx.Input.RequestBody), "in_one").String()
	out_one := gjson.Get(string(c.Ctx.Input.RequestBody), "out_one").String()
	nametwo := gjson.Get(string(c.Ctx.Input.RequestBody), "name_two").String()
	in_two := gjson.Get(string(c.Ctx.Input.RequestBody), "in_two").String()
	out_two := gjson.Get(string(c.Ctx.Input.RequestBody), "out_two").String()
	v := models.Egress{ID: 1,
		NameOne: nameone, InOne: in_one, OutOne: out_one,
		NameTwo: nametwo, InTwo: in_two, OutTwo: out_two}
	err := models.UpdateEgress(&v)
	if err != nil {
		SystemRes.Code = 500
		SystemRes.Message = err.Error()
	} else {
		SystemRes.Code = 200
		SystemRes.Message = "更新成功"
		SystemRes.Data.Items = ""
		SystemRes.Data.Total = 1
	}
	c.Data["json"] = SystemRes
	c.ServeJSON()
}
