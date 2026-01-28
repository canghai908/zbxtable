package handler

import (
	"io"
	"net/http"
	"strconv"
	"zbxtable/internal/model"

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
	
	var GrooupRes models.GroupResp
	count, hs, err := models.GetGroup(page, limit, tuserStr, name)
	if err != nil {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
		GrooupRes.Data.Items = nil
		GrooupRes.Data.Total = 0
	} else {
		GrooupRes.Code = 200
		GrooupRes.Message = "获取数据成功"
		GrooupRes.Data.Items = hs
		GrooupRes.Data.Total = count
	}
	c.JSON(http.StatusOK, GrooupRes)
}

// CreateGroup 新建用户组
func CreateGroup(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	name := gjson.Get(string(body), "name").String()
	note := gjson.Get(string(body), "note").String()
	var GrooupRes models.GroupResp
	v := models.UserGroup{Name: name, Note: note}
	_, err = models.AddUserGroup(&v)
	if err != nil {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
	} else {
		GrooupRes.Code = 200
		GrooupRes.Message = "创建用户组成功"
	}
	GrooupRes.Data.Items = nil
	GrooupRes.Data.Total = 1
	c.JSON(http.StatusOK, GrooupRes)
}

// UpdateGroup 更新用户组信息
func UpdateGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		var GrooupRes models.GroupResp
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
		c.JSON(http.StatusOK, GrooupRes)
		return
	}
	
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	name := gjson.Get(string(body), "name").String()
	note := gjson.Get(string(body), "note").String()
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}
	
	var GrooupRes models.GroupResp
	v := models.UserGroup{ID: id, Name: name, Note: note}
	err = models.UpdateUserGroup(&v, tuserStr)
	if err != nil {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
	} else {
		GrooupRes.Code = 200
		GrooupRes.Message = "修改成功"
	}
	GrooupRes.Data.Items = nil
	GrooupRes.Data.Total = 0
	c.JSON(http.StatusOK, GrooupRes)
}

// UpdateGroupMember 更新组成员
func UpdateGroupMember(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		var GrooupRes models.GroupResp
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
		c.JSON(http.StatusOK, GrooupRes)
		return
	}
	
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	member := gjson.Get(string(body), "member").String()
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}
	
	var GrooupRes models.GroupResp
	v := models.UserGroup{ID: id, Member: member}
	err = models.UpdateGroupMember(&v, tuserStr)
	if err != nil {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
	} else {
		GrooupRes.Code = 200
		GrooupRes.Message = "修改成功"
	}
	GrooupRes.Data.Items = nil
	GrooupRes.Data.Total = 0
	c.JSON(http.StatusOK, GrooupRes)
}

// DeleteGroup 删除群组
func DeleteGroup(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		var GrooupRes models.GroupResp
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
		c.JSON(http.StatusOK, GrooupRes)
		return
	}
	
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}
	
	var GrooupRes models.GroupResp
	if err := models.DeleteGroup(id, tuserStr); err == nil {
		GrooupRes.Code = 200
		GrooupRes.Message = "删除成功"
	} else {
		GrooupRes.Code = 500
		GrooupRes.Message = err.Error()
	}
	GrooupRes.Data.Items = nil
	GrooupRes.Data.Total = 0
	c.JSON(http.StatusOK, GrooupRes)
}
