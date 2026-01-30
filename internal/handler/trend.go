package handler

import (
	"net/http"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
)

// GetTrendByItemID 获取趋势数据
func GetTrendByItemID(c *gin.Context) {
	itemID := c.Query("item_id")
	limit := c.Query("limit")
	v, err := model.GetTrendByItemID(itemID, limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, v)
	}
}
