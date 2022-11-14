package controllers

import (
	"github.com/tidwall/gjson"
	"zbxtable/models"
)

// TopoDataController 绑定数据点
type TopoDataController struct {
	BaseController
}

// URLMapping ...
func (c *TopoDataController) URLMapping() {
	c.Mapping("CreateData", c.CreateData)
	c.Mapping("GetTopologyOne", c.GetTopologyOne)
	//c.Mapping("CreateTopology", c.CreateTopology)
	//c.Mapping("InitTopology", c.InitTopology)
}

// CreateTopology ...
// @Title 创建数据采集点
// @Description 创建数据采集点
// @Param	X-Token	header  string	true	"X-Token"
// @Param	body	body 	models.TopologyData true	"body for TopoData content"
// @Success 200 {object} models.TopologyData
// @Failure 403
// @router / [post]
func (c *TopoDataController) CreateData() {
	pid := gjson.Get(string(c.Ctx.Input.RequestBody), "pid").String()
	vtype := gjson.Get(string(c.Ctx.Input.RequestBody), "v_type").String()
	ptype := gjson.Get(string(c.Ctx.Input.RequestBody), "type").String()
	tid := gjson.Get(string(c.Ctx.Input.RequestBody), "tid").String()
	v := models.TopologyData{PID: pid, Type: ptype, VType: vtype, TID: tid}
	_, err := models.AddTopoData(&v)
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

// GetTopologyOne ...
// @Title 获取数据
// @Description 获取数据
// @Param	X-Token		header  string	true	"X-Token"
// @Param	id			path 	string	true	"id"
// @Success 200 {object} models.Topology
// @Failure 403
// @router /:id [get]
func (c *TopoDataController) GetTopologyOne() {
	idStr := c.Ctx.Input.Param(":id")
	//id, _ := strconv.Atoi(idStr)
	var TopologyRes models.TopologyInfo
	//err := models.UpdateEdgedata(id)
	//err := models.UpdateNodedata(id)
	//if err != nil {
	//	TopologyRes.Code = 500
	//	TopologyRes.Message = err.Error()
	//} else {

	TopologyRes.Code = 200
	TopologyRes.Message = idStr
	//}
	c.Data["json"] = TopologyRes
	c.ServeJSON()
}
