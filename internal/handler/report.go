package handler

import (
	"io"
	"strconv"
	"time"
	"zbxtable/internal/model"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetReportGin 获取报表列表
func GetReportGin(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	name := c.Query("name")
	reportType := c.Query("report_type")

	count, hs, err := model.GetAllReportsLimt(page, limit, name, reportType)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, hs, count)
}

// GetReportOneGin 获取报表详情
func GetReportOneGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	v, err := model.GetReportsByID(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "获取成功", v)
}

// CreateReportGin 创建报表
func CreateReportGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	name := gjson.Get(string(body), "name").String()
	instanceIDStr := gjson.Get(string(body), "instance").String()
	items := gjson.Get(string(body), "items").String()
	linkbandwidth := gjson.Get(string(body), "linkbandwidth").String()
	host_ids := gjson.Get(string(body), "host_ids").String()
	item_ids := gjson.Get(string(body), "item_ids").String()
	cycle := gjson.Get(string(body), "cycle").String()
	emails := gjson.Get(string(body), "emails").String()
	status := gjson.Get(string(body), "status").String()
	desc := gjson.Get(string(body), "desc").String()
	report_type := gjson.Get(string(body), "report_type").String()
	report_mode := gjson.Get(string(body), "report_mode").String()
	if report_mode == "" {
		report_mode = "scheduled" // 默认为循环报表
	}
	startTimeStr := gjson.Get(string(body), "start").String()
	endTimeStr := gjson.Get(string(body), "end").String()

	var startTime, endTime *time.Time
	loc, _ := time.LoadLocation("Local")
	if startTimeStr != "" {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, loc)
		startTime = &t
	}
	if endTimeStr != "" {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, loc)
		endTime = &t
	}

	instanceID, _ := strconv.Atoi(instanceIDStr)
	v := model.Report{Name: name, Items: items, LinkBandWidth: linkbandwidth,
		HostIds: host_ids, ItemIds: item_ids, Instance: instanceID,
		Emails: emails, Cycle: cycle, Status: status, Desc: desc, ReportType: report_type,
		Start: startTime, End: endTime, ReportMode: report_mode}
	id, err := model.AddReport(&v)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 如果是实时报表，立即生成
	if report_mode == "realtime" && report_type == "host" {
		// 重新读取完整的报表数据
		report, err := model.GetReportsByID(int(id))
		if err == nil {
			// 设置执行状态为处理中
			report.ExecStatus = "1" // 处理中
			now := time.Now()
			report.StartAt = &now
			model.UpdateReportExecStatusByID(report)

			// 异步生成报表
			go func() {
				err := model.TaskHostReport(*report)
				if err != nil {
					logger.Log.Error("实时报表生成失败:", err)
					report.ExecStatus = "3" // 失败
				} else {
					report.ExecStatus = "2" // 成功
				}
				endNow := time.Now()
				report.EndAt = &endNow
				model.UpdateReportExecStatusByID(report)
			}()
		}
	}
	response.SuccessWithMessage(c, "创建成功", nil)
}

// UpdateReportGin 更新报表
func UpdateReportGin(c *gin.Context) {
	idStr := c.Param("id")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	name := gjson.Get(string(body), "name").String()
	instanceIDStr := gjson.Get(string(body), "instance").String()
	items := gjson.Get(string(body), "items").String()
	emails := gjson.Get(string(body), "emails").String()
	linkbandwidth := gjson.Get(string(body), "linkbandwidth").String()
	host_ids := gjson.Get(string(body), "host_ids").String()
	item_ids := gjson.Get(string(body), "item_ids").String()
	cycle := gjson.Get(string(body), "cycle").String()
	status := gjson.Get(string(body), "status").String()
	desc := gjson.Get(string(body), "desc").String()
	report_type := gjson.Get(string(body), "report_type").String()
	report_mode := gjson.Get(string(body), "report_mode").String()
	if report_mode == "" {
		report_mode = "scheduled" // 默认为循环报表
	}
	startTimeStr := gjson.Get(string(body), "start").String()
	endTimeStr := gjson.Get(string(body), "end").String()

	var startTime, endTime *time.Time
	loc, _ := time.LoadLocation("Local")
	if startTimeStr != "" {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, loc)
		startTime = &t
	}
	if endTimeStr != "" {
		t, _ := time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, loc)
		endTime = &t
	}

	id, _ := strconv.Atoi(idStr)
	instanceID, _ := strconv.Atoi(instanceIDStr)
	v := model.Report{ID: id, Name: name, Items: items, LinkBandWidth: linkbandwidth,
		HostIds: host_ids, ItemIds: item_ids, Instance: instanceID,
		Emails: emails, Cycle: cycle, Status: status, Desc: desc, ReportType: report_type,
		Start: startTime, End: endTime, ReportMode: report_mode}
	if err := model.UpdateReportByID(&v); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "保存成功", nil)
}

// DeleteReportGin 删除报表
func DeleteReportGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	if err := model.DeleteReport(id); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// CheckNowGin 测试报表
func CheckNowGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	idStr := gjson.Get(string(body), "id").String()
	id, _ := strconv.Atoi(idStr)

	v := model.Report{ID: id}
	err = model.CheckNowByID(&v)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "生成成功", nil)
}

// UpdateReportStatusGin 更新报表状态
func UpdateReportStatusGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	idStr := gjson.Get(string(body), "id").String()
	id, _ := strconv.Atoi(idStr)

	v := model.Report{ID: id}
	err = model.UpdateReportsStatusByID(&v)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}
