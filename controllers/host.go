package controllers

import (
	"encoding/json"
	"strconv"

	"github.com/canghai908/zbxtable/models"
)

// HostController operations for Host
type HostController struct {
	BaseController
}

//HostRes restp
var HostRes models.HostList

// URLMapping ...
func (c *HostController) URLMapping() {
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("GetApplication", c.GetApplication)
}

// Post ...
// @Title Post
// @Description create Alarm
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	body		body 	models.Alarm	true		"body for Alarm content"
// @Success 201 {int} models.Alarm
// @Failure 403 body is empty
// @router / [post]
func (c *HostController) Post() {
	var v models.Alarm
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		if _, err := models.AddAlarm(&v); err == nil {
			c.Ctx.Output.SetStatus(201)
			c.Data["json"] = v
		} else {
			c.Data["json"] = err.Error()
		}
	} else {
		c.Data["json"] = err.Error()
	}
	c.ServeJSON()
}

//ApplicationRes str
var ApplicationRes models.ApplicationList

// GetOne ...
// @Title Get One
// @Description get Alarm by id
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Alarm
// @Failure 403 :id is empty
// @router /:id [get]
func (c *HostController) GetOne() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v, err := models.GetAlarmByID(id)
	if err != nil {
		c.Data["json"] = err.Error()
	} else {
		c.Data["json"] = v
	}
	c.ServeJSON()
}

// GetAll ...
// @Title 获取所有主机
// @Description get hosts
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	page	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	limit	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	hosts	query	string	false	"主机列表"
// @Success 200 {object} models.Alarm
// @Failure 403
// @router / [get]
func (c *HostController) GetAll() {
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	hosts := c.Ctx.Input.Query("hosts")
	hs, count, err := models.HostsList(page, limit, hosts)
	if err != nil {
		HostRes.Code = 500
		HostRes.Message = "获取错误"
		c.Data["json"] = HostRes
		c.ServeJSON()
		return
	}
	HostRes.Code = 200
	HostRes.Message = "获取数据成功"
	HostRes.Data.Items = hs
	HostRes.Data.Total = count
	c.Data["json"] = HostRes
	c.ServeJSON()
}

// GetApplication str
// @Title 获取所有主机
// @Description get hosts
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	hostid	    path	string	ture	"hostid"
// @Success 200 {object} models.Alarm
// @Failure 403
// @router /application/:hostid [get]
func (c *HostController) GetApplication() {
	hostid := c.Ctx.Input.Param(":hostid")
	hs, count, err := models.GetApplicationByHostid(hostid)
	if err != nil {
		ApplicationRes.Code = 500
		ApplicationRes.Message = "获取错误"
		c.Data["json"] = ApplicationRes
		c.ServeJSON()
		return
	}
	ApplicationRes.Code = 200
	ApplicationRes.Message = "获取数据成功"
	ApplicationRes.Data.Items = hs
	ApplicationRes.Data.Total = count
	c.Data["json"] = ApplicationRes
	c.ServeJSON()
}
