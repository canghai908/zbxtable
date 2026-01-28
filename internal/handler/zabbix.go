package handler

import (
	"net/http"
	"strconv"
	"strings"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
)

type createZabbixInstanceReq struct {
	Name   string `json:"name" binding:"required"`
	WebURL string `json:"web_url" binding:"required"`
	User   string `json:"user"`
	Pass   string `json:"pass"`
	Token  string `json:"token"`
}

type testZabbixReq struct {
	WebURL string `json:"web_url" binding:"required"`
	User   string `json:"user"`
	Pass   string `json:"pass"`
	Token  string `json:"token"`
}

type enableZabbixReq struct {
	Enabled bool `json:"enabled"`
}

type updateZabbixInstanceReq struct {
	Name    string `json:"name" binding:"required"`
	WebURL  string `json:"web_url" binding:"required"`
	User    string `json:"user"`
	Pass    string `json:"pass"`
	Token   string `json:"token"`
	Enabled bool   `json:"enabled"`
}

// toSafeResponse 将 ZabbixInstance 转换为安全的响应结构
func toSafeResponse(inst *models.ZabbixInstance) ZabbixInstanceSafeResponse {
	if inst == nil {
		return ZabbixInstanceSafeResponse{}
	}
	var lastTestAt *string
	if inst.LastTestAt != nil {
		timeStr := inst.LastTestAt.Format("2006-01-02 15:04:05")
		lastTestAt = &timeStr
	}
	return ZabbixInstanceSafeResponse{
		ID:              inst.ID,
		Name:            inst.Name,
		WebURL:          inst.WebURL,
		Enabled:         inst.Enabled,
		User:            inst.User,
		IsActive:        inst.IsActive,
		Version:         inst.Version,
		LastTestOk:      inst.LastTestOk,
		LastTestMessage: inst.LastTestMessage,
		LastTestAt:      lastTestAt,
		CreatedAt:       inst.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:       inst.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// ZabbixInstanceSafeResponse 安全的响应结构，隐藏敏感信息
type ZabbixInstanceSafeResponse struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	WebURL          string  `json:"web_url"`
	Enabled         bool    `json:"enabled"`
	User            string  `json:"user"`
	IsActive        bool    `json:"is_active"`
	Version         string  `json:"version"`
	LastTestOk      bool    `json:"last_test_ok"`
	LastTestMessage string  `json:"last_test_message"`
	LastTestAt      *string `json:"last_test_at"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	// 不包含 User, Pass, Token 等敏感字段
}

// ListZabbixInstancesGin GET /v1/zabbix/instances
func ListZabbixInstancesGin(c *gin.Context) {
	list, err := models.ListZabbixInstances()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 转换为安全的响应结构，隐藏敏感信息
	safeList := make([]ZabbixInstanceSafeResponse, 0, len(list))
	for _, inst := range list {
		var lastTestAt *string
		if inst.LastTestAt != nil {
			timeStr := inst.LastTestAt.Format("2006-01-02 15:04:05")
			lastTestAt = &timeStr
		}
		safeList = append(safeList, ZabbixInstanceSafeResponse{
			ID:              inst.ID,
			Name:            inst.Name,
			WebURL:          inst.WebURL,
			User:            inst.User,
			Enabled:         inst.Enabled,
			IsActive:        inst.IsActive,
			Version:         inst.Version,
			LastTestOk:      inst.LastTestOk,
			LastTestMessage: inst.LastTestMessage,
			LastTestAt:      lastTestAt,
			CreatedAt:       inst.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:       inst.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok", "data": safeList})
}

// TestZabbixInstanceConfigGin POST /v1/zabbix/instances/test
// 用于“新增页面先测试再创建”
func TestZabbixInstanceConfigGin(c *gin.Context) {
	var req testZabbixReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error(), "data": gin.H{"success": false}})
		return
	}
	ver, err := models.TestZabbixInstanceConfig(req.WebURL, req.User, req.Pass, req.Token)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "连接失败: " + err.Error(), "data": gin.H{"success": false}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "连接成功", "data": gin.H{"success": true, "version": ver}})
}

// TestZabbixInstanceGin POST /v1/zabbix/instances/:id/test
// 测试已保存的实例，并更新版本/状态
func TestZabbixInstanceGin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: id", "data": gin.H{"success": false}})
		return
	}
	inst, ver, err := models.TestAndUpdateZabbixInstance(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "连接失败: " + err.Error(), "data": gin.H{"success": false}})
		return
	}

	// 转换为安全响应
	safeInst := toSafeResponse(inst)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "连接成功", "data": gin.H{"success": true, "version": ver, "instance": safeInst}})
}

// CreateZabbixInstanceGin POST /v1/zabbix/instances
func CreateZabbixInstanceGin(c *gin.Context) {
	var req createZabbixInstanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 强制：新增前必须测试通过（后端再测一遍，避免绕过前端）
	ver, testErr := models.TestZabbixInstanceConfig(req.WebURL, req.User, req.Pass, req.Token)
	if testErr != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "连接失败: " + testErr.Error(), "data": gin.H{"success": false}})
		return
	}

	inst := &models.ZabbixInstance{
		Name:            strings.TrimSpace(req.Name),
		WebURL:          strings.TrimSpace(req.WebURL),
		User:            req.User,
		Pass:            req.Pass,
		Token:           req.Token,
		Enabled:         true,
		Version:         ver,
		LastTestOk:      true,
		LastTestMessage: "连接成功",
	}
	if err := models.CreateZabbixInstance(inst); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}

	// 创建成功后：若当前没有 active，则自动设为当前并预热（开始获取数据的基础：初始化 API + 版本）
	active, _ := models.GetActiveZabbixInstance()
	if active == nil {
		_, _ = models.ActivateZabbixInstance(inst.ID)
	} else if active.IsActive {
		// 不抢占当前，但也预热自身的版本/状态（异步不阻断）
		go func(id int64) {
			_, _, _ = models.TestAndUpdateZabbixInstance(id)
		}(inst.ID)
	}

	// 转换为安全响应
	safeInst := toSafeResponse(inst)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": safeInst})
}

// ActivateZabbixInstanceGin PUT /v1/zabbix/instances/:id/activate
func ActivateZabbixInstanceGin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: id"})
		return
	}
	inst, err := models.ActivateZabbixInstance(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	// 转换为安全响应
	safeInst := toSafeResponse(inst)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已切换当前 Zabbix", "data": safeInst})
}

// EnableZabbixInstanceGin PUT /v1/zabbix/instances/:id/enabled
func EnableZabbixInstanceGin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: id"})
		return
	}
	var req enableZabbixReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	inst, err := models.SetZabbixInstanceEnabled(id, req.Enabled)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	// 转换为安全响应
	safeInst := toSafeResponse(inst)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok", "data": safeInst})
}

// UpdateZabbixInstanceGin PUT /v1/zabbix/instances/:id
func UpdateZabbixInstanceGin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: id"})
		return
	}
	var req updateZabbixInstanceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	inst, err := models.UpdateZabbixInstance(id, &models.ZabbixInstance{
		Name:    strings.TrimSpace(req.Name),
		WebURL:  strings.TrimSpace(req.WebURL),
		User:    req.User,
		Pass:    req.Pass,
		Token:   req.Token,
		Enabled: req.Enabled,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	// 转换为安全响应
	safeInst := toSafeResponse(inst)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": safeInst})
}

// DeleteZabbixInstanceGin DELETE /v1/zabbix/instances/:id
func DeleteZabbixInstanceGin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: id"})
		return
	}
	if err := models.DeleteZabbixInstance(id); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// GetActiveZabbixInstanceGin GET /v1/zabbix/active
func GetActiveZabbixInstanceGin(c *gin.Context) {
	inst, err := models.GetActiveZabbixInstance()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	// 转换为安全响应
	safeInst := toSafeResponse(inst)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "ok", "data": safeInst})
}

// GetZabbixInstanceGin GET /v1/zabbix/instances/:id
// 获取单个实例详情（包含敏感信息，用于编辑）
func GetZabbixInstanceGin(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 400, "message": "参数错误: id"})
		return
	}
	inst, err := models.GetZabbixInstanceByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	// 返回完整信息（包含敏感字段），用于编辑
	var lastTestAt *string
	if inst.LastTestAt != nil {
		timeStr := inst.LastTestAt.Format("2006-01-02 15:04:05")
		lastTestAt = &timeStr
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "ok",
		"data": gin.H{
			"id":                inst.ID,
			"name":              inst.Name,
			"web_url":           inst.WebURL,
			"user":              inst.User,
			"pass":              inst.Pass,
			"token":             inst.Token,
			"enabled":           inst.Enabled,
			"is_active":         inst.IsActive,
			"version":           inst.Version,
			"last_test_ok":      inst.LastTestOk,
			"last_test_message": inst.LastTestMessage,
			"last_test_at":      lastTestAt,
			"created_at":        inst.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":        inst.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}
