package controllers

import (
	"strconv"
	"zbxtable/models"
)

type TaskLogController struct {
	BaseController
}

var TaskLogRes models.TaskRes

func (c *TaskLogController) URLMapping() {
	c.Mapping("GetTaskByReportID", c.GetTaskByReportID)
	c.Mapping("Delete", c.Delete)
	//c.Mapping("GetReportOne", c.GetReportOne)
	//c.Mapping("CreateReport", c.CreateReport)
	////c.Mapping("InitTopology", c.InitTopology)
	//c.Mapping("Delete", c.Delete)
	//c.Mapping("Put", c.Put)
	//c.Mapping("UpdateReportStatus", c.UpdateReportStatus)
}

// GetTopologyAll ...
// @Title 获取任务日志
// @Description 获取任务日志
// @Param	X-Token		header  string	true	"X-Token"
// @Param	page	query	string	false	"页数"
// @Param	limit	query	string	false	"每页数"
// @Param	name	query	string	false	"拓扑名称"
// @Success 200 {object} models.Topology
// @Failure 403
// @router / [get]
func (c *TaskLogController) GetTaskByReportID() {
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	reportid := c.Ctx.Input.Query("report_id")
	//cycle := c.Ctx.Input.Query("cycle")
	//status := c.Ctx.Input.Query("status")
	count, hs, err := models.GetTaskLogList(page, limit, reportid)
	if err != nil {
		TaskLogRes.Code = 500
		TaskLogRes.Message = err.Error()
	} else {
		TaskLogRes.Code = 200
		TaskLogRes.Message = "获取数据成功"
		TaskLogRes.Data.Items = hs
		TaskLogRes.Data.Total = count
	}
	c.Data["json"] = TaskLogRes
	c.ServeJSON()
}

// Delete ...
// @Title 删除报表日志
// @Description 删除报表任务
// @Param	X-Token	header  string	true	"X-Token"
// @Param	id		path 	string	true		"The id you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *TaskLogController) Delete() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	if err := models.DeleteTaskLog(id); err == nil {
		ReportRes.Code = 200
		ReportRes.Message = "删除成功"
	} else {
		ReportRes.Code = 500
		ReportRes.Message = err.Error()
	}
	c.Data["json"] = ReportRes
	c.ServeJSON()
}
