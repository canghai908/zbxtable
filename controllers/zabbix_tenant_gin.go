package controllers

import (
	"io"
	"net/http"
	"strconv"
	"zbxtable/models"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// ZabbixTenantBindingSafeResponse 安全的租户绑定响应结构，隐藏敏感信息
type ZabbixTenantBindingSafeResponse struct {
	ID               int    `json:"id"`
	TenantID         string `json:"tenant_id"`
	ZabbixInstanceID int    `json:"zabbix_instance_id"`
	Enabled          bool   `json:"enabled"`
	// 不包含 Token 字段
}

// toTenantSafeResponse 将租户绑定转换为安全响应
func toTenantSafeResponse(binding *models.ZabbixTenantBinding) ZabbixTenantBindingSafeResponse {
	if binding == nil {
		return ZabbixTenantBindingSafeResponse{}
	}
	return ZabbixTenantBindingSafeResponse{
		ID:               binding.ID,
		TenantID:         binding.TenantID,
		ZabbixInstanceID: binding.ZabbixInstanceID,
		Enabled:          binding.Enabled,
	}
}

// ListZabbixTenantBindingsGin 列出租户绑定
func ListZabbixTenantBindingsGin(c *gin.Context) {
	list, err := models.ListZabbixTenantBindings()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	
	// 转换为安全的响应结构，隐藏 token
	safeList := make([]ZabbixTenantBindingSafeResponse, 0, len(list))
	for _, binding := range list {
		safeList = append(safeList, toTenantSafeResponse(&binding))
	}
	
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok", "data": safeList})
}

// UpsertZabbixTenantBindingGin 新增/更新租户绑定（按 tenant_id upsert）
func UpsertZabbixTenantBindingGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	tenantID := gjson.Get(string(body), "tenant_id").String()
	token := gjson.Get(string(body), "token").String()
	zid := int(gjson.Get(string(body), "zabbix_instance_id").Int())
	enabled := gjson.Get(string(body), "enabled").Bool()

	m := &models.ZabbixTenantBinding{
		TenantID:         tenantID,
		Token:            token,
		ZabbixInstanceID: zid,
		Enabled:          enabled,
	}
	if err := models.UpsertZabbixTenantBinding(m); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok"})
}

// DeleteZabbixTenantBindingGin 删除租户绑定
func DeleteZabbixTenantBindingGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	if err := models.DeleteZabbixTenantBinding(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok"})
}
