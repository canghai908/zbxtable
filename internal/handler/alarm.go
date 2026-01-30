package handler

import (
	"io"
	"net/http"
	"strconv"
	"time"
	"zbxtable/internal/model"
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
	tenant_id := c.Query("tenant_id")
	status := c.Query("status")
	level := c.Query("level")

	var AlarmRes model.AlarmList
	cnt, al, err := model.GetAllAlarm(Begin, End, page, limit, hosts, ip, tenant_id, status, level)
	if err != nil {
		AlarmRes.Code = 200
		AlarmRes.Message = err.Error()
	} else {
		AlarmRes.Code = 200
		AlarmRes.Message = "ok"
		AlarmRes.Data.Items = al
		AlarmRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, AlarmRes)
}

// GetAlarmByID 获取单个告警
func GetAlarmByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	v, err := model.GetAlarmByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, v)
	}
}

// GetAlarmTenant 获取告警租户列表
func GetAlarmTenant(c *gin.Context) {
	var TenantRes model.AlarmTendantList
	cnt, al, err := model.GetAlarmTenant()
	if err != nil {
		TenantRes.Code = 200
		TenantRes.Message = err.Error()
		TenantRes.Data.Items = nil
		TenantRes.Data.Total = 0
	} else {
		TenantRes.Code = 200
		TenantRes.Message = "ok"
		TenantRes.Data.Items = al
		TenantRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, TenantRes)
}

// AnalysisAlarm 告警分析（支持实例筛选）
func AnalysisAlarm(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}

	var v model.ListAnalysisAlarm
	var AnalysisRes model.AnalysisList
	err = jsoniter.Unmarshal(body, &v)
	if err != nil {
		AnalysisRes.Code = 500
		AnalysisRes.Message = err.Error()
		c.JSON(http.StatusOK, AnalysisRes)
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
	instanceID := v.InstanceID
	if instanceID == "" {
		instanceID = v.TenantID // 兼容旧的 tenant_id 字段
	}
	
	arraytitle, piee, na, va, err := model.AnalysisAlarm(Start, End, instanceID)
	if err != nil {
		AnalysisRes.Code = 500
		AnalysisRes.Message = err.Error()
		c.JSON(http.StatusOK, AnalysisRes)
		return
	}
	AnalysisRes.Code = 200
	AnalysisRes.Message = "ok"
	AnalysisRes.Data.Level = arraytitle
	AnalysisRes.Data.LevelCount = piee
	AnalysisRes.Data.Host = na
	AnalysisRes.Data.HostCount = va
	c.JSON(http.StatusOK, AnalysisRes)
}

// ExportAlarm 导出告警（支持实例筛选）
func ExportAlarm(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}

	var v model.ListExportAlarm
	var AlarmRes model.AlarmList
	err = jsoniter.Unmarshal(body, &v)
	if err != nil {
		AlarmRes.Code = 500
		AlarmRes.Message = err.Error()
		c.JSON(http.StatusOK, AlarmRes)
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
	instanceID := v.InstanceID
	if instanceID == "" {
		instanceID = v.TenantID // 兼容旧的 tenant_id 字段
	}
	
	cnt, err := model.ExportAlarm(Start, End, v.Hosts, instanceID, v.Status, v.Level, v.HostIP)
	if err != nil {
		AlarmRes.Code = 200
		AlarmRes.Message = err.Error()
		c.JSON(http.StatusOK, AlarmRes)
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
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Not implemented yet"})
}

// UpdateAlarm 更新告警（占位符）
func UpdateAlarm(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Not implemented yet"})
}

// DeleteAlarm 删除告警（占位符）
func DeleteAlarm(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Not implemented yet"})
}
