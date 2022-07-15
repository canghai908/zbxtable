package controllers

import (
	"zbxtable/models"
)

// 主机监控指标获取
type ItemController struct {
	BaseController
}

//ItemRes rep
var ItemRes models.ItemList
var ItemR models.ItemRes

// URLMapping ...
func (c *ItemController) URLMapping() {
	c.Mapping("GetItemByKey", c.GetItemByKey)
	//c.Mapping("GetFlowItem", c.GetFlowItem)
	c.Mapping("GetAllItemByKey", c.GetAllItemByKey)
	c.Mapping("GetAllTraffficByKey", c.GetAllTraffficByKey)
	c.Mapping("GetAllTraffficReceive", c.GetAllTraffficReceive)
}

// GetItemByKey controller
// @Title 根据hostid及key查找item
// @Description 根据hostid及key查找item
// @Param	X-Token		header  string	true		"x-token in header"
// @Param	hostid		query 	string	true		"The key for item"
// @Param	ItemKey	query 	string	true		"The key for item"
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

// 根据hostid获取主机所有item
// @Title 获取主机所有item
// @Description 根据hostid获取item列表
// @Param	X-Token		header  string	true		"x-token in header"
// @Param	hostid		query 	string	true		"hostid"
// @Success 200 {object} models.Item
// @Failure 403
// @router /list [get]
func (c *ItemController) GetAllItemByKey() {
	HostID := c.Ctx.Input.Query("hostid")
	v, count, err := models.GetAllItemByHostID(HostID)
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

// 根据hostid获取设备所有流量指标
// @Title 获取设备流量指标Item
// @Description 获取设备流量指标Item
// @Param	X-Token		header  string	true		"x-token in header"
// @Param	hostid		query 	string	true		"hostid"
// @Success 200 {object} models.Item
// @Failure 403
// @router /traffic [get]
func (c *ItemController) GetAllTraffficByKey() {
	HostID := c.Ctx.Input.Query("hostid")
	v, count, err := models.GetAllTrafficeItemByHostID(HostID)
	if err != nil {
		ItemR.Code = 500
		ItemR.Message = "获取错误"
	} else {
		ItemR.Code = 200
		ItemR.Message = "获取数据成功"
		ItemR.Data.Items = v
		ItemR.Data.Total = count
	}
	c.Data["json"] = ItemR
	c.ServeJSON()
}

// 根据hostid获取设备所有出流量
// @Title 获取设备流量指标Item
// @Description 获取设备流量指标Item
// @Param	X-Token		header  string	true		"x-token in header"
// @Param	hostid		query 	string	true		"hostid"
// @Success 200 {object} models.Item
// @Failure 403
// @router /topotraffic [get]
func (c *ItemController) GetAllTraffficReceive() {
	HostID := c.Ctx.Input.Query("hostid")
	v, count, err := models.GetReceiveTrafficeItemByHostID(HostID)
	if err != nil {
		ItemR.Code = 500
		ItemR.Message = "获取错误"
	} else {
		ItemR.Code = 200
		ItemR.Message = "获取数据成功"
		ItemR.Data.Items = v
		ItemR.Data.Total = count
	}
	c.Data["json"] = ItemR
	c.ServeJSON()
}

//// 根据hostid获取主机所有item
//// @Title 获取主机所有item
//// @Description 根据hostid获取item列表
//// @Param	X-Token		header  string	true		"x-token in header"
//// @Param	hostid		query 	string	true		"hostid"
//// @Success 200 {object} models.Item
//// @Failure 403
//// @router /list [get]
//func (c *ItemController) GetFlowItem() {
//	HostID := c.Ctx.Input.Query("hostid")
//	v, count, err := models.GetFlowItemByHostID(HostID)
//	if err != nil {
//		ItemRes.Code = 500
//		ItemRes.Message = "获取错误"
//	} else {
//		ItemRes.Code = 200
//		ItemRes.Message = "获取数据成功"
//		ItemRes.Data.Items = v
//		ItemRes.Data.Total = count
//	}
//	c.Data["json"] = ItemRes
//	c.ServeJSON()
//}
