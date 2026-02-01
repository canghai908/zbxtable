package handler

import (
	"io"
	"net/http"
	"strconv"
	"time"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"
	"zbxtable/pkg/utils"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

// GetAllAlarm 获取告警列表
func GetAllAlarm(c *gin.Context) {
	var Begin, End time.Time
	var err error
	begin := c.Query("begin")
	end := c.Query("end")
	if begin == "" || end == "" {
		End = time.Now()
		Begin = End.Add(-168 * time.Hour)
	}
	Begin, err = utils.ParseTime(begin)
	if err != nil {
		Begin = time.Now().Add(-168 * time.Hour)
	}
	End, err = utils.ParseTime(end)
	if err != nil {
		End = time.Now()
	}
	page := c.Query("page")
	limit := c.Query("limit")
	hosts := c.Query("hosts")
	ip := c.Query("host_ip")
	tenant_id := c.Query("zid")
	status := c.Query("status")
	level := c.Query("level")

	cnt, al, err := model.GetAllAlarm(Begin, End, page, limit, hosts, ip, tenant_id, status, level)
	if err != nil {
		response.DatabaseError(c, "获取告警列表失败: "+err.Error())
		return
	}

	response.SuccessWithPage(c, al, cnt)
}

// GetAlarmByID 获取单个告警
func GetAlarmByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	v, err := model.GetAlarmByID(id)
	if err != nil {
		response.NotFound(c, "告警不存在")
		return
	}

	response.Success(c, v)
}

// GetAlarmTenant 获取告警租户列表
func GetAlarmTenant(c *gin.Context) {
	cnt, al, err := model.GetAlarmTenant()
	if err != nil {
		response.DatabaseError(c, "获取租户列表失败: "+err.Error())
		return
	}

	response.SuccessWithPage(c, al, cnt)
}

// AnalysisAlarm 告警分析（支持实例筛选）
func AnalysisAlarm(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	var v model.ListAnalysisAlarm
	err = jsoniter.Unmarshal(body, &v)
	if err != nil {
		response.BadRequest(c, "请求参数解析失败: "+err.Error())
		return
	}

	var Start, End time.Time
	if v.Begin == "" || v.End == "" {
		End := time.Now()
		Start = End.Add(-168 * time.Hour)
	}
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	Start, _ = time.ParseInLocation(timeLayout, v.Begin, loc)
	End, _ = time.ParseInLocation(timeLayout, v.End, loc)

	// 支持 instance_id 参数
	zid := v.ZID
	if zid == "" {
		zid = v.ZID // 兼容旧的 tenant_id 字段
	}

	arraytitle, piee, na, va, err := model.AnalysisAlarm(Start, End, zid)
	if err != nil {
		response.DatabaseError(c, "告警分析失败: "+err.Error())
		return
	}

	response.Success(c, map[string]interface{}{
		"level":       arraytitle,
		"level_count": piee,
		"host":        na,
		"host_count":  va,
	})
}

// ExportAlarm 导出告警（支持实例筛选）
func ExportAlarm(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	var v model.ListExportAlarm
	err = jsoniter.Unmarshal(body, &v)
	if err != nil {
		response.BadRequest(c, "请求参数解析失败: "+err.Error())
		return
	}

	var Start, End time.Time
	if v.Begin == "" || v.End == "" {
		End := time.Now()
		Start = End.Add(-168 * time.Hour)
	}
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	Start, _ = time.ParseInLocation(timeLayout, v.Begin, loc)
	End, _ = time.ParseInLocation(timeLayout, v.End, loc)

	// 支持 instance_id 参数
	zid := v.ZID
	if zid == "" {
		zid = v.ZID // 兼容旧的 tenant_id 字段
	}

	cnt, err := model.ExportAlarm(Start, End, v.Hosts, zid, v.Status, v.Level, v.HostIP)
	if err != nil {
		response.InternalError(c, "导出告警失败: "+err.Error())
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=alarm_list.xlsx")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Data(http.StatusOK, "application/octet-stream", cnt)
}

// CreateAlarm 创建告警（占位符）
func CreateAlarm(c *gin.Context) {
	response.SuccessWithMessage(c, "Not implemented yet", nil)
}

// UpdateAlarm 更新告警（占位符）
func UpdateAlarm(c *gin.Context) {
	response.SuccessWithMessage(c, "Not implemented yet", nil)
}

// DeleteAlarm 删除告警（占位符）
func DeleteAlarm(c *gin.Context) {
	response.SuccessWithMessage(c, "Not implemented yet", nil)
}
