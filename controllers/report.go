package controllers

import (
	"github.com/astaxie/beego"
	"github.com/tidwall/gjson"
	"strconv"
	"zbxtable/models"
)

// 报表接口
type ReportController struct {
	beego.Controller
}

var ReportRes models.ReportRes

// URLMapping ...
func (c *ReportController) URLMapping() {
	c.Mapping("GetReportAll", c.GetReportAll)
	c.Mapping("GetReportOne", c.GetReportOne)
	c.Mapping("CreateReport", c.CreateReport)
	//c.Mapping("InitTopology", c.InitTopology)
	c.Mapping("Delete", c.Delete)
	c.Mapping("Put", c.Put)
	c.Mapping("UpdateReportStatus", c.UpdateReportStatus)
	c.Mapping("CheckNow", c.CheckNow)
}

// GetTopologyAll ...
// @Title 获取报表列表
// @Description 获取报表列表
// @Param	X-Token		header  string	true	"X-Token"
// @Param	page	query	string	false	"页数"
// @Param	limit	query	string	false	"每页数"
// @Param	name	query	string	false	"拓扑名称"
// @Success 200 {object} models.Topology
// @Failure 403
// @router / [get]
func (c *ReportController) GetReportAll() {
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	name := c.Ctx.Input.Query("name")
	count, hs, err := models.GetAllReportsLimt(page, limit, name)
	if err != nil {
		ReportRes.Code = 500
		ReportRes.Message = err.Error()
	} else {
		ReportRes.Code = 200
		ReportRes.Message = "获取数据成功"
		ReportRes.Data.Items = hs
		ReportRes.Data.Total = count
	}
	c.Data["json"] = ReportRes
	c.ServeJSON()
}

// GetTopologyOne ...
// @Title 获取报表详情
// @Description 获取报表详情
// @Param	X-Token		header  string	true	"X-Token"
// @Param	id			path 	string	true	"id"
// @Success 200 {object} models.Topology
// @Failure 403
// @router /:id [get]
func (c *ReportController) GetReportOne() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v, err := models.GetReportsByID(id)
	if err != nil {
		ReportRes.Code = 500
		ReportRes.Message = err.Error()
	} else {
		ReportRes.Code = 200
		ReportRes.Message = "获取成功"
		ReportRes.Data.Items = v
	}
	c.Data["json"] = ReportRes
	c.ServeJSON()
}

// CreateTopology ...
// @Title 创建报表
// @Description 创建报表
// @Param	X-Token	header  string			true		"X-Token"
// @Param	body	body 	models.Report   true		"body for Topology content"
// @Success 200 {object} models.Report
// @Failure 403
// @router / [post]
func (c *ReportController) CreateReport() {
	name := gjson.Get(string(c.Ctx.Input.RequestBody), "name").String()
	items := gjson.Get(string(c.Ctx.Input.RequestBody), "items").String()
	linkbandwidth := gjson.Get(string(c.Ctx.Input.RequestBody), "linkbandwidth").String()
	cycle := gjson.Get(string(c.Ctx.Input.RequestBody), "cycle").String()
	emails := gjson.Get(string(c.Ctx.Input.RequestBody), "emails").String()
	status := gjson.Get(string(c.Ctx.Input.RequestBody), "status").String()
	desc := gjson.Get(string(c.Ctx.Input.RequestBody), "desc").String()
	report_type := gjson.Get(string(c.Ctx.Input.RequestBody), "report_type").String()
	v := models.Report{Name: name, Items: items, LinkBandWidth: linkbandwidth,
		Emails: emails, Cycle: cycle, Status: status, Desc: desc, ReportType: report_type}
	_, err := models.AddReport(&v)
	if err != nil {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	} else {
		TopologyRes.Code = 200
		TopologyRes.Message = "创建成功"
	}
	c.Data["json"] = TopologyRes
	c.ServeJSON()
}

// DeployTopology ...
// @Title 发布拓扑
// @Description 创建拓扑图
// @Param	X-Token	header  string	true	"X-Token"
// @Param	body	body 	models.Topology true	"body for Topology content"
// @Success 200 {object} models.Topology
// @Failure 403
// @router /status [post]
func (c *ReportController) UpdateReportStatus() {
	idStr := gjson.Get(string(c.Ctx.Input.RequestBody), "id").String()
	id, _ := strconv.Atoi(idStr)
	v := models.Report{ID: id}
	err := models.UpdateReportsStatusByID(&v)
	if err != nil {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	} else {
		TopologyRes.Code = 200
		TopologyRes.Message = "更新成功"
	}
	TopologyRes.Data.Items = []models.Topology{}
	c.Data["json"] = TopologyRes
	c.ServeJSON()
}

// saveTopology..
// @Title 更新报表
// @Description 更新报表
// @Param	X-Token	header  string	true	"X-Token"
// @Param	body	body 	models.Topology  true	"body for Topology content"
// @Param	id		path 	string	true	"id"
// @Success 200 {object} models.Topology
// @Failure 403
// @router /:id [put]
func (c *ReportController) Put() {
	idStr := c.Ctx.Input.Param(":id")
	name := gjson.Get(string(c.Ctx.Input.RequestBody), "name").String()
	items := gjson.Get(string(c.Ctx.Input.RequestBody), "items").String()
	emails := gjson.Get(string(c.Ctx.Input.RequestBody), "emails").String()
	linkbandwidth := gjson.Get(string(c.Ctx.Input.RequestBody), "linkbandwidth").String()
	cycle := gjson.Get(string(c.Ctx.Input.RequestBody), "cycle").String()
	status := gjson.Get(string(c.Ctx.Input.RequestBody), "status").String()
	desc := gjson.Get(string(c.Ctx.Input.RequestBody), "desc").String()
	id, _ := strconv.Atoi(idStr)
	v := models.Report{ID: id, Name: name, Items: items, LinkBandWidth: linkbandwidth,
		Emails: emails, Cycle: cycle, Status: status, Desc: desc}
	if err := models.UpdateReportByID(&v); err == nil {
		ReportRes.Code = 200
		ReportRes.Message = "保存成功"
		c.Data["json"] = "OK"
	} else {
		ReportRes.Code = 500
		ReportRes.Message = err.Error()
	}
	c.Data["json"] = ReportRes
	c.ServeJSON()
}

// Delete ...
// @Title 删除报表任务
// @Description 删除报表任务
// @Param	X-Token	header  string	true	"X-Token"
// @Param	id		path 	string	true		"The id you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *ReportController) Delete() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	if err := models.DeleteReport(id); err == nil {
		ReportRes.Code = 200
		ReportRes.Message = "删除成功"
	} else {
		ReportRes.Code = 500
		ReportRes.Message = err.Error()
	}
	c.Data["json"] = ReportRes
	c.ServeJSON()
}

// checknow report..
// @Title 测试报表
// @Description 测试报表
// @Param	X-Token	header  string	true	"X-Token"
// @Param	body	body 	models.Topology  true	"body for Topology content"
// @Param	id		path 	string	true	"id"
// @Success 200 {object} models.Topology
// @Failure 403
// @router /checknow [post]
func (c *ReportController) CheckNow() {
	idStr := gjson.Get(string(c.Ctx.Input.RequestBody), "id").String()
	id, _ := strconv.Atoi(idStr)
	v := models.Report{ID: id}
	if err := models.CheckNowByID(&v); err == nil {
		ReportRes.Code = 200
		ReportRes.Message = "生成成功"
		c.Data["json"] = "OK"
	} else {
		ReportRes.Code = 500
		ReportRes.Message = err.Error()
	}
	c.Data["json"] = ReportRes
	c.ServeJSON()
}
