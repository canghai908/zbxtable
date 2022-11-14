package controllers

import (
	"strconv"
	"time"
	"zbxtable/utils"

	"zbxtable/models"
)

// AlarmController 历史告警消息接口
type AlarmController struct {
	BaseController
}

// AlarmRes used
var AlarmRes models.AlarmList

// AnalysisRes mod
var AnalysisRes models.AnalysisList

// URLMapping ...
func (c *AlarmController) URLMapping() {
	c.Mapping("GetOne", c.GetOne)
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("GetTenant", c.GetTenant)
	c.Mapping("Analysis", c.Analysis)
	c.Mapping("Export", c.Export)
}

// GetOne ...
// @Title Get One
// @Description get Alarm by id
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Alarm
// @Failure 403 :id is empty
// @router /:id [get]
func (c *AlarmController) GetOne() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v, err := models.GetAlarmByID(id)
	if err != nil {
		c.Data["json"] = err.Error()
	} else {
		c.Data["json"] = v
	}
	c.ServeJSON()
}

// GetAll ...
// @Title 查询或获取告警接口
// @Description get Alarm
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	begin	query	string	false	"开始日期 格式 2006-01-02 15:04:05"
// @Param	end		query	string	false	"结束日期 格式 2006-01-02 15:04:05"
// @Param	page	query	string	false	"第几页"
// @Param	limit	query	string	false	"每页条数"
// @Param	hosts	query	string	false	"查询主机名包含某字符的主机"
// @Param	tenant_id	query	string	false	"租户id"
// @Success 200 {object} models.Alarm
// @Failure 403
// @router / [get]
func (c *AlarmController) GetAll() {
	var Begin, End time.Time
	var err error
	begin := c.Ctx.Input.Query("begin")
	end := c.Ctx.Input.Query("end")
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
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	hosts := c.Ctx.Input.Query("hosts")
	tenant_id := c.Ctx.Input.Query("tenant_id")
	status := c.Ctx.Input.Query("status")
	level := c.Ctx.Input.Query("level")
	cnt, al, err := models.GetAllAlarm(Begin, End, page, limit, hosts, tenant_id, status, level)
	if err != nil {
		AlarmRes.Code = 200
		AlarmRes.Message = err.Error()
	} else {
		AlarmRes.Code = 200
		AlarmRes.Message = "ok"
		AlarmRes.Data.Items = al
		AlarmRes.Data.Total = cnt
	}
	c.Data["json"] = AlarmRes
	c.ServeJSON()
}

// GetAll ...
// @Title 查询或获取告警接口
// @Description get Alarm
// @Param	X-Token		header  string			true		"x-token in header"
// @Success 200 {object} models.Alarm
// @Failure 403
// @router /tenant [get]
func (c *AlarmController) GetTenant() {
	var TenantRes models.AlarmTendantList
	cnt, al, err := models.GetAlarmTenant()
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
	c.Data["json"] = TenantRes
	c.ServeJSON()
}

// Analysis ...
// @Title 告警分析接口
// @Description 告警分析接口
// @Param	X-Token header  string			true		"x-token in header"
// @Param	body   body  models.ListAnalysisAlarm true "分析周期"
// @Success 200 {object} models.AnalysisRes
// @Failure 403
// @router /analysis [post]
func (c *AlarmController) Analysis() {
	var v models.ListAnalysisAlarm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err != nil {
		AnalysisRes.Code = 500
		AnalysisRes.Message = err.Error()
		c.Data["json"] = AnalysisRes
		c.ServeJSON()
		return
	}
	var Start, End time.Time
	//如果时间为空，默认为一周
	if v.Begin == "" || v.End == "" {
		End := time.Now()
		Start = End.Add(-168 * time.Hour)
	}
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	Start, _ = time.ParseInLocation(timeLayout, v.Begin, loc)
	End, _ = time.ParseInLocation(timeLayout, v.End, loc)
	arraytitle, piee, na, va, err := models.AnalysisAlarm(Start, End, v.TenantID)
	if err != nil {
		AnalysisRes.Code = 500
		AnalysisRes.Message = err.Error()
		c.Data["json"] = AnalysisRes
		c.ServeJSON()
		return
	}
	AnalysisRes.Code = 200
	AnalysisRes.Message = "ok"
	AnalysisRes.Data.Level = arraytitle
	AnalysisRes.Data.LevelCount = piee
	AnalysisRes.Data.Host = na
	AnalysisRes.Data.HostCount = va
	c.Data["json"] = AnalysisRes
	c.ServeJSON()
}

// Export export
// @Title 导出告警消息
// @Description 根据查询的条件导出告警报表到xlsx文件
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	body		body 	models.ListExportAlarm	true "导出条件或周期""
// @Success 200 {object} models.Group
// @Failure 403  is empty
// @router /export [post]
func (c *AlarmController) Export() {
	var v models.ListExportAlarm
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err != nil {
		AlarmRes.Code = 500
		AlarmRes.Message = err.Error()
		c.Data["json"] = AlarmRes
		c.ServeJSON()
		return
	}
	var Start, End time.Time
	//如果时间为空，默认为一周
	if v.Begin == "" || v.End == "" {
		End := time.Now()
		Start = End.Add(-168 * time.Hour)
	}
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	Start, _ = time.ParseInLocation(timeLayout, v.Begin, loc)
	End, _ = time.ParseInLocation(timeLayout, v.End, loc)
	cnt, err := models.ExportAlarm(Start, End, v.Hosts, v.TenantID, v.Status, v.Level)
	if err != nil {
		AlarmRes.Code = 200
		AlarmRes.Message = err.Error()
		c.Data["json"] = AlarmRes
		c.ServeJSON()
		return
	}
	c.Ctx.Output.Header("Content-Type", "application/octet-stream")
	c.Ctx.Output.Header("Content-Disposition", "attachment; filename=alarm_list.xlsx")
	c.Ctx.Output.Header("Content-Transfer-Encoding", "binary")
	c.Ctx.Output.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Ctx.Output.Status = 200
	c.Ctx.Output.EnableGzip = true
	c.Ctx.Output.Context.Output.Body(cnt)
	return
}
