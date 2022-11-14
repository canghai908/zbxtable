package controllers

import (
	"zbxtable/models"
)

// HostGroupsController operations for History
type HostGroupsController struct {
	BaseController
}

//HostGroupsRes is used
var HostGroupsRes models.HostGroupsList

// URLMapping ...
func (c *HostGroupsController) URLMapping() {
	c.Mapping("GetList", c.GetList)
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("GetHostByGroupID", c.GetHostByGroupID)
	c.Mapping("GetGroup", c.GetGroup)
}

// GetList Gaa
// @Title Get All groups
// @Description get groups
// @Param	X-Token		header  string			true		"x-token in header"
// @Success 200 {object} models.Alarm
// @Failure 403
// @router /list [get]
func (c *HostGroupsController) GetList() {
	var HostGroupsRes models.HostTreeList
	hs, cnt, err := models.GetAllHostGroupsList()

	if err != nil {
		HostGroupsRes.Code = 401
		HostGroupsRes.Message = err.Error()
		c.Data["json"] = HostGroupsRes
	} else {
		HostGroupsRes.Code = 200
		HostGroupsRes.Message = "获取数据成功"
		HostGroupsRes.Data.Items = hs
		HostGroupsRes.Data.Total = cnt
		c.Data["json"] = HostGroupsRes
	}
	c.ServeJSON()
}

// GetList Gaa
// @Title Get All groups
// @Description get groups
// @Param	X-Token		header  string			true		"x-token in header"
// @Success 200 {object} models.Alarm
// @Failure 403
// @router /all [get]
func (c *HostGroupsController) GetGroup() {
	var HostGroupsRes models.HostTreeList
	hs, cnt, err := models.GetAllGroupsList()

	if err != nil {
		HostGroupsRes.Code = 401
		HostGroupsRes.Message = err.Error()
		c.Data["json"] = HostGroupsRes
	} else {
		HostGroupsRes.Code = 200
		HostGroupsRes.Message = "获取数据成功"
		HostGroupsRes.Data.Items = hs
		HostGroupsRes.Data.Total = cnt
		c.Data["json"] = HostGroupsRes
	}
	c.ServeJSON()
}

// GetAll 获取主机组分页显示
// @Title Get All groups
// @Description 获获取主机组分页显示
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	page	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	limit	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	groups	query	string	false	"Limit the size of result set. Must be an integer"
// @Success 200 {object} models.Alarm
// @Failure 403
// @router / [get]
func (c *HostGroupsController) GetAll() {
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	groups := c.Ctx.Input.Query("groups")
	hs, cnt, err := models.GetAllHostGroups(page, limit, groups)
	if err != nil {
		HostGroupsRes.Code = 401
		HostGroupsRes.Message = err.Error()
		c.Data["json"] = HostGroupsRes
		c.ServeJSON()
	} else {
		HostGroupsRes.Code = 200
		HostGroupsRes.Message = "获取数据成功"
		HostGroupsRes.Data.Items = hs
		HostGroupsRes.Data.Total = cnt
		c.Data["json"] = HostGroupsRes
	}
	c.ServeJSON()
}

// GetHostByGroupID func
// @Title Get All Hosts
// @Description get hosts
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	id		    path 	string	        true		"The key for staticblock"
// @Success 200 {object} models.Alarm
// @Failure 403
// @router /list/:id [get]
func (c *HostGroupsController) GetHostByGroupID() {
	GroupID := c.Ctx.Input.Param(":id")
	hs, err := models.GetHostsByGroupID(GroupID)
	var HostsByGroupIDRes models.HostGroupBYGroupIDList
	if err != nil {
		HostsByGroupIDRes.Code = 401
		HostsByGroupIDRes.Message = err.Error()
	} else {
		HostsByGroupIDRes.Code = 200
		HostsByGroupIDRes.Message = "获取数据成功"
	}
	HostsByGroupIDRes.Data.Items = hs
	c.Data["json"] = HostsByGroupIDRes
	c.ServeJSON()
}
