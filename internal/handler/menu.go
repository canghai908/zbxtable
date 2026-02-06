package handler

import (
	"io"
	"strconv"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetAllMenus 获取所有菜单
func GetAllMenus(c *gin.Context) {
	menus, err := model.GetAllMenus()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, menus)
}

// GetMenuByID 根据ID获取菜单
func GetMenuByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	menu, err := model.GetMenuByID(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, menu)
}

// GetParentMenus 获取所有父菜单
func GetParentMenus(c *gin.Context) {
	menus, err := model.GetParentMenus()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, menus)
}

// GetMenusByParentID 根据父菜单ID获取子菜单
func GetMenusByParentID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	menus, err := model.GetMenusByParentID(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, menus)
}

// CreateMenu 创建菜单
func CreateMenu(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	// 解析请求参数
	name := gjson.Get(string(body), "name").String()
	path := gjson.Get(string(body), "path").String()
	router := gjson.Get(string(body), "router").String()
	icon := gjson.Get(string(body), "icon").String()
	role := gjson.Get(string(body), "role").String()
	permission := gjson.Get(string(body), "permission").String()
	parentId := int(gjson.Get(string(body), "parent_id").Int())
	invisible := gjson.Get(string(body), "invisible").Bool()
	isAvailable := gjson.Get(string(body), "is_available").Bool()
	highlight := gjson.Get(string(body), "highlight").String()
	cacheAble := gjson.Get(string(body), "cacheable").Bool()

	// 验证必填字段
	if name == "" || path == "" || router == "" {
		response.BadRequest(c, "菜单名称、路径和路由不能为空")
		return
	}

	// 如果没有指定角色，默认为 admin,user
	if role == "" {
		role = "admin,user"
	}

	menu := &model.Menu{
		ParentId:    parentId,
		Name:        name,
		Path:        path,
		Router:      router,
		Icon:        icon,
		Role:        role,
		Permission:  permission,
		Invisible:   invisible,
		IsAvailable: isAvailable,
		Highlight:   highlight,
		CacheAble:   cacheAble,
	}

	err = model.CreateMenu(menu)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", menu)
}

// UpdateMenu 更新菜单
func UpdateMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	// 解析请求参数
	name := gjson.Get(string(body), "name").String()
	path := gjson.Get(string(body), "path").String()
	router := gjson.Get(string(body), "router").String()
	icon := gjson.Get(string(body), "icon").String()
	role := gjson.Get(string(body), "role").String()
	permission := gjson.Get(string(body), "permission").String()
	parentId := int(gjson.Get(string(body), "parent_id").Int())
	invisible := gjson.Get(string(body), "invisible").Bool()
	isAvailable := gjson.Get(string(body), "is_available").Bool()
	highlight := gjson.Get(string(body), "highlight").String()
	cacheAble := gjson.Get(string(body), "cacheable").Bool()

	// 验证必填字段
	if name == "" || path == "" || router == "" {
		response.BadRequest(c, "菜单名称、路径和路由不能为空")
		return
	}

	menu := &model.Menu{
		Id:          id,
		ParentId:    parentId,
		Name:        name,
		Path:        path,
		Router:      router,
		Icon:        icon,
		Role:        role,
		Permission:  permission,
		Invisible:   invisible,
		IsAvailable: isAvailable,
		Highlight:   highlight,
		CacheAble:   cacheAble,
	}

	err = model.UpdateMenu(menu)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", menu)
}

// DeleteMenu 删除菜单
func DeleteMenu(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}

	err = model.DeleteMenu(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
