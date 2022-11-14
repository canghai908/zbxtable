package controllers

import (
	"github.com/tidwall/gjson"
	"strconv"
	"zbxtable/models"
)

// TopologyController 拓扑图
type TopologyController struct {
	BaseController
}

//TopologyRes is used
var TopologyRes models.TopologyList

// URLMapping ...
func (c *TopologyController) URLMapping() {
	c.Mapping("GetTopologyAll", c.GetTopologyAll)
	c.Mapping("GetTopologyOne", c.GetTopologyOne)
	c.Mapping("CreateTopology", c.CreateTopology)
	c.Mapping("Delete", c.Delete)
	c.Mapping("Put", c.Put)
	c.Mapping("DeployTopology", c.DeployTopology)
}

// GetTopologyAll ...
// @Title 获取拓扑列表
// @Description 获取拓扑列表
// @Param	X-Token		header  string	true	"X-Token"
// @Param	page	query	string	false	"页数"
// @Param	limit	query	string	false	"每页数"
// @Param	name	query	string	false	"拓扑名称"
// @Success 200 {object} models.Topology
// @Failure 403
// @router / [get]
func (c *TopologyController) GetTopologyAll() {
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	name := c.Ctx.Input.Query("name")
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
	c.Data["json"] = TopologyRes
	c.ServeJSON()
}

// GetTopologyOne ...
// @Title 获取拓扑详情
// @Description 获取拓扑详情
// @Param	X-Token		header  string	true	"X-Token"
// @Param	id			path 	string	true	"id"
// @Success 200 {object} models.Topology
// @Failure 403
// @router /:id [get]
func (c *TopologyController) GetTopologyOne() {
	idStr := c.Ctx.Input.Param(":id")
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
	c.Data["json"] = TopologyResp
	c.ServeJSON()
}

// CreateTopology ...
// @Title 创建拓扑图
// @Description 创建拓扑图
// @Param	X-Token	header  string	true	"X-Token"
// @Param	body	body 	models.Topology true	"body for Topology content"
// @Success 200 {object} models.Topology
// @Failure 403
// @router / [post]
func (c *TopologyController) CreateTopology() {
	nodes := gjson.Get(string(c.Ctx.Input.RequestBody), "nodes").String()
	edges := gjson.Get(string(c.Ctx.Input.RequestBody), "edges").String()
	topology := gjson.Get(string(c.Ctx.Input.RequestBody), "topology").String()
	v := models.Topology{Nodes: nodes, Edges: edges, Topology: topology}
	_, err := models.AddTopology(&v)
	if err != nil {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	} else {
		TopologyRes.Code = 200
		TopologyRes.Message = "创建成功"
	}
	c.Data["json"] = TopologyRes
	c.ServeJSON()
}

// DeployTopology ...
// @Title 发布拓扑
// @Description 发布拓扑
// @Param	X-Token	header  string	true	"X-Token"
// @Param	body	body 	models.Topology true	"body for Topology content"
// @Success 200 {object} models.Topology
// @Failure 403
// @router /deploy [post]
func (c *TopologyController) DeployTopology() {
	idStr := gjson.Get(string(c.Ctx.Input.RequestBody), "id").String()
	id, _ := strconv.Atoi(idStr)
	v := models.Topology{ID: id}

	err := models.UpdateTopologyStatusByID(&v)
	if err != nil {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	} else {
		TopologyRes.Code = 200
		TopologyRes.Message = "更新成功"
	}
	TopologyRes.Data.Items = []models.Topology{}
	c.Data["json"] = TopologyRes
	c.ServeJSON()
}

// saveTopology..
// @Title 更新拓扑图
// @Description 更新拓扑图
// @Param	X-Token	header  string	true	"X-Token"
// @Param	body	body 	models.Topology  true	"body for Topology content"
// @Success 200 {object} models.Topology
// @Failure 403
// @router / [put]
func (c *TopologyController) Put() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v := models.Topology{ID: id}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		if err := models.UpdateTopologyByID(&v); err == nil {
			TopologyRes.Code = 200
			TopologyRes.Message = "保存成功"
			c.Data["json"] = "OK"
		} else {
			TopologyRes.Code = 500
			TopologyRes.Message = err.Error()
		}
	} else {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	}
	c.Data["json"] = TopologyRes
	c.ServeJSON()
}

// Delete ...
// @Title 删除拓扑图
// @Description 删除拓扑图
// @Param	X-Token	header  string	true	"X-Token"
// @Param	id		path 	string	true		"The id you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *TopologyController) Delete() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	if err := models.DeleteTopology(id); err == nil {
		TopologyRes.Code = 200
		TopologyRes.Message = "删除成功"
	} else {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	}
	c.Data["json"] = TopologyRes
	c.ServeJSON()
}
