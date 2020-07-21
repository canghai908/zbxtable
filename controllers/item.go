package controllers

import (
	"encoding/json"

	"github.com/canghai908/zbxtable/models"
)

// ItemController operations for Item
type ItemController struct {
	BaseController
}

//ItemRes rep
var ItemRes models.ItemList

// URLMapping ...
func (c *ItemController) URLMapping() {
	c.Mapping("GetItemByKey", c.GetItemByKey)
	c.Mapping("GetAllItemByKey", c.GetAllItemByKey)

}

// GetItemByKey controller
// @Title Get One
// @Description get item by key
// @Param	X-Token		header  string	true		"x-token in header"
// @Param	host_id		query 	string	true		"The key for item"
// @Param	item_key	query 	string	true		"The key for item"
// @Success 200 {object} models.Item
// @Failure 403 :id is empty
// @router / [get]
func (c *ItemController) GetItemByKey() {
	HostID := c.Ctx.Input.Query("host_id")
	ItemKey := c.Ctx.Input.Query("item_key")
	// id, _ := strconv.Atoi(idStr)
	v, err := models.GetItemByKey(HostID, ItemKey)
	if err != nil {
		c.Data["json"] = err.Error()
	} else {
		c.Data["json"] = v
	}
	c.ServeJSON()
}

// GetAllItemByKey controller
// @Title Get ALl item
// @Description get ALl item by key
// @Param	X-Token		header  string	true		"x-token in header"
// @Param	id		    path 	string	true		"The host for id"
// @Success 200 {object} models.Item
// @Failure 403
// @router /list [post]
func (c *ItemController) GetAllItemByKey() {

	var p models.Hosts
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &p)
	if err != nil {
		c.Data["json"] = err.Error()
		c.ServeJSON()
		return
	}
	v, count, err := models.GetAllItemByHostID(p.HostID)
	if err != nil {
		ItemRes.Code = 500
		ItemRes.Message = "获取错误"
	} else {
		ItemRes.Code = 200
		ItemRes.Message = "获取数据成功"
		ItemRes.Data.Items = v
		ItemRes.Data.Total = count
	}
	c.Data["json"] = ItemRes
	c.ServeJSON()
}

// GetAll ...
// @Title Get All Hosts
// @Description get hosts
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	begin	query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
// @Param	end		query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	page	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	limit	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	hosts	query	string	false	"Limit the size of result set. Must be an integer"
// @Success 200 {object} models.Alarm
// @Failure 403
// @router /:itemid [get]
// func (c *ItemController) GetAll() {
// 	hs, count, err := models.HostsList()
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	HostRes.Code = 200
// 	HostRes.Message = "获取数据成功"
// 	HostRes.Data.Items = hs
// 	HostRes.Data.Total = count
// 	// fmt.Println(len(hs))
// 	//fmt.Println(hs)
// 	c.Data["json"] = HostRes
// 	c.ServeJSON()

// }
