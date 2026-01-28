package handler

import (
	"io"
	"net/http"
	"strconv"
	"time"
	"zbxtable/internal/model"
	"zbxtable/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetUserGin 获取用户列表
func GetUserGin(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	username := c.Query("username")
	status := c.Query("status")
	
	// 从上下文获取用户名
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}
	
	var res models.UserResp
	count, hs, err := models.GetUser(page, limit, tuserStr, username, status)
	if err != nil {
		res.Code = 500
		res.Message = err.Error()
		res.Data.Items = nil
		res.Data.Total = 0
	} else {
		res.Code = 200
		res.Message = "获取数据成功"
		res.Data.Items = hs
		res.Data.Total = count
	}
	c.JSON(http.StatusOK, res)
}

// CreateUserGin 新建用户
func CreateUserGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	username := gjson.Get(string(body), "username").String()
	password := gjson.Get(string(body), "password").String()
	role := gjson.Get(string(body), "role").String()
	email := gjson.Get(string(body), "email").String()
	wechat := gjson.Get(string(body), "wechat").String()
	wechat_robot_key := gjson.Get(string(body), "wechat_robot_key").String()
	phone := gjson.Get(string(body), "phone").String()
	ding_talk := gjson.Get(string(body), "ding_talk").String()
	p, _ := utils.PasswordHash(password)
	var operation string
	switch role {
	case "admin":
		operation = "['add', 'edit', 'delete','update']"
	case "user":
		operation = "[]"
	default:
		operation = "[]"
	}
	
	var res models.UserResp
	v := models.Manager{Username: username, Password: p, Operation: operation,
		Email: email, Wechat: wechat, WechatRobotKey: wechat_robot_key, Phone: phone, DingTalk: ding_talk,
		Status: 0, Role: role, Created: time.Now(),
		Avatar: "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
	}
	_, err = models.AddUser(&v)
	if err != nil {
		res.Code = 500
		res.Message = err.Error()
	} else {
		res.Code = 200
		res.Message = "创建用户成功"
	}
	res.Data.Items = nil
	res.Data.Total = 1
	c.JSON(http.StatusOK, res)
}

// UpdateUserGin 更新用户信息
func UpdateUserGin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		var res models.UserResp
		res.Code = 500
		res.Message = err.Error()
		res.Data.Items = nil
		res.Data.Total = 0
		c.JSON(http.StatusOK, res)
		return
	}
	
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	// 从上下文获取用户名
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}
	
	password := gjson.Get(string(body), "password").String()
	role := gjson.Get(string(body), "role").String()
	email := gjson.Get(string(body), "email").String()
	wechat := gjson.Get(string(body), "wechat").String()
	wechat_robot_key := gjson.Get(string(body), "wechat_robot_key").String()
	phone := gjson.Get(string(body), "phone").String()
	dingTalk := gjson.Get(string(body), "ding_talk").String()
	var operation string
	switch role {
	case "admin":
		operation = "['add', 'edit', 'delete','update']"
	case "user":
		operation = "[]"
	default:
		operation = "[]"
	}
	var pass string
	if password != "" {
		pass, _ = utils.PasswordHash(password)
	} else {
		pass = ""
	}

	var res models.UserResp
	v := models.Manager{ID: id, Password: pass, Email: email, Wechat: wechat, WechatRobotKey: wechat_robot_key,
		Phone: phone, DingTalk: dingTalk, Role: role, Operation: operation}
	err = models.UpdateUser(&v, tuserStr)
	if err != nil {
		res.Code = 500
		res.Message = err.Error()
	} else {
		res.Code = 200
		res.Message = "修改成功"
	}
	res.Data.Items = nil
	res.Data.Total = 0
	c.JSON(http.StatusOK, res)
}

// DeleteUserGin 删除用户
func DeleteUserGin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		var res models.UserResp
		res.Code = 500
		res.Message = err.Error()
		res.Data.Items = nil
		res.Data.Total = 0
		c.JSON(http.StatusOK, res)
		return
	}
	
	// 从上下文获取用户名
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}
	
	var res models.UserResp
	if err := models.DeleteUser(id, tuserStr); err == nil {
		res.Code = 200
		res.Message = "删除成功"
	} else {
		res.Code = 500
		res.Message = err.Error()
	}
	res.Data.Items = nil
	res.Data.Total = 0
	c.JSON(http.StatusOK, res)
}
