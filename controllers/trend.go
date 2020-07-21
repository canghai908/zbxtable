package controllers

import (
	"github.com/canghai908/zbxtable/models"
)

// TrendController operations for Trend
type TrendController struct {
	BaseController
}

//TrendRes resp
var TrendRes models.TrendList

// URLMapping ...
func (c *TrendController) URLMapping() {
	c.Mapping("GetTrendByItemID", c.GetTrendByItemID)
}

// GetTrendByItemID controller
// @Title Get One
// @Description get item by key
// @Param	X-Token		header  string	true	"x-token in header"
// @Param	item_id		query 	string	true	"The key for item"
// @Param	limit		query 	string	true	"The key for item"
// @Success 200 {object} models.Item
// @Failure 403 :id is empty
// @router / [get]
func (c *TrendController) GetTrendByItemID() {
	itemID := c.Ctx.Input.Query("item_id")
	limit := c.Ctx.Input.Query("limit")
	v, err := models.GetTrendByItemID(itemID, limit)
	if err != nil {
		c.Data["json"] = err.Error()
	} else {
		c.Data["json"] = v
	}
	c.ServeJSON()
}
