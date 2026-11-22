package handler

import (
	"io"
	"strconv"
	"strings"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

func gjsonCSVField(raw gjson.Result) string {
	if raw.IsArray() {
		parts := make([]string, 0, len(raw.Array()))
		for _, item := range raw.Array() {
			s := strings.TrimSpace(item.String())
			if s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ",")
	}
	return strings.TrimSpace(raw.String())
}

func parseInspectionTaskBody(body []byte, id int) *model.InspectionTask {
	res := gjson.ParseBytes(body)
	status := "1"
	statusRes := res.Get("status")
	if statusRes.Exists() {
		switch statusRes.Type {
		case gjson.True:
			status = "1"
		case gjson.False:
			status = "0"
		default:
			if strings.TrimSpace(statusRes.String()) == "0" || strings.EqualFold(strings.TrimSpace(statusRes.String()), "false") {
				status = "0"
			}
		}
	}

	return &model.InspectionTask{
		ID:             id,
		Name:           strings.TrimSpace(res.Get("name").String()),
		ZIDs:           gjsonCSVField(res.Get("z_ids")),
		HostTypes:      gjsonCSVField(res.Get("host_types")),
		GroupIDs:       gjsonCSVField(res.Get("group_ids")),
		InspectMode:    strings.TrimSpace(res.Get("inspect_mode").String()),
		Cycle:          strings.TrimSpace(res.Get("cycle").String()),
		CronExpr:       strings.TrimSpace(res.Get("cron_expr").String()),
		CpuThreshold:   res.Get("cpu_threshold").Float(),
		MemThreshold:   res.Get("mem_threshold").Float(),
		DiskThreshold:  res.Get("disk_threshold").Float(),
		NotifyChannels: gjsonCSVField(res.Get("notify_channels")),
		NotifyUserIDs:  gjsonCSVField(res.Get("notify_user_ids")),
		NotifyGroupIDs: gjsonCSVField(res.Get("notify_group_ids")),
		NotifyEmails:   gjsonCSVField(res.Get("notify_emails")),
		Status:         status,
		Desc:           strings.TrimSpace(res.Get("desc").String()),
	}
}

// GetInspectionOverviewGin 获取巡检中心概览统计
func GetInspectionOverviewGin(c *gin.Context) {
	overview, err := model.GetInspectionOverview()
	if err != nil {
		response.InternalError(c, "获取巡检概览失败: "+err.Error())
		return
	}
	response.Success(c, overview)
}

// ListInspectionTasksGin 获取巡检任务列表
func ListInspectionTasksGin(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	name := c.Query("name")
	status := c.Query("status")

	cnt, list, err := model.ListInspectionTasks(page, limit, name, status)
	if err != nil {
		response.DatabaseError(c, "获取巡检任务失败: "+err.Error())
		return
	}
	response.SuccessWithPage(c, list, cnt)
}

// GetInspectionTaskGin 获取单个巡检任务详情
func GetInspectionTaskGin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.ValidationError(c, "无效的任务 ID")
		return
	}
	task, err := model.GetInspectionTaskByID(id)
	if err != nil {
		response.NotFound(c, "巡检任务不存在")
		return
	}
	response.Success(c, task)
}

// CreateInspectionTaskGin 创建巡检任务
func CreateInspectionTaskGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}
	task := parseInspectionTaskBody(body, 0)
	id, err := model.CreateInspectionTask(task)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", gin.H{"id": id})
}

// UpdateInspectionTaskGin 更新巡检任务
func UpdateInspectionTaskGin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.ValidationError(c, "无效的任务 ID")
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}
	task := parseInspectionTaskBody(body, id)
	if err := model.UpdateInspectionTask(task); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "保存成功", nil)
}

// UpdateInspectionTaskStatusGin 启用或停用巡检任务
func UpdateInspectionTaskStatusGin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.ValidationError(c, "无效的任务 ID")
		return
	}
	body, _ := io.ReadAll(c.Request.Body)
	statusVal := gjson.GetBytes(body, "status")
	statusStr := ""
	if statusVal.Exists() {
		switch statusVal.Type {
		case gjson.True:
			statusStr = "1"
		case gjson.False:
			statusStr = "0"
		default:
			statusStr = strings.TrimSpace(statusVal.String())
		}
	}
	if err := model.UpdateInspectionTaskStatus(id, statusStr); err != nil {
		response.DatabaseError(c, "更新任务状态失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "状态更新成功", nil)
}

// DeleteInspectionTaskGin 删除巡检任务
func DeleteInspectionTaskGin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.ValidationError(c, "无效的任务 ID")
		return
	}
	if err := model.DeleteInspectionTask(id); err != nil {
		response.DatabaseError(c, "删除巡检任务失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// RunInspectionTaskNowGin 立即触发一次巡检
func RunInspectionTaskNowGin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.ValidationError(c, "无效的任务 ID")
		return
	}
	report, err := model.TriggerInspectionTaskNow(id)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "巡检任务已提交", report)
}

// ListInspectionReportsGin 获取巡检报告列表
func ListInspectionReportsGin(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	taskID := c.Query("task_id")
	grade := c.Query("health_grade")
	status := c.Query("status")

	cnt, list, err := model.ListInspectionReports(page, limit, taskID, grade, status)
	if err != nil {
		response.DatabaseError(c, "获取巡检报告列表失败: "+err.Error())
		return
	}
	response.SuccessWithPage(c, list, cnt)
}

// GetInspectionReportDetailGin 获取巡检报告详细评分与主机/告警明细
func GetInspectionReportDetailGin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.ValidationError(c, "无效的报告 ID")
		return
	}
	detail, err := model.GetInspectionReportDetail(id)
	if err != nil {
		response.NotFound(c, "巡检报告不存在")
		return
	}
	response.Success(c, detail)
}

// DeleteInspectionReportGin 删除指定巡检报告
func DeleteInspectionReportGin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		response.ValidationError(c, "无效的报告 ID")
		return
	}
	if err := model.DeleteInspectionReport(id); err != nil {
		response.DatabaseError(c, "删除巡检报告失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
