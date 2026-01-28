package controllers

import (
	"net/http"
	"zbxtable/models"

	"github.com/gin-gonic/gin"
)

// GetAllTemplate 获取模板列表
func GetAllTemplate(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	templates := c.Query("templates")
	
	var TemplateRes models.TemplateList
	b, cnt, err := models.TemplateGet(page, limit, templates)
	if err != nil {
		TemplateRes.Code = 500
		TemplateRes.Message = "获取模版错误"
		c.JSON(http.StatusOK, TemplateRes)
		return
	}
	TemplateRes.Code = 200
	TemplateRes.Message = "获取成功"
	TemplateRes.Data.Items = b
	TemplateRes.Data.Total = cnt
	c.JSON(http.StatusOK, TemplateRes)
}

// GetAllTemplateAll 获取所有模板
func GetAllTemplateAll(c *gin.Context) {
	var TemplateRes models.TemplateList
	b, cnt, err := models.TemplateAllGet()
	if err != nil {
		TemplateRes.Code = 500
		TemplateRes.Message = "获取模版错误"
	} else {
		TemplateRes.Code = 200
		TemplateRes.Message = "获取成功"
		TemplateRes.Data.Items = b
		TemplateRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, TemplateRes)
}

// GetAllTemplateList 获取所有模板列表
func GetAllTemplateList(c *gin.Context) {
	var TemplateRes models.TemplateList
	b, cnt, err := models.TemplateListGet()
	if err != nil {
		TemplateRes.Code = 500
		TemplateRes.Message = "获取模版错误"
	} else {
		TemplateRes.Code = 200
		TemplateRes.Message = "获取成功"
		TemplateRes.Data.Items = b
		TemplateRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, TemplateRes)
}

// GetItemByTemplateID 根据模板ID获取监控项
func GetItemByTemplateID(c *gin.Context) {
	templateid := c.Param("templateid")
	var TemplateRes models.TemplateList
	b, cnt, err := models.TemplateByItem(templateid)
	if err != nil {
		TemplateRes.Code = 500
		TemplateRes.Message = "获取模版错误"
	} else {
		TemplateRes.Code = 200
		TemplateRes.Message = "获取成功"
		TemplateRes.Data.Items = b
		TemplateRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, TemplateRes)
}
