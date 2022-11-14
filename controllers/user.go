package controllers

import (
	"github.com/tidwall/gjson"
	"strconv"
	"time"
	"zbxtable/models"
	"zbxtable/utils"
)

type UserController struct {
	BaseController
}

func (c *UserController) URLMapping() {
	c.Mapping("Get", c.Get)
	//c.Mapping("GetOne", c.GetOne)
	c.Mapping("Post", c.Post)
	c.Mapping("Delete", c.Delete)
	c.Mapping("Put", c.Put)
	c.Mapping("StatusPut", c.StatusPut)
	//c.Mapping("GetReportOne", c.GetReportOne)
	//c.Mapping("CreateReport", c.CreateReport)
	////c.Mapping("InitTopology", c.InitTopology)
	//c.Mapping("Delete", c.Delete)
	//c.Mapping("Put", c.Put)
	//c.Mapping("UpdateReportStatus", c.UpdateReportStatus)
}

var UserRes models.UserResp

// Get ...
// @Title 获取用户列表
// @Description 获取用户列表
// @Param	X-Token		header  string	true	"X-Token"
// @Param	page	query	string	false	"页数"
// @Param	limit	query	string	false	"每页数"
// @Param	name	query	string	false	"拓扑名称"
// @Success 200 {object} models.Topology
// @Failure 403
// @router / [get]
func (c *UserController) Get() {
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	username := c.Ctx.Input.Query("username")
	status := c.Ctx.Input.Query("status")
	count, hs, err := models.GetUser(page, limit, Tuser, username, status)
	if err != nil {
		UserRes.Code = 500
		UserRes.Message = err.Error()
		UserRes.Data.Items = nil
		UserRes.Data.Total = 0
	} else {
		UserRes.Code = 200
		UserRes.Message = "获取数据成功"
		UserRes.Data.Items = hs
		UserRes.Data.Total = count
	}
	c.Data["json"] = UserRes
	c.ServeJSON()
}

// Post ...
// @Title 新建用户
// @Description 新建用户
// @Param	X-Token		header  string	true	"X-Token"
// @Param	body		body 	models.Manager   true		"manager"
// @Success 200 {object} models.Topology
// @Failure 403
// @router / [post]
func (c *UserController) Post() {
	username := gjson.Get(string(c.Ctx.Input.RequestBody), "username").String()
	password := gjson.Get(string(c.Ctx.Input.RequestBody), "password").String()
	role := gjson.Get(string(c.Ctx.Input.RequestBody), "role").String()
	email := gjson.Get(string(c.Ctx.Input.RequestBody), "email").String()
	wechat := gjson.Get(string(c.Ctx.Input.RequestBody), "wechat").String()
	phone := gjson.Get(string(c.Ctx.Input.RequestBody), "phone").String()
	ding_talk := gjson.Get(string(c.Ctx.Input.RequestBody), "ding_talk").String()
	p, _ := utils.PasswordHash(password)
	var operation string
	switch role {
	case "admin":
		operation = "['add', 'edit', 'delete','update']"
	case "user":
		operation = "[]"
	default:
		operation = "[]"
	}
	v := models.Manager{Username: username, Password: p, Operation: operation,
		Email: email, Wechat: wechat, Phone: phone, DingTalk: ding_talk,
		Status: 0, Role: role, Created: time.Now(),
		Avatar: "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
	}
	_, err := models.AddUser(&v)
	if err != nil {
		UserRes.Code = 500
		UserRes.Message = err.Error()
	} else {
		UserRes.Code = 200
		UserRes.Message = "创建用户成功"
	}
	UserRes.Data.Items = nil
	UserRes.Data.Total = 1
	c.Data["json"] = UserRes
	c.ServeJSON()
}

// Put ...
// @Title 更新用户信息
// @Description 更新用户信息
// @Param	X-Token		header  string	true	"X-Token"
// @Param	body		body 	models.Manager   true		"manager"
// @Success 200 {object} models.UserResp
// @Failure 403
// @router /:id [put]
func (c *UserController) Put() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		UserRes.Code = 500
		UserRes.Message = err.Error()
		UserRes.Data.Items = nil
		UserRes.Data.Total = 0
		c.Data["json"] = UserRes
		c.ServeJSON()
	}
	password := gjson.Get(string(c.Ctx.Input.RequestBody), "password").String()
	role := gjson.Get(string(c.Ctx.Input.RequestBody), "role").String()
	email := gjson.Get(string(c.Ctx.Input.RequestBody), "email").String()
	wechat := gjson.Get(string(c.Ctx.Input.RequestBody), "wechat").String()
	phone := gjson.Get(string(c.Ctx.Input.RequestBody), "phone").String()
	dingTalk := gjson.Get(string(c.Ctx.Input.RequestBody), "ding_talk").String()
	var operation string
	switch role {
	case "admin":
		operation = "['add', 'edit', 'delete','update']"
	case "user":
		operation = "[]"
	default:
		operation = "[]"
	}
	var pass string
	if password != "" {
		pass, _ = utils.PasswordHash(password)
	} else {
		pass = ""
	}

	v := models.Manager{ID: id, Password: pass, Email: email, Wechat: wechat,
		Phone: phone, DingTalk: dingTalk, Role: role, Operation: operation}
	err = models.UpdateUser(&v, Tuser)
	if err != nil {
		UserRes.Code = 500
		UserRes.Message = err.Error()
	} else {
		UserRes.Code = 200
		UserRes.Message = "修改成功"
	}
	UserRes.Data.Items = nil
	UserRes.Data.Total = 0
	c.Data["json"] = UserRes
	c.ServeJSON()
}

// Put ...
// @Title 更新用户状态
// @Description 更新用户状态
// @Param	X-Token		header  string	true	"X-Token"
// @Param	body		body 	models.Manager   true		"manager"
// @Success 200 {object} models.UserResp
// @Failure 403
// @router /status/:id [put]
func (c *UserController) StatusPut() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		UserRes.Code = 500
		UserRes.Message = err.Error()
		UserRes.Data.Items = nil
		UserRes.Data.Total = 0
		c.Data["json"] = UserRes
		c.ServeJSON()
	}
	status := gjson.Get(string(c.Ctx.Input.RequestBody), "status").Int()
	v := models.Manager{ID: id, Status: status}
	err = models.UpdateUserStatus(&v, Tuser)
	if err != nil {
		UserRes.Code = 500
		UserRes.Message = err.Error()
	} else {
		UserRes.Code = 200
		UserRes.Message = "更新成功"
	}
	UserRes.Data.Items = nil
	UserRes.Data.Total = 0
	c.Data["json"] = UserRes
	c.ServeJSON()
}

// Delete ...
// @Title 删除用户
// @Description 删除用户
// @Param	X-Token	header  string	true	"X-Token"
// @Param	id		path 	string	true	"The id you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *UserController) Delete() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		UserRes.Code = 500
		UserRes.Message = err.Error()
		UserRes.Data.Items = nil
		UserRes.Data.Total = 0
		c.Data["json"] = UserRes
		c.ServeJSON()
	}
	if err := models.DeleteUser(id, Tuser); err == nil {
		UserRes.Code = 200
		UserRes.Message = "删除成功"
	} else {
		UserRes.Code = 500
		UserRes.Message = err.Error()
	}
	UserRes.Data.Items = nil
	UserRes.Data.Total = 0
	c.Data["json"] = UserRes
	c.ServeJSON()
}
