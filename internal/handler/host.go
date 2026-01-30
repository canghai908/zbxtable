package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

// GetAllHost 获取主机列表（多实例聚合）
func GetAllHost(c *gin.Context) {
	HostType := c.Query("hosttype")
	page := c.Query("page")
	limit := c.Query("limit")
	hosts := c.Query("hosts")
	mode := c.Query("model")
	ip := c.Query("ip")
	available := c.Query("available")

	var HostRes model.HostList
	// 使用多实例查询
	hs, count, err := model.HostsListMultiInstance(HostType, page, limit, hosts, mode, ip, available)
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

// GetHostByID 获取单个主机信息（多实例支持）
func GetHostByID(c *gin.Context) {
	idStr := c.Param("hostid")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	if id <= 10084 {
		idStr = "10084"
	}

	// 查找该主机所属的实例
	instanceID, err := model.FindInstanceByHostID(idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("无法找到主机所属实例: %v", err)})
		return
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("获取实例连接失败: %v", err)})
		return
	}

	// 使用该实例的 API 查询主机详情
	v, err := model.GetHostFromInstance(inst, idStr)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 200, "message": "获取数据成功", "data": v})
	}
}

// SearchHost 搜索主机
func SearchHost(c *gin.Context) {
	name := c.Query("name")
	var HostRes model.HostList
	val, err := model.GetNetHostByName(name)
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

	var v model.Hosts
	var HostInfoRes model.HostInfo
	if err := jsoniter.Unmarshal(body, &v); err == nil {
		if _, err := model.UpdateHost(&v); err == nil {
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

// GetMonItem 获取设备监控指标（多实例支持）
func GetMonItem(c *gin.Context) {
	hostid := c.Param("hostid")

	// 查找该主机所属的实例
	instanceID, err := model.FindInstanceByHostID(hostid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("无法找到主机所属实例: %v", err)})
		return
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": fmt.Sprintf("获取实例连接失败: %v", err)})
		return
	}

	// 使用该实例的 API 查询监控指标
	hs, err := model.GetMonItemFromInstance(inst, hostid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hs)
}

// GetMonInterface 获取网络设备接口流量（多实例支持）
func GetMonInterface(c *gin.Context) {
	hostid := c.Param("hostid")

	// 查找该主机所属的实例
	instanceID, err := model.FindInstanceByHostID(hostid)
	if err != nil {
		var HostInterfaceRes model.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = fmt.Sprintf("无法找到主机所属实例: %v", err)
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		var HostInterfaceRes model.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = fmt.Sprintf("获取实例连接失败: %v", err)
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}

	var HostInterfaceRes model.HostInterfaceInfo
	// 使用该实例的 API 查询接口数据
	hs, err := model.GetInterfaceDataFromInstance(inst, hostid)
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

	var v model.InterfaceData
	if err := jsoniter.Unmarshal(body, &v); err == nil {
		b, err := model.GetInterfaceGraphData(v)
		if err != nil {
			c.JSON(http.StatusOK, b)
			return
		}
		c.JSON(http.StatusOK, b)
	} else {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
	}
}

// GetMonWinFileSystem 获取windows系统监控指标（多实例支持）
func GetMonWinFileSystem(c *gin.Context) {
	hostid := c.Param("hostid")

	// 查找该主机所属的实例
	instanceID, err := model.FindInstanceByHostID(hostid)
	if err != nil {
		var HostInterfaceRes model.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = fmt.Sprintf("无法找到主机所属实例: %v", err)
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		var HostInterfaceRes model.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = fmt.Sprintf("获取实例连接失败: %v", err)
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}

	var HostInterfaceRes model.HostInterfaceInfo
	// 使用该实例的 API 查询 Windows 监控数据
	hs, err := model.GetMonWinDataFromInstance(inst, hostid)
	if err != nil {
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
	} else {
		HostInterfaceRes.Code = 200
		HostInterfaceRes.Message = "获取数据成功"
		HostInterfaceRes.Data.Items = hs
		HostInterfaceRes.Data.Total = 2
	}
	c.JSON(http.StatusOK, HostInterfaceRes)
}

// GetMonLinFileSystem 获取文件系统详情（多实例支持）
func GetMonLinFileSystem(c *gin.Context) {
	hostid := c.Param("hostid")

	// 查找该主机所属的实例
	instanceID, err := model.FindInstanceByHostID(hostid)
	if err != nil {
		var HostInterfaceRes model.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = fmt.Sprintf("无法找到主机所属实例: %v", err)
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		var HostInterfaceRes model.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = fmt.Sprintf("获取实例连接失败: %v", err)
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}

	var HostInterfaceRes model.HostInterfaceInfo
	// 使用该实例的 API 查询 Linux 监控数据
	hs, err := model.GetMonLinDataFromInstance(inst, hostid)
	if err != nil {
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
	} else {
		HostInterfaceRes.Code = 200
		HostInterfaceRes.Message = "获取数据成功"
		HostInterfaceRes.Data.Items = hs
		HostInterfaceRes.Data.Total = 2
	}
	c.JSON(http.StatusOK, HostInterfaceRes)
}

// GetHostGraph 查看主机图形（自动识别所属实例）
func GetHostGraph(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		var HostInterfaceRes model.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = "请求体读取失败"
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}

	var v model.GraphReq
	err = json.Unmarshal(body, &v)
	if err != nil {
		var HostInterfaceRes model.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = err.Error()
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}

	hostId := c.Param("hostid")

	// 查找该主机所属的实例
	instanceID, err := model.FindInstanceByHostID(hostId)
	if err != nil {
		var HostInterfaceRes model.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = fmt.Sprintf("无法找到主机所属实例: %v", err)
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		var HostInterfaceRes model.HostInterfaceInfo
		HostInterfaceRes.Code = 500
		HostInterfaceRes.Message = fmt.Sprintf("获取实例连接失败: %v", err)
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}

	// 检查是否配置了用户名和密码（查看图形需要）
	tenant, _ := model.GetZabbixTenantByID(int64(instanceID))
	if tenant == nil || tenant.User == "" || tenant.Pass == "" {
		var HostInterfaceRes model.HostInterfaceInfo
		HostInterfaceRes.Code = 403
		HostInterfaceRes.Message = fmt.Sprintf("实例 %s 未配置用户名和密码，无法查看图形。请在【系统设置 > Zabbix 实例管理】中配置用户名和密码（注意：使用 Token 方式无法查看图形）", inst.Name)
		c.JSON(http.StatusOK, HostInterfaceRes)
		return
	}

	var HostInterfaceRes model.HostInterfaceInfo
	// 使用该实例的 API 查询图形数据
	hs, err := model.GetGraphDataFromInstance(inst, hostId, v.Start, v.End)
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
