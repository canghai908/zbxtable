package controllers

import (
	"github.com/canghai908/zbxtable/models"
)

// HistoryController operations for History
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
// @Title Get One
// @Description get item by key
// @Param	X-Token		header  string	true		"x-token in header"
// @Param	item_id		query 	string	true		"The key for item"
// @Param	history 	query 	string	true		"history type"
// @Param	limit	 	query 	int		true		"The key for limit"
// @Success 200 {object} models.Item
// @Failure 403 :id is empty
// @router / [get]
func (c *HistoryController) GetHistoryByItemID() {
	itemID := c.Ctx.Input.Query("item_id")
	history := c.Ctx.Input.Query("history")
	limit := c.Ctx.Input.Query("limit")
	// id, _ := strconv.Atoi(idStr)
	v, err := models.GetHistoryByItemID(itemID, history, limit)
	if err != nil {
		c.Data["json"] = err.Error()
	} else {
		c.Data["json"] = v
	}
	c.ServeJSON()
}
