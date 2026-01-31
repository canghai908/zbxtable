package handler

import (
	"strconv"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetTaskLogByReportID 获取任务日志
func GetTaskLogByReportID(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	reportid := c.Query("report_id")

	count, hs, err := model.GetTaskLogList(page, limit, reportid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, hs, count)
}

// DeleteTaskLog 删除任务日志
func DeleteTaskLog(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	if err := model.DeleteTaskLog(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
