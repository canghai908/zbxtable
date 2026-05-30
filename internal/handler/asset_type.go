package handler

import (
	"io"
	"net/http"
	"strconv"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetAllAssetTypes 获取所有资产类型列表
func GetAllAssetTypes(c *gin.Context) {
	list, err := model.GetAllAssetTypes()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, list)
}

// GetAssetTypeByID 获取单个资产类型
func GetAssetTypeByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	at, err := model.GetAssetTypeByID(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, at)
}

// CreateAssetType 创建资产类型
func CreateAssetType(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	name := gjson.Get(string(body), "name").String()
	typeCode := gjson.Get(string(body), "type_code").String()
	if name == "" || typeCode == "" {
		response.BadRequest(c, "名称和类型代码不能为空")
		return
	}

	monitorType := gjson.Get(string(body), "monitor_type").String()
	if monitorType == "" {
		monitorType = "agent"
	}
	menuGroup := gjson.Get(string(body), "menu_group").String()
	at := &model.AssetType{
		Name:        name,
		TypeCode:    typeCode,
		Icon:        gjson.Get(string(body), "icon").String(),
		Description: gjson.Get(string(body), "description").String(),
		MonitorType: monitorType,
		SortOrder:   int(gjson.Get(string(body), "sort_order").Int()),
		MenuGroup:   menuGroup,
	}
	if err := model.CreateAssetType(at); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", at)
}

// UpdateAssetType 更新资产类型
func UpdateAssetType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	name := gjson.Get(string(body), "name").String()
	if name == "" {
		response.BadRequest(c, "名称不能为空")
		return
	}

	// type_code 更新时不可改，从数据库读取原值
	existing, err := model.GetAssetTypeByID(id)
	if err != nil {
		response.InternalError(c, "资产类型不存在")
		return
	}

	monitorType := gjson.Get(string(body), "monitor_type").String()
	if monitorType == "" {
		monitorType = existing.MonitorType
	}
	menuGroup := gjson.Get(string(body), "menu_group").String()
	if menuGroup == "" {
		menuGroup = existing.MenuGroup
	}
	at := &model.AssetType{
		ID:          id,
		Name:        name,
		TypeCode:    existing.TypeCode,
		Icon:        gjson.Get(string(body), "icon").String(),
		Description: gjson.Get(string(body), "description").String(),
		MonitorType: monitorType,
		SortOrder:   int(gjson.Get(string(body), "sort_order").Int()),
		MenuGroup:   menuGroup,
	}
	if err := model.UpdateAssetType(at); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", at)
}

// GetAssetTypeFields 获取指定资产类型的列表字段配置
func GetAssetTypeFields(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	fields, err := model.GetAssetTypeFields(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, fields)
}

// UpdateAssetTypeFields 更新指定资产类型的列表字段配置
func UpdateAssetTypeFields(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}
	var fields []model.AssetTypeField
	if err := json.Unmarshal(body, &fields); err != nil {
		response.BadRequest(c, "字段配置格式错误: "+err.Error())
		return
	}
	if err := model.UpdateAssetTypeFields(id, fields); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "字段配置已保存", fields)
}

// DeleteAssetType 删除资产类型（若有 System 记录引用则返回错误）
func DeleteAssetType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	if err := model.DeleteAssetType(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
