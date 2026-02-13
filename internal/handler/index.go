package handler

import (
	"strconv"

	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// VersionInfo 版本信息响应结构
type VersionInfo struct {
	Version   string `json:"version"`
	GitHash   string `json:"gitHash"`
	BuildTime string `json:"buildTime"`
}

// GetRouters 获取异步路由
func GetRouters(c *gin.Context) {
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}
	info, err := model.GetRouter(tuserStr)
	if err != nil {
		response.InternalError(c, "获取失败")
		return
	}
	response.SuccessWithMessage(c, "获取成功", info)
}

// GetBaseInfo 获取首页基础信息
func GetBaseInfo(c *gin.Context) {
	info, err := model.GetCountHost()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "获取成功", info)
}

// GetResourceTop 获取资源使用TopN
func GetResourceTop(c *gin.Context) {
	host_type := c.Query("host_type")
	metrics_type := c.Query("metrics_type")
	top_num := c.Query("top_num")
	// 兜底：如果未传 top_num，则读取系统配置 dash_top_num（再兜底到 5）
	// 未传/非法都兜底到 dash_top_num（再兜底到 5）
	if _, err := strconv.ParseInt(top_num, 10, 64); top_num == "" || err != nil {
		top_num = model.GetConfigValueByKey("dash_top_num", "5")
	}
	info, err := model.GetTopList(host_type, metrics_type, top_num)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "获取成功", info)
}

// GetInventory 获取图谱资源
func GetInventory(c *gin.Context) {
	info, err := model.GetInventory()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "获取成功", info)
}

// GetOverview 获取汇总状态
func GetOverview(c *gin.Context) {
	info, err := model.GetOverviewData()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "获取成功", info)
}

// GetEgressData 获取出口带宽数据（新版本，支持多个出口）
func GetEgressData(c *gin.Context) {
	// 尝试从新缓存读取数据
	data, err := model.CacheGet("egress_data")
	if err == nil && data != "" {
		// 返回新格式的数据（JSON 数组）
		var egressList []map[string]interface{}
		jsonErr := json.Unmarshal([]byte(data), &egressList)
		if jsonErr == nil && len(egressList) > 0 {
			response.SuccessWithMessage(c, "获取成功", egressList)
			return
		}
	}

	// 如果新缓存没有数据，尝试使用旧的 API（向后兼容）
	info, err := model.GetEgressData()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "获取成功", info)
}

// GetVersion 获取版本信息
func GetVersion(c *gin.Context) {
	versionInfo := VersionInfo{
		Version:   model.Version,
		GitHash:   model.GitHash,
		BuildTime: model.BuildTime,
	}
	response.SuccessWithMessage(c, "获取成功", versionInfo)
}

// GetZbxSession 获取Zabbix Session
func GetZbxSession(c *gin.Context) {
	session, err := model.GetZbxSession()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "获取成功", session)
}
