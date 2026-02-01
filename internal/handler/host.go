package handler

import (
	"zbxtable/pkg/response"
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
	instanceID := c.Query("instance_id")

	var hs []model.Hosts
	var count int64
	var err error

	// 如果指定了实例ID，从指定实例获取主机
	if instanceID != "" {
		hs, count, err = model.HostsListFromInstance(instanceID, HostType, page, limit, hosts, mode, ip, available)
	} else {
		// 否则使用多实例查询（查询所有实例）
		hs, count, err = model.HostsListMultiInstance(HostType, page, limit, hosts, mode, ip, available)
	}

	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, hs, count)
}

// GetHostByID 获取单个主机信息（多实例支持）
func GetHostByID(c *gin.Context) {
	idStr := c.Param("hostid")
	instanceIDStr := c.Query("instance_id") // 从查询参数获取实例ID
	
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	if id <= 10084 {
		idStr = "10084"
	}

	var instanceID int
	if instanceIDStr != "" {
		// 如果提供了实例ID，直接使用
		instanceID, err = strconv.Atoi(instanceIDStr)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("无效的实例ID: %v", err))
			return
		}
	} else {
		// 如果没有提供实例ID，查找该主机所属的实例
		instanceID, err = model.FindInstanceByHostID(idStr)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("无法找到主机所属实例: %v", err))
			return
		}
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("获取实例连接失败: %v", err))
		return
	}

	// 使用该实例的 API 查询主机详情
	v, err := model.GetHostFromInstance(inst, idStr)
	if err != nil {
		response.InternalError(c, err.Error())
	} else {
		response.SuccessWithMessage(c, "获取数据成功", v)
	}
}

// SearchHost 搜索主机（支持实例筛选）
func SearchHost(c *gin.Context) {
	name := c.Query("name")
	tenantID := c.Query("tenant_id")
	instanceID := c.Query("instance_id")
	
	// 如果提供了实例ID或租户ID，从指定实例查询
	if instanceID != "" || tenantID != "" {
		var inst *model.APIInstance
		var err error
		
		if instanceID != "" {
			// 如果提供了数字 instance_id，直接使用
			instID, err := strconv.Atoi(instanceID)
			if err != nil {
				response.InternalError(c, "无效的实例ID")
				return
			}
			inst, err = model.GetAPIByInstanceID(instID)
		} else {
			// 如果提供了 tenant_id 字符串，使用 GetZabbixInstanceAPI
			inst, err = model.GetZabbixInstanceAPI(tenantID)
		}
		
		if err != nil {
			response.InternalError(c, fmt.Sprintf("获取实例连接失败: %v", err))
			return
		}
		
		// 从该实例查询主机
		val, err := model.SearchHostFromInstance(inst, name)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}
		response.SuccessWithPage(c, val, int64(len(val)))
	} else {
		// 兼容旧逻辑：从缓存查询
		val, err := model.GetNetHostByName(name)
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}
		response.SuccessWithPage(c, val, int64(len(val)))
	}
}

// UpdateHost 更新主机信息
func UpdateHost(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	var v model.Hosts
	if err := jsoniter.Unmarshal(body, &v); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	
	if _, err := model.UpdateHost(&v); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	
	response.SuccessWithMessage(c, "保存成功", v)
}

// GetMonItem 获取设备监控指标（多实例支持）
func GetMonItem(c *gin.Context) {
	hostid := c.Param("hostid")

	// 查找该主机所属的实例
	instanceID, err := model.FindInstanceByHostID(hostid)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("无法找到主机所属实例: %v", err))
		return
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("获取实例连接失败: %v", err))
		return
	}

	// 使用该实例的 API 查询监控指标
	hs, err := model.GetMonItemFromInstance(inst, hostid)
	if err != nil {
		response.InternalError(c, err.Error())
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
		response.InternalError(c, fmt.Sprintf("无法找到主机所属实例: %v", err))
		return
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("获取实例连接失败: %v", err))
		return
	}

	// 使用该实例的 API 查询接口数据
	hs, err := model.GetInterfaceDataFromInstance(inst, hostid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, hs, int64(len(hs)))
}

// GetOneInterface 获取网络设备接口流量详情图标数据
func GetOneInterface(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	var v model.InterfaceData
	if err := jsoniter.Unmarshal(body, &v); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	
	b, err := model.GetInterfaceGraphData(v)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, b)
}

// GetMonWinFileSystem 获取windows系统监控指标（多实例支持）
func GetMonWinFileSystem(c *gin.Context) {
	hostid := c.Param("hostid")
	instanceIDStr := c.Query("instance_id") // 从查询参数获取实例ID

	var instanceID int
	var err error
	if instanceIDStr != "" {
		// 如果提供了实例ID，直接使用
		instanceID, err = strconv.Atoi(instanceIDStr)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("无效的实例ID: %v", err))
			return
		}
	} else {
		// 如果没有提供实例ID，查找该主机所属的实例
		instanceID, err = model.FindInstanceByHostID(hostid)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("无法找到主机所属实例: %v", err))
			return
		}
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("获取实例连接失败: %v", err))
		return
	}

	// 使用该实例的 API 查询 Windows 监控数据
	hs, err := model.GetMonWinDataFromInstance(inst, hostid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, hs, 2)
}

// GetMonLinFileSystem 获取文件系统详情（多实例支持）
func GetMonLinFileSystem(c *gin.Context) {
	hostid := c.Param("hostid")
	instanceIDStr := c.Query("instance_id") // 从查询参数获取实例ID

	var instanceID int
	var err error
	if instanceIDStr != "" {
		// 如果提供了实例ID，直接使用
		instanceID, err = strconv.Atoi(instanceIDStr)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("无效的实例ID: %v", err))
			return
		}
	} else {
		// 如果没有提供实例ID，查找该主机所属的实例
		instanceID, err = model.FindInstanceByHostID(hostid)
		if err != nil {
			response.InternalError(c, fmt.Sprintf("无法找到主机所属实例: %v", err))
			return
		}
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("获取实例连接失败: %v", err))
		return
	}

	// 使用该实例的 API 查询 Linux 监控数据
	hs, err := model.GetMonLinDataFromInstance(inst, hostid)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, hs, 2)
}

// GetHostGraph 查看主机图形（自动识别所属实例）
func GetHostGraph(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	var v model.GraphReq
	err = json.Unmarshal(body, &v)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	hostId := c.Param("hostid")

	// 查找该主机所属的实例
	instanceID, err := model.FindInstanceByHostID(hostId)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("无法找到主机所属实例: %v", err))
		return
	}

	// 获取该实例的 API 连接
	inst, err := model.GetAPIByInstanceID(instanceID)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("获取实例连接失败: %v", err))
		return
	}

	// 检查是否配置了用户名和密码（查看图形需要）
	tenant, _ := model.GetZabbixTenantByID(int64(instanceID))
	if tenant == nil || tenant.User == "" || tenant.Pass == "" {
		response.Forbidden(c, fmt.Sprintf("实例 %s 未配置用户名和密码，无法查看图形。请在【系统设置 > Zabbix 实例管理】中配置用户名和密码（注意：使用 Token 方式无法查看图形）", inst.Name))
		return
	}

	// 使用该实例的 API 查询图形数据
	hs, err := model.GetGraphDataFromInstance(inst, hostId, v.Start, v.End)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, hs, int64(len(hs)))
}
