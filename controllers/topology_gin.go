package controllers

import (
	"io"
	"net/http"
	"strconv"
	"zbxtable/models"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetAllTopology 获取拓扑列表
func GetAllTopology(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	name := c.Query("name")
	
	var TopologyRes models.TopologyList
	count, hs, err := models.GetAllTopology(page, limit, name)
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
	var TopologyResp models.TopologyInfo
	v, err := models.GetTopologyById(id)
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
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	nodes := gjson.Get(string(body), "nodes").String()
	edges := gjson.Get(string(body), "edges").String()
	topology := gjson.Get(string(body), "topology").String()
	
	var TopologyRes models.TopologyList
	v := models.Topology{Nodes: nodes, Edges: edges, Topology: topology}
	_, err = models.AddTopology(&v)
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
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	id, _ := strconv.Atoi(idStr)
	nodes := gjson.Get(string(body), "nodes").String()
	edges := gjson.Get(string(body), "edges").String()
	topology := gjson.Get(string(body), "topology").String()
	
	var TopologyRes models.TopologyList
	v := models.Topology{ID: id, Nodes: nodes, Edges: edges, Topology: topology}
	err = models.UpdateTopologyByID(&v)
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
	
	var TopologyRes models.TopologyList
	if err := models.DeleteTopology(id); err == nil {
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
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	idStr := gjson.Get(string(body), "id").String()
	id, _ := strconv.Atoi(idStr)
	var TopologyRes models.TopologyList
	v := models.Topology{ID: id}
	err = models.UpdateTopologyStatusByID(&v)
	if err != nil {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	} else {
		TopologyRes.Code = 200
		TopologyRes.Message = "更新成功"
	}
	TopologyRes.Data.Items = []models.Topology{}
	c.JSON(http.StatusOK, TopologyRes)
}
