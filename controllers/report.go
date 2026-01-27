package controllers

import (
	"strconv"
	"time"
	"zbxtable/models"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/tidwall/gjson"
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
	c.Mapping("GetHosts", c.GetHosts)
	c.Mapping("GetItems", c.GetItems)
}

// GetTopologyAll ...
// @Title 获取报表列表
// @Description 获取报表列表
// @Param	X-Token		header  string	true	"X-Token"
// @Param	page	query	string	false	"页数"
// @Param	limit	query	string	false	"每页数"
// @Param	name	query	string	false	"拓扑名称"
// @Param	report_type	query	string	false	"报表类型"
// @Success 200 {object} models.Topology
// @Failure 403
// @router / [get]
func (c *ReportController) GetReportAll() {
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	name := c.Ctx.Input.Query("name")
	reportType := c.Ctx.Input.Query("report_type")
	count, hs, err := models.GetAllReportsLimt(page, limit, name, reportType)
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
	host_ids := gjson.Get(string(c.Ctx.Input.RequestBody), "host_ids").String()
	item_ids := gjson.Get(string(c.Ctx.Input.RequestBody), "item_ids").String()
	cycle := gjson.Get(string(c.Ctx.Input.RequestBody), "cycle").String()
	emails := gjson.Get(string(c.Ctx.Input.RequestBody), "emails").String()
	status := gjson.Get(string(c.Ctx.Input.RequestBody), "status").String()
	desc := gjson.Get(string(c.Ctx.Input.RequestBody), "desc").String()
	report_type := gjson.Get(string(c.Ctx.Input.RequestBody), "report_type").String()
	report_mode := gjson.Get(string(c.Ctx.Input.RequestBody), "report_mode").String()
	if report_mode == "" {
		report_mode = "scheduled" // 默认为循环报表
	}
	startTimeStr := gjson.Get(string(c.Ctx.Input.RequestBody), "start").String()
	endTimeStr := gjson.Get(string(c.Ctx.Input.RequestBody), "end").String()

	var startTime, endTime time.Time
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if startTimeStr != "" {
		startTime, _ = time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, loc)
	}
	if endTimeStr != "" {
		endTime, _ = time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, loc)
	}
	v := models.Report{Name: name, Items: items, LinkBandWidth: linkbandwidth,
		HostIds: host_ids, ItemIds: item_ids,
		Emails: emails, Cycle: cycle, Status: status, Desc: desc, ReportType: report_type,
		Start: startTime, End: endTime, ReportMode: report_mode}
	id, err := models.AddReport(&v)
	if err != nil {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	} else {
		// 如果是实时报表，立即生成
		if report_mode == "realtime" && report_type == "host" {
			// 重新读取完整的报表数据
			report, err := models.GetReportsByID(int(id))
			if err == nil {
				// 设置执行状态为处理中
				report.ExecStatus = "1" // 处理中
				report.StartAt = time.Now()
				models.UpdateReportExecStatusByID(report)

				// 异步生成报表
				go func() {
					err := models.TaskHostReport(*report)
					if err != nil {
						logs.Error("实时报表生成失败:", err)
						report.ExecStatus = "3" // 失败
					} else {
						report.ExecStatus = "2" // 成功
					}
					report.EndAt = time.Now()
					models.UpdateReportExecStatusByID(report)
				}()
			}
		}
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
	host_ids := gjson.Get(string(c.Ctx.Input.RequestBody), "host_ids").String()
	item_ids := gjson.Get(string(c.Ctx.Input.RequestBody), "item_ids").String()
	cycle := gjson.Get(string(c.Ctx.Input.RequestBody), "cycle").String()
	status := gjson.Get(string(c.Ctx.Input.RequestBody), "status").String()
	desc := gjson.Get(string(c.Ctx.Input.RequestBody), "desc").String()
	report_type := gjson.Get(string(c.Ctx.Input.RequestBody), "report_type").String()
	report_mode := gjson.Get(string(c.Ctx.Input.RequestBody), "report_mode").String()
	if report_mode == "" {
		report_mode = "scheduled" // 默认为循环报表
	}
	startTimeStr := gjson.Get(string(c.Ctx.Input.RequestBody), "start").String()
	endTimeStr := gjson.Get(string(c.Ctx.Input.RequestBody), "end").String()

	var startTime, endTime time.Time
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if startTimeStr != "" {
		startTime, _ = time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, loc)
	}
	if endTimeStr != "" {
		endTime, _ = time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, loc)
	}

	id, _ := strconv.Atoi(idStr)
	v := models.Report{ID: id, Name: name, Items: items, LinkBandWidth: linkbandwidth,
		HostIds: host_ids, ItemIds: item_ids,
		Emails: emails, Cycle: cycle, Status: status, Desc: desc, ReportType: report_type,
		Start: startTime, End: endTime, ReportMode: report_mode}
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
// @Title test report
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
	err := models.CheckNowByID(&v)
	if err == nil {
		ReportRes.Code = 200
		ReportRes.Message = "生成成功"
	} else {
		ReportRes.Code = 500
		ReportRes.Message = err.Error()
	}
	ReportRes.Data.Items = nil
	ReportRes.Data.Total = 0
	c.Data["json"] = ReportRes
	c.ServeJSON()
}

// GetHosts ...
// @Title 获取主机列表
// @Description 获取主机列表用于报表配置
// @Param	X-Token		header  string	true	"X-Token"
// @Param	host_type	query	string	false	"主机类型"
// @Success 200 {object} models.HostList
// @Failure 403
// @router /hosts [get]
func (c *ReportController) GetHosts() {
	hostType := c.Ctx.Input.Query("host_type")
	if hostType == "" {
		hostType = "VM_LIN"
	}
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	if page == "" {
		page = "1"
	}
	if limit == "" {
		limit = "1000"
	}
	hosts, count, err := models.HostsList(hostType, page, limit, "", "", "", "")
	if err != nil {
		var res models.HostList
		res.Code = 500
		res.Message = err.Error()
		c.Data["json"] = res
		c.ServeJSON()
		return
	}
	var res models.HostList
	res.Code = 200
	res.Message = "获取成功"
	res.Data.Items = hosts
	res.Data.Total = count
	c.Data["json"] = res
	c.ServeJSON()
}

// GetItems ...
// @Title 获取指标列表
// @Description 根据主机ID获取指标列表
// @Param	X-Token		header  string	true	"X-Token"
// @Param	host_id		query	string	true	"主机ID"
// @Success 200 {object} models.ItemList
// @Failure 403
// @router /items [get]
func (c *ReportController) GetItems() {
	hostID := c.Ctx.Input.Query("host_id")
	if hostID == "" {
		var res models.ItemList
		res.Code = 500
		res.Message = "host_id参数不能为空"
		c.Data["json"] = res
		c.ServeJSON()
		return
	}
	items, count, err := models.GetAllItemByHostID(hostID)
	if err != nil {
		var res models.ItemList
		res.Code = 500
		res.Message = err.Error()
		c.Data["json"] = res
		c.ServeJSON()
		return
	}
	var res models.ItemList
	res.Code = 200
	res.Message = "获取成功"
	res.Data.Items = items
	res.Data.Total = count
	c.Data["json"] = res
	c.ServeJSON()
}
