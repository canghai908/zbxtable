package handler

import (
	"net/http"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
)

// GetAllTrigger 获取触发器列表
func GetAllTrigger(c *gin.Context) {
	var TriggersRes model.TriggersRes
	b, cnt, err := model.GetTriggers()
	if err != nil {
		TriggersRes.Code = 500
		TriggersRes.Message = err.Error()
	} else {
		TriggersRes.Code = 200
		TriggersRes.Message = "获取成功"
		TriggersRes.Data.Items = b
		TriggersRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, TriggersRes)
}

// GetTriggerList 根据主机ID获取Trigger列表
func GetTriggerList(c *gin.Context) {
	HostID := c.Query("hostid")
	var TriggersListRes model.TriggersListRes
	b, cnt, err := model.GetTriggerList(HostID)
	if err != nil {
		TriggersListRes.Code = 500
		TriggersListRes.Message = err.Error()
	} else {
		TriggersListRes.Code = 200
		TriggersListRes.Message = "获取成功"
		TriggersListRes.Data.Items = b
		TriggersListRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, TriggersListRes)
}
