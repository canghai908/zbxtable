package controllers

import (
	"zbxtable/models"
)

//IndexController 首页数据获取
type IndexController struct {
	BaseController
}

//InfoRes resp
var (
	InfoRes models.InfoRes
	TopRes  models.TopRes
	TreeRes models.TreeRes
	OverRes models.OverviewRes
	VerRes  models.VerRes
	EgrRes  models.EgressRes
)

//URLMapping beego
func (c *IndexController) URLMapping() {
	c.Mapping("Get", c.GetInfo)
	c.Mapping("GetResrouceTop", c.GetResrouceTop)
	c.Mapping("Inventory", c.GetInventory)
	c.Mapping("GetOverview", c.GetOverview)
	c.Mapping("GerVersion", c.GerVersion)
	c.Mapping("GetEgress", c.GetEgress)
}

// GetInfo ...
// @Title 首页数据
// @Description 基本信息获取
// @Param	X-Token		header  string			true		"x-token in header"
// @Success 200 {object} models.Manager
// @Failure 403 :id is empty
// @router /baseinfo/ [get]
func (c *IndexController) GetInfo() {
	info, err := models.GetCountHost()
	if err != nil {
		InfoRes.Code = 500
		InfoRes.Message = err.Error()
	} else {
		InfoRes.Code = 200
		InfoRes.Message = "获取成功"
		InfoRes.Data.Items = info
	}
	c.Data["json"] = InfoRes
	c.ServeJSON()
}

// GetLinuxCPUTop ...
// @Title Linux CPU使用率top5
// @Description Linux CPU使用率top5数据
// @Param	X-Token		header  string			true		"x-token in header"
// @Success 200 {object} models.Manager
// @Failure 403 :id is empty
// @router /lincputop [get]
func (c *IndexController) GetLinuxCPUTop() {
	info, err := models.GetCountHost()
	if err != nil {
		InfoRes.Code = 500
		InfoRes.Message = err.Error()
	} else {
		InfoRes.Code = 200
		InfoRes.Message = "获取成功"
		InfoRes.Data.Items = info
	}
	c.Data["json"] = InfoRes
	c.ServeJSON()
}

// GetResrouceTop ...
// @Title 资源使用TopN接口
// @Description 主机资源使用TopN接口
// @Param	X-Token			header  string	true	"x-token in header"
// @Param	host_type		query	string	true	"主机类型 VM_WIN VM_LIN"
// @Param	metrics_type	query	string	true	"指标类型 CPU MEM"
// @Param	top_num			query	string	false	"TOPN"
// @Success 200 {object} models.TopRes
// @Failure 403 :id is empty
// @router /restop [get]
func (c *IndexController) GetResrouceTop() {
	host_type := c.Ctx.Input.Query("host_type")
	metrics_type := c.Ctx.Input.Query("metrics_type")
	top_num := c.Ctx.Input.Query("top_num")
	info, err := models.GetTopList(host_type, metrics_type, top_num)
	if err != nil {
		TopRes.Code = 500
		TopRes.Message = err.Error()
	} else {
		TopRes.Code = 200
		TopRes.Message = "获取成功"
		TopRes.Data.Items = info
	}
	c.Data["json"] = TopRes
	c.ServeJSON()
}

// GetResrouceTop ...
// @Title 获取图谱资源
// @Description 主机资源使用TopN接口
// @Param	X-Token			header  string	true	"x-token in header"
// @Success 200 {object} models.TopRes
// @Failure 403 :id is empty
// @router /inventory [get]
func (c *IndexController) GetInventory() {
	info, err := models.GetInventory()
	if err != nil {
		TreeRes.Code = 500
		TreeRes.Message = err.Error()
	} else {
		TreeRes.Code = 200
		TreeRes.Message = "获取成功"
		TreeRes.Data.Items = info
	}
	c.Data["json"] = TreeRes
	c.ServeJSON()
}

// GetResrouceTop ...
// @Title 获取汇总状态
// @Description 获取汇总状态数据
// @Param	X-Token			header  string	true	"x-token in header"
// @Success 200 {object} models.TopRes
// @Failure 403 :id is empty
// @router /overview [get]
func (c *IndexController) GetOverview() {
	info, err := models.GetOverviewData()
	if err != nil {
		OverRes.Code = 500
		OverRes.Message = err.Error()
	} else {
		OverRes.Code = 200
		OverRes.Message = "获取成功"
		OverRes.Data.Items = info
	}
	c.Data["json"] = OverRes
	c.ServeJSON()
}

// GetResrouceTop ...
// @Title 获取汇总状态
// @Description 获取汇总状态数据
// @Param	X-Token			header  string	true	"x-token in header"
// @Success 200 {object} models.TopRes
// @Failure 403 :id is empty
// @router /egress [get]
func (c *IndexController) GetEgress() {
	info, err := models.GetEgressData()
	if err != nil {
		EgrRes.Code = 500
		EgrRes.Message = err.Error()
		EgrRes.Data.Items = info
	} else {
		EgrRes.Code = 200
		EgrRes.Message = "获取成功"
		EgrRes.Data.Items = info
	}
	c.Data["json"] = EgrRes
	c.ServeJSON()
}

// GetResrouceTop ...
// @Title 获取版本
// @Description 获取版本
// @Param	X-Token			header  string	true	"x-token in header"
// @Success 200 {object} models.TopRes
// @Failure 403 :id is empty
// @router /zbx [get]
func (c *IndexController) GerVersion() {
	VerRes.Code = 200
	VerRes.Message = "获取成功"
	VerRes.Data.Items = models.ZBX_V
	c.Data["json"] = VerRes
	c.ServeJSON()
}
