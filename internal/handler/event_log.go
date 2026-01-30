package handler

import (
	"net/http"
	"strconv"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
)

// GetEventLogByAlarmID 获取事件日志
func GetEventLogByAlarmID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var EveLogResp model.EventLogRes
	v, err := model.GetEventLogByAlarmID(id)
	if err != nil {
		EveLogResp.Code = 500
		EveLogResp.Message = "获取失败"
		EveLogResp.Data.Items = nil
		EveLogResp.Data.Total = 0
	} else {
		EveLogResp.Code = 200
		EveLogResp.Message = "获取成功"
		EveLogResp.Data.Items = v
		EveLogResp.Data.Total = int64(len(v))
	}
	c.JSON(http.StatusOK, EveLogResp)
}
