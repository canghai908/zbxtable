package controllers

import (
	"github.com/canghai908/zbxtable/models"
)

//TriggersController funct
type TriggersController struct {
	BaseController
}

//TriggersRes resp
var TriggersRes models.TriggersRes

//URLMapping beego
func (c *TriggersController) URLMapping() {
	c.Mapping("Get", c.GetInfo)
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
