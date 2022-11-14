package controllers

import (
	"strconv"
	"zbxtable/models"
)

type EventLogController struct {
	BaseController
}

var EveLogResp models.EventLogRes

// URLMapping ...
func (c *EventLogController) URLMapping() {
	c.Mapping("GetOne", c.GetOne)
}

// GetOne ...
// @Title Get One
// @Description get eventlog by id
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Alarm
// @Failure 403 :id is empty
// @router /:id [get]
func (c *EventLogController) GetOne() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v, err := models.GetEventLogByAlarmID(id)
	if err != nil {
		EveLogResp.Code = 500
		EveLogResp.Message = "获取失败"
		EveLogResp.Data.Items = nil
		EveLogResp.Data.Total = 0
	} else {
		EveLogResp.Code = 200
		EveLogResp.Message = "获取成功"
		EveLogResp.Data.Items = v
		EveLogResp.Data.Total = int64(len(v))
		c.Data["json"] = v
	}
	c.Data["json"] = EveLogResp
	c.ServeJSON()
}
