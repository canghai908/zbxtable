package handler

import (
	"io"
	"strconv"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetAllRule 获取规则列表
func GetAllRule(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	name := c.Query("name")
	zid := c.Query("zids")
	m_type := c.Query("m_type")
	status := c.Query("status")

	cnt, al, err := model.GetRule(page, limit, name, zid, m_type, status)
	if err != nil {
		response.DatabaseError(c, "获取规则列表失败: "+err.Error())
		return
	}

	response.SuccessWithPage(c, al, cnt)
}

// GetRuleByID 获取单个规则
func GetRuleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	v, err := model.GetRuleByID(id)
	if err != nil {
		response.NotFound(c, "规则不存在")
		return
	}

	response.Success(c, v)
}

// CreateRule 创建规则
func CreateRule(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	name := gjson.Get(string(body), "name").String()
	zids := gjson.Get(string(body), "z_ids").String()
	conditions := gjson.Get(string(body), "conditions").String()
	sweek := gjson.Get(string(body), "s_week").String()
	stime := gjson.Get(string(body), "s_time").String()
	etime := gjson.Get(string(body), "e_time").String()
	channel := gjson.Get(string(body), "channel").String()
	user_ids := gjson.Get(string(body), "user_ids").String()
	group_ids := gjson.Get(string(body), "group_ids").String()
	m_type := gjson.Get(string(body), "m_type").String()
	note := gjson.Get(string(body), "note").String()
	status := gjson.Get(string(body), "status").String()

	if name == "" {
		response.ValidationError(c, "规则名称不能为空")
		return
	}

	v := model.Rule{Name: name, Conditions: conditions,
		Sweek: sweek, Stime: stime, Etime: etime, Channel: channel, MType: m_type,
		UserIds: user_ids, GroupIds: group_ids, ZIDs: zids, Status: status, Note: note}
	_, err = model.AddRule(&v)
	if err != nil {
		response.DatabaseError(c, "创建规则失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "创建成功", nil)
}

// UpdateRule 更新规则
func UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	name := gjson.Get(string(body), "name").String()
	zids := gjson.Get(string(body), "z_ids").String()
	conditions := gjson.Get(string(body), "conditions").String()
	sweek := gjson.Get(string(body), "s_week").String()
	stime := gjson.Get(string(body), "s_time").String()
	etime := gjson.Get(string(body), "e_time").String()
	channel := gjson.Get(string(body), "channel").String()
	user_ids := gjson.Get(string(body), "user_ids").String()
	group_ids := gjson.Get(string(body), "group_ids").String()
	m_type := gjson.Get(string(body), "m_type").String()
	note := gjson.Get(string(body), "note").String()
	status := gjson.Get(string(body), "status").String()

	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}

	v := model.Rule{ID: id, Name: name, Conditions: conditions,
		Sweek: sweek, Stime: stime, Etime: etime, Channel: channel, MType: m_type,
		UserIds: user_ids, GroupIds: group_ids, ZIDs: zids, Status: status, Note: note}
	err = model.UpdateRule(&v, tuserStr)
	if err != nil {
		response.DatabaseError(c, "更新规则失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "修改成功", nil)
}

// UpdateRuleStatus 更新规则状态
func UpdateRuleStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	status := gjson.Get(string(body), "status").String()
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}

	v := model.Rule{ID: id, Status: status}
	err = model.UpdateRuleStatus(&v, tuserStr)
	if err != nil {
		response.DatabaseError(c, "更新规则状态失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteRule 删除规则
func DeleteRule(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}

	err := model.DeleteRule(id, tuserStr)
	if err != nil {
		response.DatabaseError(c, "删除规则失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
