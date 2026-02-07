package handler

import (
	"io"
	"strconv"
	"time"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"
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

	count, hs, err := model.GetUser(page, limit, tuserStr, username, status)
	if err != nil {
		response.DatabaseError(c, "获取用户列表失败: "+err.Error())
		return
	}

	response.SuccessWithPage(c, hs, count)
}

// CreateUserGin 新建用户
func CreateUserGin(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
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

	// 验证必填字段
	if username == "" {
		response.ValidationError(c, "用户名不能为空")
		return
	}
	if password == "" {
		response.ValidationError(c, "密码不能为空")
		return
	}

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

	v := model.User{Username: username, Password: p, Operation: operation,
		Email: email, Wechat: wechat, WechatRobotKey: wechat_robot_key, Phone: phone, DingTalk: ding_talk,
		Status: 0, Role: role, Created: time.Now(),
		Avatar: "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
	}
	_, err = model.AddUser(&v)
	if err != nil {
		response.DatabaseError(c, "创建用户失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "创建用户成功", nil)
}

// UpdateUserGin 更新用户信息
func UpdateUserGin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	// 从上下文获取用户名
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}

	// 解析 JSON 以检查哪些字段实际存在
	bodyJSON := gjson.Parse(string(body))

	// 构建用户对象，只设置请求中实际包含的字段
	v := model.User{ID: id}

	// 密码处理
	if bodyJSON.Get("password").Exists() {
		password := bodyJSON.Get("password").String()
		if password != "" {
			pass, _ := utils.PasswordHash(password)
			v.Password = pass
		}
	}

	// 角色处理
	if bodyJSON.Get("role").Exists() {
		role := bodyJSON.Get("role").String()
		v.Role = role
		switch role {
		case "admin":
			v.Operation = "['add', 'edit', 'delete','update']"
		case "user":
			v.Operation = "[]"
		default:
			v.Operation = "[]"
		}
	}

	// 其他字段 - 只有在请求中存在时才设置
	if bodyJSON.Get("email").Exists() {
		v.Email = bodyJSON.Get("email").String()
	}
	if bodyJSON.Get("wechat").Exists() {
		v.Wechat = bodyJSON.Get("wechat").String()
	}
	if bodyJSON.Get("wechat_robot_key").Exists() {
		v.WechatRobotKey = bodyJSON.Get("wechat_robot_key").String()
	}
	if bodyJSON.Get("phone").Exists() {
		v.Phone = bodyJSON.Get("phone").String()
	}
	if bodyJSON.Get("ding_talk").Exists() {
		v.DingTalk = bodyJSON.Get("ding_talk").String()
	}
	if bodyJSON.Get("theme").Exists() {
		v.Theme = bodyJSON.Get("theme").String()
	}

	err = model.UpdateUser(&v, tuserStr)
	if err != nil {
		response.DatabaseError(c, "更新用户失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "修改成功", nil)
}

// DeleteUserGin 删除用户
func DeleteUserGin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	// 从上下文获取用户名
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}

	if err := model.DeleteUser(id, tuserStr); err != nil {
		response.DatabaseError(c, "删除用户失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
