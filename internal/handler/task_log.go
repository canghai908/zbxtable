package handler

import (
	"net/http"
	"strconv"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
)

// GetTaskLogByReportID 获取任务日志
func GetTaskLogByReportID(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	reportid := c.Query("report_id")

	var TaskLogRes model.TaskRes
	count, hs, err := model.GetTaskLogList(page, limit, reportid)
	if err != nil {
		TaskLogRes.Code = 500
		TaskLogRes.Message = err.Error()
	} else {
		TaskLogRes.Code = 200
		TaskLogRes.Message = "获取数据成功"
		TaskLogRes.Data.Items = hs
		TaskLogRes.Data.Total = count
	}
	c.JSON(http.StatusOK, TaskLogRes)
}

// DeleteTaskLog 删除任务日志
func DeleteTaskLog(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var ReportRes model.ReportRes
	if err := model.DeleteTaskLog(id); err == nil {
		ReportRes.Code = 200
		ReportRes.Message = "删除成功"
	} else {
		ReportRes.Code = 500
		ReportRes.Message = err.Error()
	}
	c.JSON(http.StatusOK, ReportRes)
}
