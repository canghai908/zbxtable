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

// GetAllTopology 获取拓扑列表
func GetAllTopology(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	name := c.Query("name")

	var TopologyRes model.TopologyList
	count, hs, err := model.GetAllTopology(page, limit, name)
	if err != nil {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	} else {
		TopologyRes.Code = 200
		TopologyRes.Message = "获取数据成功"
		TopologyRes.Data.Items = hs
		TopologyRes.Data.Total = count
	}
	c.JSON(http.StatusOK, TopologyRes)
}

// GetTopologyByID 获取拓扑详情
func GetTopologyByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var TopologyResp model.TopologyInfo
	v, err := model.GetTopologyById(id)
	if err != nil {
		TopologyResp.Code = 500
		TopologyResp.Message = err.Error()
	} else {
		TopologyResp.Code = 200
		TopologyResp.Message = "获取成功"
		TopologyResp.Data.Items = v
	}
	c.JSON(http.StatusOK, TopologyResp)
}

// CreateTopology 创建拓扑
func CreateTopology(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	nodes := gjson.Get(string(body), "nodes").String()
	edges := gjson.Get(string(body), "edges").String()
	topology := gjson.Get(string(body), "topology").String()

	var TopologyRes model.TopologyList
	v := model.Topology{Nodes: nodes, Edges: edges, Topology: topology}
	_, err = model.AddTopology(&v)
	if err != nil {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	} else {
		TopologyRes.Code = 200
		TopologyRes.Message = "创建成功"
	}
	c.JSON(http.StatusOK, TopologyRes)
}

// UpdateTopology 更新拓扑
func UpdateTopology(c *gin.Context) {
	idStr := c.Param("id")
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	id, _ := strconv.Atoi(idStr)
	nodes := gjson.Get(string(body), "nodes").String()
	edges := gjson.Get(string(body), "edges").String()
	topology := gjson.Get(string(body), "topology").String()

	var TopologyRes model.TopologyList
	v := model.Topology{ID: id, Nodes: nodes, Edges: edges, Topology: topology}
	err = model.UpdateTopologyByID(&v)
	if err != nil {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	} else {
		TopologyRes.Code = 200
		TopologyRes.Message = "更新成功"
	}
	c.JSON(http.StatusOK, TopologyRes)
}

// DeleteTopology 删除拓扑
func DeleteTopology(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var TopologyRes model.TopologyList
	if err := model.DeleteTopology(id); err == nil {
		TopologyRes.Code = 200
		TopologyRes.Message = "删除成功"
	} else {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	}
	c.JSON(http.StatusOK, TopologyRes)
}

// UpdateTopologyStatus 更新拓扑状态
func UpdateTopologyStatus(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	idStr := gjson.Get(string(body), "id").String()
	id, _ := strconv.Atoi(idStr)
	var TopologyRes model.TopologyList
	v := model.Topology{ID: id}
	err = model.UpdateTopologyStatusByID(&v)
	if err != nil {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	} else {
		TopologyRes.Code = 200
		TopologyRes.Message = "更新成功"
	}
	TopologyRes.Data.Items = []model.Topology{}
	c.JSON(http.StatusOK, TopologyRes)
}

// GetPublicTopologyByID 公开访问拓扑详情（无需认证，仅限已共享的拓扑）
func GetPublicTopologyByID(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	var TopologyResp model.TopologyInfo
	v, err := model.GetTopologyById(id)
	if err != nil {
		TopologyResp.Code = 500
		TopologyResp.Message = err.Error()
		c.JSON(http.StatusOK, TopologyResp)
		return
	}

	// 只允许访问已共享的拓扑
	if v.Status != "1" {
		TopologyResp.Code = 403
		TopologyResp.Message = "该拓扑未共享，无法公开访问"
		c.JSON(http.StatusOK, TopologyResp)
		return
	}

	TopologyResp.Code = 200
	TopologyResp.Message = "获取成功"
	TopologyResp.Data.Items = v
	c.JSON(http.StatusOK, TopologyResp)
}
