package handler

import (
	"net/http"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
)

// GetItemByKey 根据hostid及key查找item
func GetItemByKey(c *gin.Context) {
	HostID := c.Query("host_id")
	ItemKey := c.Query("item_key")
	v, err := models.GetItemByKey(HostID, ItemKey)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, v)
	}
}

// GetAllItemByHostID 根据hostid获取主机所有item
func GetAllItemByHostID(c *gin.Context) {
	HostID := c.Query("hostid")
	var ItemRes models.ItemList
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
	c.JSON(http.StatusOK, ItemRes)
}

// GetAllTrafficItem 获取设备所有流量指标
func GetAllTrafficItem(c *gin.Context) {
	HostID := c.Query("hostid")
	var ItemR models.ItemRes
	v, count, err := models.GetAllTrafficItemByHostID(HostID)
	if err != nil {
		ItemR.Code = 500
		ItemR.Message = "获取错误"
	} else {
		ItemR.Code = 200
		ItemR.Message = "获取数据成功"
		ItemR.Data.Items = v
		ItemR.Data.Total = count
	}
	c.JSON(http.StatusOK, ItemR)
}

// GetAllReceiveTrafficItem 获取设备所有出流量
func GetAllReceiveTrafficItem(c *gin.Context) {
	HostID := c.Query("hostid")
	var ItemR models.ItemRes
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
	c.JSON(http.StatusOK, ItemR)
}
