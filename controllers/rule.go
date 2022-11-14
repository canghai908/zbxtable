package controllers

import (
	"github.com/tidwall/gjson"
	"strconv"
	"zbxtable/models"
)

type RuleController struct {
	BaseController
}

func (c *RuleController) URLMapping() {
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("GetOne", c.GetOne)
	c.Mapping("Delete", c.Delete)
	c.Mapping("Put", c.Put)
	c.Mapping("Create", c.Create)
	c.Mapping("PutStatus", c.PutStatus)

}

var RulRes models.RuleResp

// CreateRule ...
// @Title 创建规则
// @Description 创建规则
// @Param	X-Token	header  string			true		"X-Token"
// @Param	body	body 	models.Report   true		"body for Topology content"
// @Success 200 {object} models.Report
// @Failure 403
// @router / [post]
func (c *RuleController) Create() {
	name := gjson.Get(string(c.Ctx.Input.RequestBody), "name").String()
	tenant_id := gjson.Get(string(c.Ctx.Input.RequestBody), "tenant_id").String()
	conditions := gjson.Get(string(c.Ctx.Input.RequestBody), "conditions").String()
	sweek := gjson.Get(string(c.Ctx.Input.RequestBody), "s_week").String()
	stime := gjson.Get(string(c.Ctx.Input.RequestBody), "s_time").String()
	etime := gjson.Get(string(c.Ctx.Input.RequestBody), "e_time").String()
	channel := gjson.Get(string(c.Ctx.Input.RequestBody), "channel").String()
	user_ids := gjson.Get(string(c.Ctx.Input.RequestBody), "user_ids").String()
	group_ids := gjson.Get(string(c.Ctx.Input.RequestBody), "group_ids").String()
	m_type := gjson.Get(string(c.Ctx.Input.RequestBody), "m_type").String()
	note := gjson.Get(string(c.Ctx.Input.RequestBody), "note").String()
	status := gjson.Get(string(c.Ctx.Input.RequestBody), "status").String()
	v := models.Rule{Name: name, Conditions: conditions,
		Sweek: sweek, Stime: stime, Etime: etime, Channel: channel, MType: m_type,
		UserIds: user_ids, GroupIds: group_ids, TenantID: tenant_id, Status: status, Note: note}
	_, err := models.AddRule(&v)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
	} else {
		RulRes.Code = 200
		RulRes.Message = "创建成功"
	}
	c.Data["json"] = RulRes
	c.ServeJSON()
}

// GetAll ...
// @Title 查询或获取告警接口
// @Description get Alarm
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	begin	query	string	false	"开始日期 格式 2006-01-02 15:04:05"
// @Param	end		query	string	false	"结束日期 格式 2006-01-02 15:04:05"
// @Param	page	query	string	false	"第几页"
// @Param	limit	query	string	false	"每页条数"
// @Param	hosts	query	string	false	"查询主机名包含某字符的主机"
// @Param	tenant_id	query	string	false	"租户id"
// @Success 200 {object} models.Alarm
// @Failure 403
// @router / [get]
func (c *RuleController) GetAll() {
	var err error
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	name := c.Ctx.Input.Query("name")
	tenant_id := c.Ctx.Input.Query("tenant_id")
	m_type := c.Ctx.Input.Query("m_type")
	status := c.Ctx.Input.Query("status")
	//level := c.Ctx.Input.Query("level")
	cnt, al, err := models.GetRule(page, limit, name, tenant_id, m_type, status)
	if err != nil {
		RulRes.Code = 200
		RulRes.Message = err.Error()
	} else {
		RulRes.Code = 200
		RulRes.Message = "ok"
		RulRes.Data.Items = al
		RulRes.Data.Total = cnt
	}
	c.Data["json"] = RulRes
	c.ServeJSON()
}

// GetOne ...
// @Title 获取rule
// @Description 获取rule
// @Param	X-Token		header  string	true	"X-Token"
// @Param	body		body 	models.UserGroup   true		"UserGroup"
// @Success 200 {object} models.UserResp
// @Failure 403
// @router /:id [get]
func (c *RuleController) GetOne() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
		RulRes.Data.Items = nil
		RulRes.Data.Total = 0
		c.Data["json"] = RulRes
		c.ServeJSON()
	}
	v, err := models.GetRuleByID(id)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
		RulRes.Data.Items = nil
		RulRes.Data.Total = 0
	} else {
		RulRes.Code = 200
		RulRes.Message = "获取成功"
		RulRes.Data.Items = v
		RulRes.Data.Total = 0
	}
	c.Data["json"] = RulRes
	c.ServeJSON()
}

// Put ...
// @Title 更新rule规则
// @Description 更新rule规则
// @Param	X-Token		header  string	true	"X-Token"
// @Param	body		body 	models.UserGroup   true		"UserGroup"
// @Success 200 {object} models.UserResp
// @Failure 403
// @router /:id [put]
func (c *RuleController) Put() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
		RulRes.Data.Items = nil
		RulRes.Data.Total = 0
		c.Data["json"] = RulRes
		c.ServeJSON()
	}
	name := gjson.Get(string(c.Ctx.Input.RequestBody), "name").String()
	tenant_id := gjson.Get(string(c.Ctx.Input.RequestBody), "tenant_id").String()
	conditions := gjson.Get(string(c.Ctx.Input.RequestBody), "conditions").String()
	sweek := gjson.Get(string(c.Ctx.Input.RequestBody), "s_week").String()
	stime := gjson.Get(string(c.Ctx.Input.RequestBody), "s_time").String()
	etime := gjson.Get(string(c.Ctx.Input.RequestBody), "e_time").String()
	channel := gjson.Get(string(c.Ctx.Input.RequestBody), "channel").String()
	user_ids := gjson.Get(string(c.Ctx.Input.RequestBody), "user_ids").String()
	group_ids := gjson.Get(string(c.Ctx.Input.RequestBody), "group_ids").String()
	note := gjson.Get(string(c.Ctx.Input.RequestBody), "note").String()
	status := gjson.Get(string(c.Ctx.Input.RequestBody), "status").String()
	m_type := gjson.Get(string(c.Ctx.Input.RequestBody), "m_type").String()
	v := models.Rule{ID: id, Name: name, Conditions: conditions,
		Sweek: sweek, Stime: stime, Etime: etime, Channel: channel, MType: m_type,
		UserIds: user_ids, GroupIds: group_ids, TenantID: tenant_id, Status: status, Note: note}
	err = models.UpdateRule(&v, Tuser)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
	} else {
		RulRes.Code = 200
		RulRes.Message = "修改成功"
	}
	RulRes.Data.Items = nil
	RulRes.Data.Total = 0
	c.Data["json"] = RulRes
	c.ServeJSON()
}

// Put ...
// @Title 更新rule规则状态
// @Description 更新rule规则状态
// @Param	X-Token		header  string	true	"X-Token"
// @Param	body		body 	models.UserGroup   true		"UserGroup"
// @Success 200 {object} models.UserResp
// @Failure 403
// @router /status/:id [put]
func (c *RuleController) PutStatus() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
		RulRes.Data.Items = nil
		RulRes.Data.Total = 0
		c.Data["json"] = RulRes
		c.ServeJSON()
	}
	status := gjson.Get(string(c.Ctx.Input.RequestBody), "status").String()
	v := models.Rule{ID: id, Status: status}
	err = models.UpdateRuleStatus(&v, Tuser)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
	} else {
		RulRes.Code = 200
		RulRes.Message = "修改成功"
	}
	RulRes.Data.Items = nil
	RulRes.Data.Total = 0
	c.Data["json"] = RulRes
	c.ServeJSON()
}

// Put ...
// @Title 删除rule规则
// @Description 删除rule规则
// @Param	X-Token		header  string	true	"X-Token"
// @Param	body		body 	models.UserGroup   true		"UserGroup"
// @Success 200 {object} models.UserResp
// @Failure 403
// @router /:id [delete]
func (c *RuleController) Delete() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
		RulRes.Data.Items = nil
		RulRes.Data.Total = 0
		c.Data["json"] = RulRes
		c.ServeJSON()
	}
	err = models.DeleteRule(id, Tuser)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
	} else {
		RulRes.Code = 200
		RulRes.Message = "删除成功"
	}
	RulRes.Data.Items = nil
	RulRes.Data.Total = 0
	c.Data["json"] = RulRes
	c.ServeJSON()
}
