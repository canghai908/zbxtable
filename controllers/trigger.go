package controllers

import (
	"zbxtable/models"
)

//TriggersController funct
type TriggersController struct {
	BaseController
}

//TriggersRes resp
var TriggersRes models.TriggersRes
var TriggersListRes models.TriggersListRes

//URLMapping beego
func (c *TriggersController) URLMapping() {
	c.Mapping("Get", c.GetInfo)
	c.Mapping("GetOne", c.GetOne)
}

// GetInfo 获取未恢复告警
// @Title 获取未恢复告警据
// @Description 获取未恢复告警
// @Param	X-Token		header  string			true		"x-token in header"
// @Success 200 {object} models.Triggers
// @Failure 403 :id is empty
// @router / [get]
func (c *TriggersController) GetInfo() {
	b, cnt, err := models.GetTriggers()
	if err != nil {
		TriggersRes.Code = 500
		TriggersRes.Message = err.Error()
	} else {
		TriggersRes.Code = 200
		TriggersRes.Message = "获取成功"
		TriggersRes.Data.Items = b
		TriggersRes.Data.Total = cnt
	}
	c.Data["json"] = TriggersRes
	c.ServeJSON()
}

// GetOne 根据主机ID获取Trigger列表
// @Title 根据主机ID获取Trigger列表,
// @Description 根据主机ID获取Trigger列表
// @Param	X-Token		header  string	true		"x-token in header"\
// @Param	hostid		query 	string	true		"hostid"
// @Success 200 {object} models.Triggers
// @Failure 403 :id is empty
// @router /list [get]
func (c *TriggersController) GetOne() {
	HostID := c.Ctx.Input.Query("hostid")
	b, cnt, err := models.GetTriggerList(HostID)
	if err != nil {
		TriggersListRes.Code = 500
		TriggersListRes.Message = err.Error()
	} else {
		TriggersListRes.Code = 200
		TriggersListRes.Message = "获取成功"
		TriggersListRes.Data.Items = b
		TriggersListRes.Data.Total = cnt
	}
	c.Data["json"] = TriggersListRes
	c.ServeJSON()
}
