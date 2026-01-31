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

// GetAllTemplateAll 获取所有模板
func GetAllTemplateAll(c *gin.Context) {
	b, cnt, err := model.TemplateAllGet()
	if err != nil {
		response.InternalError(c, "获取模版错误")
		return
	}
	response.SuccessWithPage(c, b, cnt)
}

// GetAllTemplateList 获取所有模板列表
func GetAllTemplateList(c *gin.Context) {
	instanceID := c.Query("instance_id")
	
	b, cnt, err := model.TemplateListGetFromInstance(instanceID)
	if err != nil {
		response.InternalError(c, "获取模版错误")
		return
	}
	response.SuccessWithPage(c, b, cnt)
}

// GetItemByTemplateID 根据模板ID获取监控项
func GetItemByTemplateID(c *gin.Context) {
	templateid := c.Param("templateid")
	instanceID := c.Query("instance_id")
	
	b, cnt, err := model.TemplateByItemFromInstance(templateid, instanceID)
	if err != nil {
		response.InternalError(c, "获取模版错误")
		return
	}
	response.SuccessWithPage(c, b, cnt)
}
