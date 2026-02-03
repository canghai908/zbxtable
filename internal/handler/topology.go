package handler

import (
	"io"
	"strconv"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetAllTopology 获取拓扑列表
func GetAllTopology(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	name := c.Query("name")

	count, hs, err := model.GetAllTopology(page, limit, name)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, hs, count)
}

// GetTopologyByID 获取拓扑详情
func GetTopologyByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的拓扑ID")
		return
	}

	v, err := model.GetTopologyById(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "获取成功", v)
}

// CreateTopology 创建拓扑
func CreateTopology(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	nodes := gjson.Get(string(body), "nodes").String()
	edges := gjson.Get(string(body), "edges").String()
	topology := gjson.Get(string(body), "topology").String()
	backgroundImage := gjson.Get(string(body), "background_image").String()
	canvasWidth := int(gjson.Get(string(body), "canvas_width").Int())
	canvasHeight := int(gjson.Get(string(body), "canvas_height").Int())

	// 验证必填字段
	if topology == "" {
		response.ValidationError(c, "拓扑名称不能为空")
		return
	}

	// 设置默认值
	if canvasWidth == 0 {
		canvasWidth = 3000
	}
	if canvasHeight == 0 {
		canvasHeight = 2000
	}

	v := model.Topology{
		Nodes:           nodes,
		Edges:           edges,
		Topology:        topology,
		BackgroundImage: backgroundImage,
		CanvasWidth:     canvasWidth,
		CanvasHeight:    canvasHeight,
	}

	id, err := model.AddTopology(&v)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "创建成功", map[string]interface{}{
		"id": id,
	})
}

// UpdateTopology 更新拓扑
func UpdateTopology(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的拓扑ID")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	nodes := gjson.Get(string(body), "nodes").String()
	edges := gjson.Get(string(body), "edges").String()
	topology := gjson.Get(string(body), "topology").String()
	backgroundImage := gjson.Get(string(body), "background_image").String()
	canvasWidth := int(gjson.Get(string(body), "canvas_width").Int())
	canvasHeight := int(gjson.Get(string(body), "canvas_height").Int())

	// 验证必填字段
	if topology == "" {
		response.ValidationError(c, "拓扑名称不能为空")
		return
	}

	// 设置默认值
	if canvasWidth == 0 {
		canvasWidth = 3000
	}
	if canvasHeight == 0 {
		canvasHeight = 2000
	}

	v := model.Topology{
		ID:              id,
		Nodes:           nodes,
		Edges:           edges,
		Topology:        topology,
		BackgroundImage: backgroundImage,
		CanvasWidth:     canvasWidth,
		CanvasHeight:    canvasHeight,
	}

	err = model.UpdateTopologyByID(&v)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteTopology 删除拓扑
func DeleteTopology(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的拓扑ID")
		return
	}

	err = model.DeleteTopology(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// UpdateTopologyStatus 更新拓扑状态
func UpdateTopologyStatus(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	idStr := gjson.Get(string(body), "id").String()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的拓扑ID")
		return
	}

	v := model.Topology{ID: id}
	err = model.UpdateTopologyStatusByID(&v)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "状态更新成功", nil)
}

// GetPublicTopologyByID 公开访问拓扑详情（无需认证，仅限已共享的拓扑）
func GetPublicTopologyByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "无效的拓扑ID")
		return
	}

	v, err := model.GetTopologyById(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	// 只允许访问已共享的拓扑
	if v.Status != "1" {
		response.Forbidden(c, "该拓扑未共享，无法公开访问")
		return
	}

	response.SuccessWithMessage(c, "获取成功", v)
}
