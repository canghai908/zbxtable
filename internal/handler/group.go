package handler

import (
	"io"
	"strconv"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetAllGroup 获取用户组列表
func GetAllGroup(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	name := c.Query("name")

	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}

	count, hs, err := model.GetGroup(page, limit, tuserStr, name)
	if err != nil {
		response.DatabaseError(c, "获取用户组列表失败: "+err.Error())
		return
	}
	
	response.SuccessWithPage(c, hs, count)
}

// CreateGroup 新建用户组
func CreateGroup(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	name := gjson.Get(string(body), "name").String()
	note := gjson.Get(string(body), "note").String()
	
	if name == "" {
		response.ValidationError(c, "用户组名称不能为空")
		return
	}
	
	v := model.UserGroup{Name: name, Note: note}
	_, err = model.AddUserGroup(&v)
	if err != nil {
		response.DatabaseError(c, "创建用户组失败: "+err.Error())
		return
	}
	
	response.SuccessWithMessage(c, "创建用户组成功", nil)
}

// UpdateGroup 更新用户组信息
func UpdateGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户组ID")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	name := gjson.Get(string(body), "name").String()
	note := gjson.Get(string(body), "note").String()
	
	if name == "" {
		response.ValidationError(c, "用户组名称不能为空")
		return
	}
	
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}

	v := model.UserGroup{ID: id, Name: name, Note: note}
	err = model.UpdateUserGroup(&v, tuserStr)
	if err != nil {
		response.DatabaseError(c, "更新用户组失败: "+err.Error())
		return
	}
	
	response.SuccessWithMessage(c, "修改成功", nil)
}

// UpdateGroupMember 更新组成员
func UpdateGroupMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户组ID")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	member := gjson.Get(string(body), "member").String()
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}

	v := model.UserGroup{ID: id, Member: member}
	err = model.UpdateGroupMember(&v, tuserStr)
	if err != nil {
		response.DatabaseError(c, "更新组成员失败: "+err.Error())
		return
	}
	
	response.SuccessWithMessage(c, "修改成功", nil)
}

// DeleteGroup 删除群组
func DeleteGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户组ID")
		return
	}

	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}

	if err := model.DeleteGroup(id, tuserStr); err != nil {
		response.DatabaseError(c, "删除用户组失败: "+err.Error())
		return
	}
	
	response.SuccessWithMessage(c, "删除成功", nil)
}
