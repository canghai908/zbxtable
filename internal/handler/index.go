package handler

import (
	"net/http"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
)

// GetRouters 获取异步路由
func GetRouters(c *gin.Context) {
	tuser, _ := c.Get("username")
	tuserStr := ""
	if tuser != nil {
		tuserStr = tuser.(string)
	}
	var routeRes model.RouRes
	info, err := model.GetRouter(tuserStr)
	if err != nil {
		routeRes.Code = 500
		routeRes.Message = "获取失败"
		routeRes.Data.Items = info
	} else {
		routeRes.Code = 200
		routeRes.Message = "获取成功"
		routeRes.Data.Items = info
	}
	c.JSON(http.StatusOK, routeRes)
}

// GetBaseInfo 获取首页基础信息
func GetBaseInfo(c *gin.Context) {
	var InfoRes model.InfoRes
	info, err := model.GetCountHost()
	if err != nil {
		InfoRes.Code = 500
		InfoRes.Message = err.Error()
	} else {
		InfoRes.Code = 200
		InfoRes.Message = "获取成功"
		InfoRes.Data.Items = info
	}
	c.JSON(http.StatusOK, InfoRes)
}

// GetResourceTop 获取资源使用TopN
func GetResourceTop(c *gin.Context) {
	host_type := c.Query("host_type")
	metrics_type := c.Query("metrics_type")
	top_num := c.Query("top_num")
	var TopRes model.TopRes
	info, err := model.GetTopList(host_type, metrics_type, top_num)
	if err != nil {
		TopRes.Code = 500
		TopRes.Message = err.Error()
	} else {
		TopRes.Code = 200
		TopRes.Message = "获取成功"
		TopRes.Data.Items = info
	}
	c.JSON(http.StatusOK, TopRes)
}

// GetInventory 获取图谱资源
func GetInventory(c *gin.Context) {
	var TreeRes model.TreeRes
	info, err := model.GetInventory()
	if err != nil {
		TreeRes.Code = 500
		TreeRes.Message = err.Error()
	} else {
		TreeRes.Code = 200
		TreeRes.Message = "获取成功"
		TreeRes.Data.Items = info
	}
	c.JSON(http.StatusOK, TreeRes)
}

// GetOverview 获取汇总状态
func GetOverview(c *gin.Context) {
	var OverRes model.OverviewRes
	info, err := model.GetOverviewData()
	if err != nil {
		OverRes.Code = 500
		OverRes.Message = err.Error()
	} else {
		OverRes.Code = 200
		OverRes.Message = "获取成功"
		OverRes.Data.Items = info
	}
	c.JSON(http.StatusOK, OverRes)
}

// GetEgressData 获取出口带宽数据
func GetEgressData(c *gin.Context) {
	var EgrRes model.EgressRes
	info, err := model.GetEgressData()
	if err != nil {
		EgrRes.Code = 500
		EgrRes.Message = err.Error()
		EgrRes.Data.Items = info
	} else {
		EgrRes.Code = 200
		EgrRes.Message = "获取成功"
		EgrRes.Data.Items = info
	}
	c.JSON(http.StatusOK, EgrRes)
}

// GetVersion 获取版本信息
func GetVersion(c *gin.Context) {
	var VerRes model.VerRes
	VerRes.Code = 200
	VerRes.Message = "获取成功"
	VerRes.Data.Items.ZabbixVersion = model.ZBX_VER
	VerRes.Data.Items.Version = model.Version
	VerRes.Data.Items.GitHash = model.GitHash
	VerRes.Data.Items.BuildTime = model.BuildTime
	c.JSON(http.StatusOK, VerRes)
}

// GetZbxSession 获取Zabbix Session
func GetZbxSession(c *gin.Context) {
	var SesRes model.SessionRes
	session, err := model.GetZbxSession()
	if err != nil {
		SesRes.Code = 500
		SesRes.Message = err.Error()
	} else {
		SesRes.Code = 200
		SesRes.Message = "获取成功"
		SesRes.Data.Items = session
	}
	c.JSON(http.StatusOK, SesRes)
}
