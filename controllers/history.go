package controllers

import (
	"time"
	"zbxtable/models"
)

// HistoryController 监控详情数据获取
type HistoryController struct {
	BaseController
}

//HistoryRes is used
var HistoryRes models.HistoryList

// URLMapping ...
func (c *HistoryController) URLMapping() {
	c.Mapping("GetHistoryByItemID", c.GetHistoryByItemID)
}

// GetHistoryByItemID controller
// @Title 监控详情数据获取
// @Description 根据ItemID、item类型、开始、结束时间获取监控数据
// @Param	X-Token	header  string				true	"x-token in header"
// @Param	body	body 	models.HistoryQuery	true    "查询"
// @Success 200 {object} models.History
// @Failure 403 :id is empty
// @router / [post]
func (c *HistoryController) GetHistoryByItemID() {
	var v models.HistoryQuery
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	if err != nil {
		HistoryRes.Code = 500
		HistoryRes.Message = err.Error()
		c.Data["json"] = HistoryRes
		c.ServeJSON()
		return
	}
	var Start, End int64
	//如果时间为空，默认为10分钟
	if len(v.Period) == 0 || v.Period[0] == "" || v.Period[1] == "" {
		tEnd := time.Now()
		End = tEnd.Unix()
		Start = tEnd.Add(-10 * time.Minute).Unix()
	}
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	st, _ := time.ParseInLocation(timeLayout, v.Period[0], loc)
	en, _ := time.ParseInLocation(timeLayout, v.Period[1], loc)
	Start = st.Unix()
	End = en.Unix()
	his, err := models.GetHistoryByItemID(v.Itemids, v.History, Start, End)
	if err != nil {
		HistoryRes.Code = 500
		HistoryRes.Message = err.Error()
		c.Data["json"] = HistoryRes
		c.ServeJSON()
		return
	}
	HistoryRes.Code = 200
	HistoryRes.Message = "获取成功"
	HistoryRes.Data.Items = his
	c.Data["json"] = HistoryRes
	c.ServeJSON()
	return
}
