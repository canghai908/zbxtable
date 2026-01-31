package handler

import (
	"strconv"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetEventLogByAlarmID 获取事件日志
func GetEventLogByAlarmID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	v, err := model.GetEventLogByAlarmID(id)
	if err != nil {
		response.InternalError(c, "获取失败")
		return
	}
	response.SuccessWithPage(c, v, int64(len(v)))
}
