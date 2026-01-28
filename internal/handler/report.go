package handler

import (
	"io"
	"net/http"
	"strconv"
	"time"
	"zbxtable/internal/model"
	"zbxtable/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetReportGin 获取报表列表
func GetReportGin(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	name := c.Query("name")
	reportType := c.Query("report_type")

	var res models.ReportRes
	count, hs, err := models.GetAllReportsLimt(page, limit, name, reportType)
	if err != nil {
		res.Code = 500
		res.Message = err.Error()
	} else {
		res.Code = 200
		res.Message = "获取数据成功"
		res.Data.Items = hs
		res.Data.Total = count
	}
	c.JSON(http.StatusOK, res)
}

// GetReportOneGin 获取报表详情
func GetReportOneGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var res models.ReportRes
	v, err := models.GetReportsByID(id)
	if err != nil {
		res.Code = 500
		res.Message = err.Error()
	} else {
		res.Code = 200
		res.Message = "获取成功"
		res.Data.Items = v
	}
	c.JSON(http.StatusOK, res)
}

// CreateReportGin 创建报表
func CreateReportGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}

	name := gjson.Get(string(body), "name").String()
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

	var startTime, endTime time.Time
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if startTimeStr != "" {
		startTime, _ = time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, loc)
	}
	if endTimeStr != "" {
		endTime, _ = time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, loc)
	}

	var res models.ReportRes
	v := models.Report{Name: name, Items: items, LinkBandWidth: linkbandwidth,
		HostIds: host_ids, ItemIds: item_ids,
		Emails: emails, Cycle: cycle, Status: status, Desc: desc, ReportType: report_type,
		Start: startTime, End: endTime, ReportMode: report_mode}
	id, err := models.AddReport(&v)
	if err != nil {
		res.Code = 500
		res.Message = err.Error()
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
						utils.Log.Error("实时报表生成失败:", err)
						report.ExecStatus = "3" // 失败
					} else {
						report.ExecStatus = "2" // 成功
					}
					report.EndAt = time.Now()
					models.UpdateReportExecStatusByID(report)
				}()
			}
		}
		res.Code = 200
		res.Message = "创建成功"
	}
	c.JSON(http.StatusOK, res)
}

// UpdateReportGin 更新报表
func UpdateReportGin(c *gin.Context) {
	idStr := c.Param("id")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}

	name := gjson.Get(string(body), "name").String()
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

	var startTime, endTime time.Time
	loc, _ := time.LoadLocation("Asia/Shanghai")
	if startTimeStr != "" {
		startTime, _ = time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, loc)
	}
	if endTimeStr != "" {
		endTime, _ = time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, loc)
	}

	id, _ := strconv.Atoi(idStr)
	var res models.ReportRes
	v := models.Report{ID: id, Name: name, Items: items, LinkBandWidth: linkbandwidth,
		HostIds: host_ids, ItemIds: item_ids,
		Emails: emails, Cycle: cycle, Status: status, Desc: desc, ReportType: report_type,
		Start: startTime, End: endTime, ReportMode: report_mode}
	if err := models.UpdateReportByID(&v); err == nil {
		res.Code = 200
		res.Message = "保存成功"
	} else {
		res.Code = 500
		res.Message = err.Error()
	}
	c.JSON(http.StatusOK, res)
}

// DeleteReportGin 删除报表
func DeleteReportGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var res models.ReportRes
	if err := models.DeleteReport(id); err == nil {
		res.Code = 200
		res.Message = "删除成功"
	} else {
		res.Code = 500
		res.Message = err.Error()
	}
	c.JSON(http.StatusOK, res)
}

// CheckNowGin 测试报表
func CheckNowGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}

	idStr := gjson.Get(string(body), "id").String()
	id, _ := strconv.Atoi(idStr)

	var res models.ReportRes
	v := models.Report{ID: id}
	err = models.CheckNowByID(&v)
	if err == nil {
		res.Code = 200
		res.Message = "生成成功"
	} else {
		res.Code = 500
		res.Message = err.Error()
	}
	res.Data.Items = nil
	res.Data.Total = 0
	c.JSON(http.StatusOK, res)
}

// UpdateReportStatusGin 更新报表状态
func UpdateReportStatusGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}

	idStr := gjson.Get(string(body), "id").String()
	id, _ := strconv.Atoi(idStr)

	var res models.TopologyList
	v := models.Report{ID: id}
	err = models.UpdateReportsStatusByID(&v)
	if err != nil {
		res.Code = 500
		res.Message = err.Error()
	} else {
		res.Code = 200
		res.Message = "更新成功"
	}
	res.Data.Items = []models.Topology{}
	c.JSON(http.StatusOK, res)
}

// GetReportHostsGin 获取主机列表
func GetReportHostsGin(c *gin.Context) {
	hostType := c.Query("host_type")
	if hostType == "" {
		hostType = "VM_LIN"
	}
	page := c.Query("page")
	limit := c.Query("limit")
	if page == "" {
		page = "1"
	}
	if limit == "" {
		limit = "1000"
	}

	var res models.HostList
	hosts, count, err := models.HostsList(hostType, page, limit, "", "", "", "")
	if err != nil {
		res.Code = 500
		res.Message = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}
	res.Code = 200
	res.Message = "获取成功"
	res.Data.Items = hosts
	res.Data.Total = count
	c.JSON(http.StatusOK, res)
}

// GetReportItemsGin 获取指标列表
func GetReportItemsGin(c *gin.Context) {
	hostID := c.Query("host_id")
	if hostID == "" {
		var res models.ItemList
		res.Code = 500
		res.Message = "host_id参数不能为空"
		c.JSON(http.StatusOK, res)
		return
	}

	var res models.ItemList
	items, count, err := models.GetAllItemByHostID(hostID)
	if err != nil {
		res.Code = 500
		res.Message = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}
	res.Code = 200
	res.Message = "获取成功"
	res.Data.Items = items
	res.Data.Total = count
	c.JSON(http.StatusOK, res)
}
