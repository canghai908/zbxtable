package handler

import (
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetAllTemplate 获取模板列表
func GetAllTemplate(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	templates := c.Query("templates")

	b, cnt, err := model.TemplateGet(page, limit, templates)
	if err != nil {
		response.InternalError(c, "获取模版错误")
		return
	}
	response.SuccessWithPage(c, b, cnt)
}

// GetAllTemplateList 获取所有模板列表
func GetAllTemplateList(c *gin.Context) {
	zidStr := c.Query("zid")
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
	list, err := model.TemplateByItemFromInstance(templateid, zidStr)
	if err != nil {
		response.InternalError(c, "获取模版错误")
		return
	}
	response.Success(c, list)
}
