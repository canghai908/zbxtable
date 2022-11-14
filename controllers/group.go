package controllers

import (
	"github.com/tidwall/gjson"
	"strconv"
	"zbxtable/models"
)

type GroupControllers struct {
	BaseController
}

var GrooupRes models.GroupResp

func (c *GroupControllers) URLMapping() {
	c.Mapping("Get", c.Get)
	c.Mapping("Post", c.Post)
	c.Mapping("Delete", c.Delete)
	c.Mapping("Put", c.Put)
	c.Mapping("PutMembers", c.PutMembers)
	//c.Mapping("GetReportOne", c.GetReportOne)
	//c.Mapping("CreateReport", c.CreateReport)
	////c.Mapping("InitTopology", c.InitTopology)
	//c.Mapping("Delete", c.Delete)
	//c.Mapping("Put", c.Put)
	//c.Mapping("UpdateReportStatus", c.UpdateReportStatus)
}

// Get ...
// @Title 获取组列表
// @Description 获取组列表
// @Param	X-Token		header  string	true	"X-Token"
// @Param	page	query	string	false	"页数"
// @Param	limit	query	string	false	"每页数"
// @Param	name	query	string	false	"拓扑名称"
// @Success 200 {object} models.Topology
// @Failure 403
// @router / [get]
func (c *GroupControllers) Get() {
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	name := c.Ctx.Input.Query("name")
	count, hs, err := models.GetGroup(page, limit, Tuser, name)
	if err != nil {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
		GrooupRes.Data.Items = nil
		GrooupRes.Data.Total = 0
	} else {
		GrooupRes.Code = 200
		GrooupRes.Message = "获取数据成功"
		GrooupRes.Data.Items = hs
		GrooupRes.Data.Total = count
	}
	c.Data["json"] = GrooupRes
	c.ServeJSON()
}

// Post ...
// @Title 新建用户组
// @Description 新建用户
// @Param	X-Token		header  string	true	"X-Token"
// @Param	body		body 	models.Manager   true		"manager"
// @Success 200 {object} models.Topology
// @Failure 403
// @router / [post]
func (c *GroupControllers) Post() {
	name := gjson.Get(string(c.Ctx.Input.RequestBody), "name").String()
	note := gjson.Get(string(c.Ctx.Input.RequestBody), "note").String()
	v := models.UserGroup{Name: name, Note: note}
	_, err := models.AddUserGroup(&v)
	if err != nil {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
	} else {
		GrooupRes.Code = 200
		GrooupRes.Message = "创建用户组成功"
	}
	GrooupRes.Data.Items = nil
	GrooupRes.Data.Total = 1
	c.Data["json"] = GrooupRes
	c.ServeJSON()
}

// Put ...
// @Title 更新用户组信息
// @Description 更新用户组信息
// @Param	X-Token		header  string	true	"X-Token"
// @Param	body		body 	models.UserGroup   true		"UserGroup"
// @Success 200 {object} models.UserResp
// @Failure 403
// @router /:id [put]
func (c *GroupControllers) Put() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
		GrooupRes.Data.Items = nil
		GrooupRes.Data.Total = 0
		c.Data["json"] = GrooupRes
		c.ServeJSON()
	}
	name := gjson.Get(string(c.Ctx.Input.RequestBody), "name").String()
	note := gjson.Get(string(c.Ctx.Input.RequestBody), "note").String()
	v := models.UserGroup{ID: id, Name: name, Note: note}
	err = models.UpdateUserGroup(&v, Tuser)
	if err != nil {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
	} else {
		GrooupRes.Code = 200
		GrooupRes.Message = "修改成功"
	}
	GrooupRes.Data.Items = nil
	GrooupRes.Data.Total = 0
	c.Data["json"] = GrooupRes
	c.ServeJSON()
}

// PutMembers ...
// @Title 更新组成员
// @Description 更新组成员
// @Param	X-Token		header  string	true	"X-Token"
// @Param	body		body 	models.UserGroup   true		"UserGroup"
// @Success 200 {object} models.UserResp
// @Failure 403
// @router /member/:id [put]
func (c *GroupControllers) PutMembers() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
		GrooupRes.Data.Items = nil
		GrooupRes.Data.Total = 0
		c.Data["json"] = GrooupRes
		c.ServeJSON()
	}
	member := gjson.Get(string(c.Ctx.Input.RequestBody), "member").String()
	//member := c.GetString("jsoninfo")
	v := models.UserGroup{ID: id, Member: member}
	err = models.UpdateGroupMember(&v, Tuser)
	if err != nil {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
	} else {
		GrooupRes.Code = 200
		GrooupRes.Message = "修改成功"
	}
	GrooupRes.Data.Items = nil
	GrooupRes.Data.Total = 0
	c.Data["json"] = GrooupRes
	c.ServeJSON()
}

// Delete ...
// @Title 删除群组
// @Description 删除群组
// @Param	X-Token		header  string	true	"X-Token"
// @Param	id		path 	string	true	"The id you want to delete"
// @Success 200 {object} models.UserResp
// @Failure 403
// @router /:id [delete]
func (c *GroupControllers) Delete() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
		GrooupRes.Data.Items = nil
		GrooupRes.Data.Total = 0
		c.Data["json"] = GrooupRes
		c.ServeJSON()
	}
	if err := models.DeleteGroup(id, Tuser); err == nil {
		GrooupRes.Code = 200
		GrooupRes.Message = "删除成功"
	} else {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
		GrooupRes.Message = "修改成功"
	}
	GrooupRes.Data.Items = nil
	GrooupRes.Data.Total = 0
	c.Data["json"] = GrooupRes
	c.ServeJSON()
}
