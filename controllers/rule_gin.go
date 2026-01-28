package controllers

import (
	"io"
	"net/http"
	"strconv"
	"zbxtable/models"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetAllRule 获取规则列表
func GetAllRule(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	name := c.Query("name")
	tenant_id := c.Query("tenant_id")
	m_type := c.Query("m_type")
	status := c.Query("status")
	
	var RulRes models.RuleResp
	cnt, al, err := models.GetRule(page, limit, name, tenant_id, m_type, status)
	if err != nil {
		RulRes.Code = 200
		RulRes.Message = err.Error()
	} else {
		RulRes.Code = 200
		RulRes.Message = "ok"
		RulRes.Data.Items = al
		RulRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, RulRes)
}

// GetRuleByID 获取单个规则
func GetRuleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var RulRes models.RuleResp
	v, err := models.GetRuleByID(id)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
	} else {
		RulRes.Code = 200
		RulRes.Message = "ok"
		RulRes.Data.Items = v
		RulRes.Data.Total = 1
	}
	c.JSON(http.StatusOK, RulRes)
}

// CreateRule 创建规则
func CreateRule(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	name := gjson.Get(string(body), "name").String()
	tenant_id := gjson.Get(string(body), "tenant_id").String()
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
	
	var RulRes models.RuleResp
	v := models.Rule{Name: name, Conditions: conditions,
		Sweek: sweek, Stime: stime, Etime: etime, Channel: channel, MType: m_type,
		UserIds: user_ids, GroupIds: group_ids, TenantID: tenant_id, Status: status, Note: note}
	_, err = models.AddRule(&v)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
	} else {
		RulRes.Code = 200
		RulRes.Message = "创建成功"
	}
	c.JSON(http.StatusOK, RulRes)
}

// UpdateRule 更新规则
func UpdateRule(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	name := gjson.Get(string(body), "name").String()
	tenant_id := gjson.Get(string(body), "tenant_id").String()
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
	
	var RulRes models.RuleResp
	v := models.Rule{ID: id, Name: name, Conditions: conditions,
		Sweek: sweek, Stime: stime, Etime: etime, Channel: channel, MType: m_type,
		UserIds: user_ids, GroupIds: group_ids, TenantID: tenant_id, Status: status, Note: note}
	err = models.UpdateRule(&v, tuserStr)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
	} else {
		RulRes.Code = 200
		RulRes.Message = "修改成功"
	}
	RulRes.Data.Items = nil
	RulRes.Data.Total = 0
	c.JSON(http.StatusOK, RulRes)
}

// UpdateRuleStatus 更新规则状态
func UpdateRuleStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	status := gjson.Get(string(body), "status").String()
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}
	
	var RulRes models.RuleResp
	v := models.Rule{ID: id, Status: status}
	err = models.UpdateRuleStatus(&v, tuserStr)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
	} else {
		RulRes.Code = 200
		RulRes.Message = "更新成功"
	}
	RulRes.Data.Items = nil
	RulRes.Data.Total = 0
	c.JSON(http.StatusOK, RulRes)
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
	
	var RulRes models.RuleResp
	err := models.DeleteRule(id, tuserStr)
	if err != nil {
		RulRes.Code = 500
		RulRes.Message = err.Error()
	} else {
		RulRes.Code = 200
		RulRes.Message = "删除成功"
	}
	RulRes.Data.Items = nil
	RulRes.Data.Total = 0
	c.JSON(http.StatusOK, RulRes)
}
