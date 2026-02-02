package handler

import (
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetAllTrigger 获取触发器列表
func GetAllTrigger(c *gin.Context) {
	b, cnt, err := model.GetTriggers()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, b, cnt)
}

// GetTriggerList 根据主机ID获取Trigger列表
func GetTriggerList(c *gin.Context) {
	HostID := c.Query("hostid")
	zid := c.Query("zid")

	var b interface{}
	var cnt int64
	var err error

	// 如果指定了实例ID，从指定实例获取触发器
	if zid != "" {
		b, cnt, err = model.GetTriggerListFromInstance(zid, HostID)
	} else {
		// 否则使用全局API
		b, cnt, err = model.GetTriggerList(HostID)
	}

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, b, cnt)
}
