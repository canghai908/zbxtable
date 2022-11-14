package controllers

import (
	"strconv"
	"zbxtable/models"
)

// HostController 设备接口
type HostController struct {
	BaseController
}

//HostRes restp
var HostRes models.HostList
var HostInfoRes models.HostInfo
var HostInterfaceRes models.HostInterfaceInfo

// URLMapping ...
func (c *HostController) URLMapping() {
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("GetOne", c.GetOne)
	c.Mapping("GetMonItem", c.GetMonItem)
	c.Mapping("GetMonInterface", c.GetMonInterface)
	c.Mapping("GetMonWinFileSystem", c.GetMonWinFileSystem)
	c.Mapping("GetMonLinFileSystem", c.GetMonLinFileSystem)
}

// Post ...
// @Title 更新设备信息
// @Description 更新设备资产信息
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	body		body 	models.Hosts	true		"hosts info"
// @Success 201 {int} models.Hosts
// @Failure 403 body is empty
// @router / [post]
func (c *HostController) Post() {
	var v models.Hosts
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		if _, err := models.UpdateHost(&v); err == nil {
			//c.Ctx.Output.SetStatus(201)
			HostInfoRes.Code = 200
			HostInfoRes.Message = "保存成功"
			HostInfoRes.Data.Items = v
			c.Data["json"] = HostInfoRes
		} else {
			HostInfoRes.Code = 500
			HostInfoRes.Message = err.Error()
			c.Data["json"] = HostInfoRes
		}
	} else {
		HostInfoRes.Code = 500
		HostInfoRes.Message = err.Error()
		c.Data["json"] = HostInfoRes
	}
	c.ServeJSON()
}

//ApplicationRes str
var ApplicationRes models.ApplicationList

// GetOne ...
// @Title 获取单个设备信息
// @Description 获取单个设备的资产信息
// @Param	X-Token		header  string	true	"x-token in header"
// @Param	hostid			path 	string	true	"hostid"
// @Success 200 {object} models.hosts
// @Failure 403 :id is empty
// @router /:hostid [get]
func (c *HostController) GetOne() {
	var idStr string
	idStr = c.Ctx.Input.Param(":hostid")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return
	}
	//hostsid <10084
	if id <= 10084 {
		idStr = "10084"
	}
	v, err := models.GetHost(idStr)
	if err != nil {
		c.Data["json"] = v.Error
	} else {
		c.Data["json"] = v
		c.ServeJSON()
		return
	}
	c.ServeJSON()
	return
}

// GetAll ...
// @Title 获取设备列表
// @Description 根据设备类型，获取设备列表
// @Param	X-Token		header  string	true	"X-Token"
// @Param	hosttype	query  string	true	"设备类型 hosttype VM_WIN WIndow VM_LIN linux HW_SRV 硬件服务器 HW_NET 网络设备"
// @Param	page	query	string	false	"页数"
// @Param	limit	query	string	false	"每页数"
// @Param	hosts	query	string	false	"主机"
// @Success 200 {object} models.Alarm
// @Failure 403
// @router / [get]
func (c *HostController) GetAll() {
	HostType := c.Ctx.Input.Query("hosttype")
	page := c.Ctx.Input.Query("page")
	limit := c.Ctx.Input.Query("limit")
	hosts := c.Ctx.Input.Query("hosts")
	model := c.Ctx.Input.Query("model")
	ip := c.Ctx.Input.Query("ip")
	available := c.Ctx.Input.Query("available")
	hs, count, err := models.HostsList(HostType, page, limit, hosts, model, ip, available)
	if err != nil {
		HostRes.Code = 500
		HostRes.Message = err.Error()
		c.Data["json"] = HostRes
		c.ServeJSON()
		return
	}
	HostRes.Code = 200
	HostRes.Message = "获取数据成功"
	HostRes.Data.Items = hs
	HostRes.Data.Total = count
	c.Data["json"] = HostRes
	c.ServeJSON()
}

// GetAll ...
// @Title 搜索主机
// @Description 根据设备类型，获取设备列表
// @Param	X-Token		header  string	true	"X-Token"
// @Param	name		query	string	true	"主机名"
// @Success 200 {object} models.host
// @Failure 403
// @router /search [get]
func (c *HostController) Search() {
	name := c.Ctx.Input.Query("name")
	val, err := models.GetNetHostByName(name)
	if err != nil {
		HostRes.Code = 500
		HostRes.Message = err.Error()
		HostRes.Data.Items = nil
		HostRes.Data.Total = 0
	} else {
		HostRes.Code = 200
		HostRes.Message = "获取数据成功"
		HostRes.Data.Items = val
		HostRes.Data.Total = int64(len(val))
	}
	c.Data["json"] = HostRes
	c.ServeJSON()
}

// GetMonItem
// @Title 获取设备监控指标
// @Description 获取设备CPU、内存、磁盘、网卡流量的相关Iitem
// @Param	X-Token	header  string	true	"x-token in header"
// @Param	hostid	path	string	ture	"hostid"
// @Success 200 {object} models.MonItemList
// @Failure 403
// @router /monitem/:hostid [get]
func (c *HostController) GetMonItem() {
	hostid := c.Ctx.Input.Param(":hostid")
	hs, err := models.GetMonItem(hostid)
	if err != nil {
		c.Data["json"] = err
		c.ServeJSON()
		return
	}
	c.Data["json"] = hs
	c.ServeJSON()
}

// GetMonItem
// @Title 获取网络设备接口流量
// @Description 获取网络设备接口流量
// @Param	X-Token	header  string	true	"x-token in header"
// @Param	hostid	path	string	ture	"hostid"
// @Success 200 {object} models.MonItemList
// @Failure 403
// @router /interface/:hostid [get]
func (c *HostController) GetMonInterface() {
	hostid := c.Ctx.Input.Param(":hostid")
	hs, err := models.GetInterfaceData(hostid)
	if err != nil {
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
	}
	HostInterfaceRes.Code = 200
	HostInterfaceRes.Message = "获取数据成功"
	HostInterfaceRes.Data.Items = hs
	HostInterfaceRes.Data.Total = int64(len(hs))
	c.Data["json"] = HostInterfaceRes
	c.ServeJSON()
}

// GetMonWinFilesystem
// @Title 获取windows系统监控指标
// @Description 获取windows网卡、磁盘信息
// @Param	X-Token	header  string	true	"x-token in header"
// @Param	hostid	path	string	ture	"hostid"
// @Success 200 {object} models.MonItemList
// @Failure 403
// @router /winmon/:hostid [get]
func (c *HostController) GetMonWinFileSystem() {
	hostid := c.Ctx.Input.Param(":hostid")
	hs, err := models.GetMonWinData(hostid)
	if err != nil {
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
	}
	HostInterfaceRes.Code = 200
	HostInterfaceRes.Message = "获取数据成功"
	HostInterfaceRes.Data.Items = hs
	HostInterfaceRes.Data.Total = 2
	c.Data["json"] = HostInterfaceRes
	c.ServeJSON()
}

// GetMonLinFilesystem
// @Title 获取文件系统详情
// @Description 获取文件系统详情
// @Param	X-Token	header  string	true	"x-token in header"
// @Param	hostid	path	string	ture	"hostid"
// @Success 200 {object} models.MonItemList
// @Failure 403
// @router /linmon/:hostid [get]
func (c *HostController) GetMonLinFileSystem() {
	hostid := c.Ctx.Input.Param(":hostid")
	hs, err := models.GetMonLinData(hostid)
	if err != nil {
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
	}
	HostInterfaceRes.Code = 200
	HostInterfaceRes.Message = "获取数据成功"
	HostInterfaceRes.Data.Items = hs
	HostInterfaceRes.Data.Total = 2
	c.Data["json"] = HostInterfaceRes
	c.ServeJSON()
}
