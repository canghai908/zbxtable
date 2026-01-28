package controllers

import (
	"io"
	"net/http"
	"strconv"
	"zbxtable/models"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

// GetAllHost 获取主机列表
func GetAllHost(c *gin.Context) {
	HostType := c.Query("hosttype")
	page := c.Query("page")
	limit := c.Query("limit")
	hosts := c.Query("hosts")
	model := c.Query("model")
	ip := c.Query("ip")
	available := c.Query("available")
	
	var HostRes models.HostList
	hs, count, err := models.HostsList(HostType, page, limit, hosts, model, ip, available)
	if err != nil {
		HostRes.Code = 500
		HostRes.Message = err.Error()
		c.JSON(http.StatusOK, HostRes)
		return
	}
	HostRes.Code = 200
	HostRes.Message = "获取数据成功"
	HostRes.Data.Items = hs
	HostRes.Data.Total = count
	c.JSON(http.StatusOK, HostRes)
}

// GetHostByID 获取单个主机信息
func GetHostByID(c *gin.Context) {
	idStr := c.Param("hostid")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	if id <= 10084 {
		idStr = "10084"
	}
	v, err := models.GetHost(idStr)
	if err != nil {
		c.JSON(http.StatusOK, v.Error)
	} else {
		c.JSON(http.StatusOK, v)
	}
}

// SearchHost 搜索主机
func SearchHost(c *gin.Context) {
	name := c.Query("name")
	var HostRes models.HostList
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
	c.JSON(http.StatusOK, HostRes)
}

// UpdateHost 更新主机信息
func UpdateHost(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	var v models.Hosts
	var HostInfoRes models.HostInfo
	if err := jsoniter.Unmarshal(body, &v); err == nil {
		if _, err := models.UpdateHost(&v); err == nil {
			HostInfoRes.Code = 200
			HostInfoRes.Message = "保存成功"
			HostInfoRes.Data.Items = v
		} else {
			HostInfoRes.Code = 500
			HostInfoRes.Message = err.Error()
		}
	} else {
		HostInfoRes.Code = 500
		HostInfoRes.Message = err.Error()
	}
	c.JSON(http.StatusOK, HostInfoRes)
}

// GetMonItem 获取设备监控指标
func GetMonItem(c *gin.Context) {
	hostid := c.Param("hostid")
	hs, err := models.GetMonItem(hostid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hs)
}

// GetMonInterface 获取网络设备接口流量
func GetMonInterface(c *gin.Context) {
	hostid := c.Param("hostid")
	var HostInterfaceRes models.HostInterfaceInfo
	hs, err := models.GetInterfaceData(hostid)
	if err != nil {
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
	} else {
		HostInterfaceRes.Code = 200
		HostInterfaceRes.Message = "获取数据成功"
		HostInterfaceRes.Data.Items = hs
		HostInterfaceRes.Data.Total = int64(len(hs))
	}
	c.JSON(http.StatusOK, HostInterfaceRes)
}

// GetOneInterface 获取网络设备接口流量详情图标数据
func GetOneInterface(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	
	var v models.InterfaceData
	if err := jsoniter.Unmarshal(body, &v); err == nil {
		b, err := models.GetInterfaceGraphData(v)
		if err != nil {
			c.JSON(http.StatusOK, b)
			return
		}
		c.JSON(http.StatusOK, b)
	} else {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
	}
}

// GetMonWinFileSystem 获取windows系统监控指标
func GetMonWinFileSystem(c *gin.Context) {
	hostid := c.Param("hostid")
	var HostInterfaceRes models.HostInterfaceInfo
	hs, err := models.GetMonWinData(hostid)
	if err != nil {
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
	}
	HostInterfaceRes.Code = 200
	HostInterfaceRes.Message = "获取数据成功"
	HostInterfaceRes.Data.Items = hs
	HostInterfaceRes.Data.Total = 2
	c.JSON(http.StatusOK, HostInterfaceRes)
}

// GetMonLinFileSystem 获取文件系统详情
func GetMonLinFileSystem(c *gin.Context) {
	hostid := c.Param("hostid")
	var HostInterfaceRes models.HostInterfaceInfo
	hs, err := models.GetMonLinData(hostid)
	if err != nil {
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
	}
	HostInterfaceRes.Code = 200
	HostInterfaceRes.Message = "获取数据成功"
	HostInterfaceRes.Data.Items = hs
	HostInterfaceRes.Data.Total = 2
	c.JSON(http.StatusOK, HostInterfaceRes)
}

// GetHostGraph 查看主机图形
func GetHostGraph(c *gin.Context) {
	// 检查是否配置了密码
	if !models.IsPasswordConfigured() {
		var HostInterfaceRes models.HostInterfaceInfo
		HostInterfaceRes.Code = 403
		HostInterfaceRes.Message = "配置文件中zabbix_pass没有配置,无法查看图形,请配置后重启应用查看"
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}
	
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		var HostInterfaceRes models.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}
	
	var v models.GraphReq
	err = json.Unmarshal(body, &v)
	if err != nil {
		var HostInterfaceRes models.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}
	hostId := c.Param("hostid")
	var HostInterfaceRes models.HostInterfaceInfo
	hs, err := models.GetGraphData(hostId, v.Start, v.End)
	if err != nil {
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
	} else {
		HostInterfaceRes.Code = 200
		HostInterfaceRes.Message = "获取数据成功"
		HostInterfaceRes.Data.Items = hs
		HostInterfaceRes.Data.Total = int64(len(hs))
	}
	c.JSON(http.StatusOK, HostInterfaceRes)
}
