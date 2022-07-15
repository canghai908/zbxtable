package controllers

import (
	"zbxtable/models"
)

//TemplateController a
type TemplateController struct {
	BaseController
}

//TemplateRes rest
var TemplateRes models.TemplateList

//URLMapping beego
func (c *TemplateController) URLMapping() {
	c.Mapping("Get", c.GetInfo)
	c.Mapping("GetALl", c.GetAll)
	c.Mapping("GetAllList", c.GetAllList)
	c.Mapping("GetItemByTempID", c.GetItemByTempID)
}

// GetInfo 获取模版
// @Title 获取模版
// @Description 获取模版
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	page	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	limit	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	templates	query	string	false	"Limit the size of result set. Must be an integer"
// @Success 200 {object} models.Manager
// @Failure 403 :id is empty
// @router / [get]
func (c *TemplateController) GetInfo() {
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	templates := c.Ctx.Input.Query("templates")
	b, cnt, err := models.TemplateGet(page, limit, templates)
	if err != nil {
		TemplateRes.Code = 500
		TemplateRes.Message = "获取模版错误"
		c.Data["json"] = TemplateRes
		c.ServeJSON()
		return
	}
	TemplateRes.Code = 200
	TemplateRes.Message = "获取成功"
	TemplateRes.Data.Items = b
	TemplateRes.Data.Total = cnt
	c.Data["json"] = TemplateRes
	c.ServeJSON()
}

// GetAll 获取所有模版
// @Title 获取模版
// @Description 获取所有模板列表
// @Param	X-Token		header  string			true		"x-token in header"
// @Success 200 {object} models.Manager
// @Failure 403 :id is empty
// @router /all [get]
func (c *TemplateController) GetAll() {
	b, cnt, err := models.TemplateAllGet()
	if err != nil {
		TemplateRes.Code = 500
		TemplateRes.Message = "获取模版错误"
		c.Data["json"] = TemplateRes
		c.ServeJSON()
		return
	}
	TemplateRes.Code = 200
	TemplateRes.Message = "获取成功"
	TemplateRes.Data.Items = b
	TemplateRes.Data.Total = cnt
	c.Data["json"] = TemplateRes
	c.ServeJSON()
}

// GetAll 获取所有模版
// @Title 获取模版
// @Description 获取所有模板列表
// @Param	X-Token		header  string			true		"x-token in header"
// @Success 200 {object} models.Manager
// @Failure 403 :id is empty
// @router /list [get]
func (c *TemplateController) GetAllList() {
	b, cnt, err := models.TemplateListGet()
	if err != nil {
		TemplateRes.Code = 500
		TemplateRes.Message = "获取模版错误"
	} else {
		TemplateRes.Code = 200
		TemplateRes.Message = "获取成功"
		TemplateRes.Data.Items = b
		TemplateRes.Data.Total = cnt
	}
	c.Data["json"] = TemplateRes
	c.ServeJSON()
}

// GetAll 获取所有模版
// @Title 获取模版
// @Description 获取所有模板列表
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	templateid	path	string	ture	"templateid"
// @Success 200 {object} models.Manager
// @Failure 403 :id is empty
// @router /item/:templateid [get]
func (c *TemplateController) GetItemByTempID() {
	templateid := c.Ctx.Input.Param(":templateid")
	b, cnt, err := models.TemplateByItem(templateid)
	if err != nil {
		TemplateRes.Code = 500
		TemplateRes.Message = "获取模版错误"
	} else {
		TemplateRes.Code = 200
		TemplateRes.Message = "获取成功"
		TemplateRes.Data.Items = b
		TemplateRes.Data.Total = cnt
	}
	c.Data["json"] = TemplateRes
	c.ServeJSON()
}
