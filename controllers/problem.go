package controllers

import (
	"github.com/canghai908/zbxtable/models"
)

//ProblemsController funct
type ProblemsController struct {
	BaseController
}

//ProblemsRes resp
var ProblemsRes models.ProblemsRes

//URLMapping beego
func (c *ProblemsController) URLMapping() {
	c.Mapping("Get", c.GetInfo)
}

// GetInfo 获取未恢复告警
// @Title 获取未恢复告警据
// @Description 获取未恢复告警
// @Param	X-Token		header  string			true		"x-token in header"
// @Success 200 {object} models.Manager
// @Failure 403 :id is empty
// @router / [get]
func (c *ProblemsController) GetInfo() {
	b, cnt, err := models.GetProblems()
	if err != nil {
		ProblemsRes.Code = 500
		ProblemsRes.Message = err.Error()
	} else {
		ProblemsRes.Code = 200
		ProblemsRes.Message = "获取成功"
		ProblemsRes.Data.Items = b
		ProblemsRes.Data.Total = cnt
	}
	c.Data["json"] = ProblemsRes
	c.ServeJSON()
}
