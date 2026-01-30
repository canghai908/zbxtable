package handler

import (
	"net/http"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
)

// GetAllHostGroup 获取主机组列表
func GetAllHostGroup(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	groups := c.Query("groups")

	var HostGroupsRes model.HostGroupsList
	hs, cnt, err := model.GetAllHostGroups(page, limit, groups)
	if err != nil {
		HostGroupsRes.Code = 401
		HostGroupsRes.Message = err.Error()
		c.JSON(http.StatusOK, HostGroupsRes)
		return
	}
	HostGroupsRes.Code = 200
	HostGroupsRes.Message = "获取数据成功"
	HostGroupsRes.Data.Items = hs
	HostGroupsRes.Data.Total = cnt
	c.JSON(http.StatusOK, HostGroupsRes)
}

// GetAllHostGroupsList 获取所有主机组列表（树形）
func GetAllHostGroupsList(c *gin.Context) {
	var HostGroupsRes model.HostTreeList
	hs, cnt, err := model.GetAllHostGroupsList()
	if err != nil {
		HostGroupsRes.Code = 401
		HostGroupsRes.Message = err.Error()
	} else {
		HostGroupsRes.Code = 200
		HostGroupsRes.Message = "获取数据成功"
		HostGroupsRes.Data.Items = hs
		HostGroupsRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, HostGroupsRes)
}

// GetAllGroupsList 获取所有组列表
func GetAllGroupsList(c *gin.Context) {
	instanceID := c.Query("instance_id")
	
	var HostGroupsRes model.HostTreeList
	hs, cnt, err := model.GetAllGroupsListFromInstance(instanceID)
	if err != nil {
		HostGroupsRes.Code = 401
		HostGroupsRes.Message = err.Error()
	} else {
		HostGroupsRes.Code = 200
		HostGroupsRes.Message = "获取数据成功"
		HostGroupsRes.Data.Items = hs
		HostGroupsRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, HostGroupsRes)
}

// GetHostsByGroupID 根据组ID获取主机列表
func GetHostsByGroupID(c *gin.Context) {
	GroupID := c.Param("id")
	hs, err := model.GetHostsByGroupID(GroupID)
	var HostsByGroupIDRes model.HostGroupBYGroupIDList
	if err != nil {
		HostsByGroupIDRes.Code = 401
		HostsByGroupIDRes.Message = err.Error()
	} else {
		HostsByGroupIDRes.Code = 200
		HostsByGroupIDRes.Message = "获取数据成功"
	}
	HostsByGroupIDRes.Data.Items = hs
	c.JSON(http.StatusOK, HostsByGroupIDRes)
}
