package handler

import (
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetItemByKey 根据hostid及key查找item
func GetItemByKey(c *gin.Context) {
	HostID := c.Query("host_id")
	ItemKey := c.Query("item_key")
	v, err := model.GetItemByKey(HostID, ItemKey)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, v)
}

// GetAllItemByHostID 根据hostid获取主机所有item
func GetAllItemByHostID(c *gin.Context) {
	HostID := c.Query("hostid")
	instanceID := c.Query("instance_id")
	
	var v interface{}
	var count int64
	var err error
	
	// 如果指定了实例ID，从指定实例获取监控项
	if instanceID != "" {
		v, count, err = model.GetAllItemByHostIDFromInstance(instanceID, HostID)
	} else {
		// 否则使用全局API
		v, count, err = model.GetAllItemByHostID(HostID)
	}
	
	if err != nil {
		response.InternalError(c, "获取错误")
		return
	}
	response.SuccessWithPage(c, v, count)
}

// GetAllTrafficItem 获取设备所有流量指标
func GetAllTrafficItem(c *gin.Context) {
	HostID := c.Query("hostid")
	instanceID := c.Query("instance_id")
	
	var v interface{}
	var count int64
	var err error
	
	// 如果指定了实例ID，从指定实例获取流量监控项
	if instanceID != "" {
		v, count, err = model.GetAllTrafficItemByHostIDFromInstance(instanceID, HostID)
	} else {
		// 否则使用全局API
		v, count, err = model.GetAllTrafficItemByHostID(HostID)
	}
	
	if err != nil {
		response.InternalError(c, "获取错误")
		return
	}
	response.SuccessWithPage(c, v, count)
}

// GetAllReceiveTrafficItem 获取设备所有出流量
func GetAllReceiveTrafficItem(c *gin.Context) {
	HostID := c.Query("hostid")
	v, count, err := model.GetReceiveTrafficeItemByHostID(HostID)
	if err != nil {
		response.InternalError(c, "获取错误")
		return
	}
	response.SuccessWithPage(c, v, count)
}
