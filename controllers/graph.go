package controllers

import (
	"strconv"
	"time"

	"zbxtable/models"
)

// GraphController operations for Host
type GraphController struct {
	BaseController
}

//GraphRes restp
var GraphRes models.GraphList

// URLMapping ...
func (c *GraphController) URLMapping() {
	//c.Mapping("Post", c.Post)
	c.Mapping("Post", c.Post)
	c.Mapping("Exp", c.Exp)

}

// Post t
// @Title 根据Hostid查看主机图形
// @Description get graphs by hostid
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	body		body 	models.GraphListQuery	true		"body for Host content"
// @Success 200 {object} models.GraphInfo
// @Failure 403 :id is empty
// @router /:hostid [post]
func (c *GraphController) Post() {
	var v models.GraphListQuery
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err != nil {
		GraphRes.Code = 200
		GraphRes.Message = err.Error()
		c.Data["json"] = GraphRes
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

	id, _ := strconv.Atoi(v.Hostid)
	vv, count, err := models.GetGraphByHostID(id, Start, End)
	if err != nil {
		GraphRes.Code = 500
		GraphRes.Message = "获取图形数据错误"
		c.Data["json"] = GraphRes
		c.ServeJSON()
		return
	}
	GraphRes.Code = 200
	GraphRes.Message = "获取数据成功"
	GraphRes.Data.Items = vv
	GraphRes.Data.Total = count
	c.Data["json"] = GraphRes
	c.ServeJSON()
}

// Exp get
// @Title 导主机组或主机图形为PDF
// @Description get graphs by hostid
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	body		body 	models.GraphExpQuery	true		"body for Host content"
// @Success 200 {object} models.GraphInfo
// @Failure 403 :id is empty
// @router /exp [post]
func (c *GraphController) Exp() {
	var v models.GraphExpQuery
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err != nil {
		GraphRes.Code = 200
		GraphRes.Message = err.Error()
		c.Data["json"] = GraphRes
		c.ServeJSON()
		return
	}

	var Start, End string

	//如果时间为空，默认为一周
	if len(v.Period) == 0 || v.Period[0] == "" || v.Period[1] == "" {
		tEnd := time.Now()
		End = tEnd.Format("2006-01-02 15:04:05")
		Start = tEnd.Add(-168 * time.Hour).Format("2006-01-02 15:04:05")
	}

	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	st, _ := time.ParseInLocation(timeLayout, v.Period[0], loc)
	en, _ := time.ParseInLocation(timeLayout, v.Period[1], loc)
	Start = st.Format("2006-01-02 15:04:05")
	End = en.Format("2006-01-02 15:04:05")
	vvv, err := models.SaveImagePDF(v.Hostids, Start, End)
	// vv, err := ch
	if err != nil {
		GraphRes.Code = 500
		GraphRes.Message = "获取图形数据错误"
		c.Data["json"] = GraphRes
		c.ServeJSON()
		return
	}
	c.Ctx.Output.Header("Content-Type", "application/octet-stream")
	c.Ctx.Output.Header("Content-Disposition", "attachment; filename=graph_export.pdf")
	c.Ctx.Output.Header("Content-Transfer-Encoding", "binary")
	c.Ctx.Output.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Ctx.Output.Status = 200
	c.Ctx.Output.EnableGzip = true
	c.Ctx.Output.Context.Output.Body(vvv)
	return
}
