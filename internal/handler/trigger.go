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
	b, cnt, err := model.GetTriggerList(HostID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, b, cnt)
}
