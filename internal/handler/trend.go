package handler

import (
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetTrendByItemID 获取趋势数据
func GetTrendByItemID(c *gin.Context) {
	itemID := c.Query("item_id")
	limit := c.Query("limit")
	v, err := model.GetTrendByItemID(itemID, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, v)
}
