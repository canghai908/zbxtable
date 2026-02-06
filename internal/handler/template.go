package handler

import (
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetAllTemplateList 获取所有模板列表
func GetTemplateList(c *gin.Context) {
	zidStr := c.Query("zid")
	if zidStr == "" {
		response.InternalError(c, "实例ID为空")
		return
	}
	list, err := model.TemplateListGetFromInstance(zidStr)
	if err != nil {
		response.InternalError(c, "获取模版错误")
		return
	}
	response.Success(c, list)
}

// GetItemByTemplateID 根据模板ID获取监控项
func GetItemByTemplateID(c *gin.Context) {
	templateid := c.Param("templateid")
	zidStr := c.Query("zid")
	if zidStr == "" {
		response.InternalError(c, "缺少实例ID")
	}
	list, err := model.TemplateByItemFromInstance(zidStr, templateid)
	if err != nil {
		response.InternalError(c, "获取模版错误")
		return
	}
	response.Success(c, list)
}
