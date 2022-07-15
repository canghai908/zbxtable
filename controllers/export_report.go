package controllers

import (
	"github.com/astaxie/beego/logs"
	"net/url"
	"strconv"
	"time"

	"zbxtable/models"
)

// ExpController operations for Group
type ExpController struct {
	BaseController
}

// ExpRes is used
//var GroupRes models.GroupList
//var ExpRes list
var ExpRes models.ExpList

// URLMapping ...
func (c *ExpController) URLMapping() {
	c.Mapping("GetHostInfo", c.GetHostList)
	c.Mapping("GetItemTrend", c.GetItemTrend)
	c.Mapping("GetItemHistory", c.GetItemHistory)
	c.Mapping("Inspect", c.Inspect)
}

// GetItemTrend export
// @Title Item趋势数据导出
// @Description 根据ItemID导出趋势数据为xlsx文件
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	body		body 	models.ListQueryAll	true		"body for Host content"
// @Success 200 {object} models.Group
// @Failure 403  is empty
// @router /trend [post]
func (c *ExpController) GetItemTrend() {
	var v models.ListQueryAll
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err != nil {
		ExpRes.Code = 500
		ExpRes.Message = err.Error()
		c.Data["json"] = ExpRes
		c.ServeJSON()
		return
	}

	var Start, End int64

	//如果时间为空，默认为一周
	if len(v.Period) == 0 || v.Period[0] == "" || v.Period[1] == "" {
		tEnd := time.Now()
		End = tEnd.Unix()
		Start = tEnd.Add(-168 * time.Hour).Unix()
	}
	//时间格式化
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	st, _ := time.ParseInLocation(timeLayout, v.Period[0], loc)
	en, _ := time.ParseInLocation(timeLayout, v.Period[1], loc)
	Start = st.Unix()
	End = en.Unix()

	iodata, err := models.GetTrenDataFileName(v, Start, End)
	if err != nil {
		ExpRes.Code = 200
		ExpRes.Message = err.Error()
		c.Data["json"] = ExpRes
		c.ServeJSON()
		return
	}
	oldfilename := v.Host.Name + "_" + v.Item.Name + "_trend" + ".xlsx"
	filename := url.QueryEscape(oldfilename)
	c.Ctx.Output.Header("Content-Type", "application/octet-stream")
	c.Ctx.Output.Header("Content-Disposition", "attachment;filename="+filename)
	c.Ctx.Output.Header("Content-Transfer-Encoding", "binary")
	c.Ctx.Output.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Ctx.Output.Status = 200
	c.Ctx.Output.EnableGzip = true
	c.Ctx.Output.Context.Output.Body(iodata)
	return
}

// GetItemHistory export
// @Title 数据导出
// @Description 根据ItemID导出详情数据为xlsx文件
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	body		body 	models.ListQueryAll	true    "body for Host content"
// @Success 200 {object} models.Group
// @Failure 403  is empty
// @router /history [post]
func (c *ExpController) GetItemHistory() {
	var v models.ListQueryAll
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err != nil {
		ExpRes.Code = 500
		ExpRes.Message = err.Error()
		c.Data["json"] = ExpRes
		c.ServeJSON()
		return
	}
	var Start, End int64

	//如果时间为空，默认为一周
	if len(v.Period) == 0 || v.Period[0] == "" || v.Period[1] == "" {
		tEnd := time.Now()
		End = tEnd.Unix()
		Start = tEnd.Add(-168 * time.Hour).Unix()
	}

	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	st, _ := time.ParseInLocation(timeLayout, v.Period[0], loc)
	en, _ := time.ParseInLocation(timeLayout, v.Period[1], loc)
	Start = st.Unix()
	End = en.Unix()

	iodata, err := models.GetHistoryDataFileName(v, Start, End)
	if err != nil {
		ExpRes.Code = 200
		ExpRes.Message = err.Error()
		c.Data["json"] = ExpRes
		c.ServeJSON()
		return
	}
	oldfilename := v.Host.Name + "_" + v.Item.Name + "_history" + ".xlsx"
	filename := url.QueryEscape(oldfilename)
	c.Ctx.Output.Header("Content-Type", "application/octet-stream")
	c.Ctx.Output.Header("Content-Disposition", "attachment; filename="+filename)
	c.Ctx.Output.Header("Content-Transfer-Encoding", "binary")
	c.Ctx.Output.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Ctx.Output.Status = 200
	c.Ctx.Output.EnableGzip = true
	c.Ctx.Output.Context.Output.Body(iodata)
	return
}

// Inspect ITm
// @Title 巡检报告导出
// @Description 按主机组导出巡检表
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	body		body 	models.ListQueryNew	true		"body for Host content"
// @Success 200 {object} models.Group
// @Failure 403  is empty
// @router /inspect [post]
func (c *ExpController) Inspect() {
	var v models.HostGroupsPlist
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err != nil {
		ExpRes.Code = 200
		ExpRes.Message = err.Error()
		c.Data["json"] = ExpRes
		c.ServeJSON()
		return
	}
	hostdata, err := models.GetHostsByGroupIDList(v.GroupID)
	if err != nil {
		ExpRes.Code = 200
		ExpRes.Message = err.Error()
		c.Data["json"] = ExpRes
		c.ServeJSON()
	}
	type ss struct {
		V1 float64 `json:"v1"`
		V2 float64 `json:"v2"`
		V3 float64 `json:"v3"`
		V4 float64 `json:"v4"`
	}
	tt := make([]ss, len(hostdata))
	yy := make([]models.Insp, len(hostdata))
	for kk, v := range hostdata {
		b, err := models.GetItemByKey(v.HostID, "system.cpu.util[,idle]")
		if err != nil {
			logs.Error(err)
			ExpRes.Code = 200
			ExpRes.Message = err.Error()
			c.Data["json"] = ExpRes
			c.ServeJSON()
		}

		mem1, err := models.GetItemByKey(v.HostID, "vm.memory.size[total]")

		if err != nil {
			logs.Error(err)
			ExpRes.Code = 200
			ExpRes.Message = err.Error()
			c.Data["json"] = ExpRes
			c.ServeJSON()
		}
		mem2, err := models.GetItemByKey(v.HostID, "vm.memory.size[available]")
		if err != nil {
			logs.Error(err)
			ExpRes.Code = 200
			ExpRes.Message = err.Error()
			c.Data["json"] = ExpRes
			c.ServeJSON()
		}
		for _, v := range mem1 {

			vint64, _ := strconv.ParseFloat(v.Lastvalue, 64)
			tt[kk].V1 = vint64
		}
		for _, v := range mem2 {
			vint64, _ := strconv.ParseFloat(v.Lastvalue, 64)
			tt[kk].V2 = vint64
		}
		for _, v := range b {
			vint64, _ := strconv.ParseFloat(v.Lastvalue, 64)
			if vint64 != 0 {
				tt[kk].V3 = models.Round(100-vint64, 2)
			} else {
				tt[kk].V3 = 0
			}
			if tt[kk].V1 != 0 {
				tt[kk].V4 = models.Round(tt[kk].V2/tt[kk].V1, 2)
			} else {
				tt[kk].V4 = 0
			}
		}
		yy[kk].HostName = v.Name
		yy[kk].CPULoad = tt[kk].V3
		yy[kk].MemPct = tt[kk].V4
	}
	ByteData, err := models.ExpInspect(v.Name, yy)
	if err != nil {
		ExpRes.Code = 200
		ExpRes.Message = err.Error()
		c.Data["json"] = ExpRes
		c.ServeJSON()
		return
	}
	oldfilename := v.Name + ".xlsx"
	filename := url.QueryEscape(oldfilename)
	c.Ctx.Output.Header("Content-Type", "application/octet-stream")
	c.Ctx.Output.Header("Content-Disposition", "attachment; filename="+filename)
	c.Ctx.Output.Header("Content-Transfer-Encoding", "binary")
	c.Ctx.Output.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Ctx.Output.Status = 200
	c.Ctx.Output.EnableGzip = true
	c.Ctx.Output.Context.Output.Body(ByteData)
	return
}

// GetHostInfo ...
// @Title 导出设备列表
// @Description 根据设备类型导出设备列表到excel
// @Param	X-Token		header  string	true	"X-Token"
// @Param	body		body 	models.ExportHosts	true "导出条件或周期""
// @Success 200 {object} models.Alarm
// @Failure 403
// @router /hosts [post]
func (c *ExpController) GetHostList() {
	var v models.ExportHosts
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err != nil {
		HostRes.Code = 500
		HostRes.Message = err.Error()
		c.Data["json"] = HostRes
		c.ServeJSON()
		return
	}
	hs, err := models.GetHostList(v.Hosttype, v.Hosts, v.Model, v.Ip, v.Available)
	if err != nil {
		HostRes.Code = 500
		HostRes.Message = err.Error()
		c.Data["json"] = HostRes
		c.ServeJSON()
		return
	}
	//oldfilename := "host_list.xlsx"
	//filename := url.QueryEscape(oldfilename)
	c.Ctx.Output.Header("Content-Type", "application/octet-stream")
	c.Ctx.Output.Header("Content-Disposition", "attachment; filename=host_list.xlsx")
	c.Ctx.Output.Header("Content-Transfer-Encoding", "binary")
	c.Ctx.Output.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Ctx.Output.Status = 200
	c.Ctx.Output.EnableGzip = true
	c.Ctx.Output.Context.Output.Body(hs)
	return
}
