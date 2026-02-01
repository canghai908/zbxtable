package handler

import (
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetAllHostGroup 获取主机组列表
func GetAllHostGroup(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	groups := c.Query("groups")
	hs, cnt, err := model.GetAllHostGroups(page, limit, groups)
	if err != nil {
		response.DatabaseError(c, "获取主机组列表失败: "+err.Error())
		return
	}
	response.SuccessWithPage(c, hs, cnt)
}

// GetAllHostGroupsList 获取所有主机组列表（树形）
func GetAllHostGroupsList(c *gin.Context) {
	hs, cnt, err := model.GetAllHostGroupsList()
	if err != nil {
		response.DatabaseError(c, "获取主机组列表失败: "+err.Error())
		return
	}
	response.SuccessWithPage(c, hs, cnt)
}

// GetAllGroupsList 获取所有组列表
func GetAllGroupsList(c *gin.Context) {
	zidStr := c.Query("zid")

	hs, cnt, err := model.GetAllGroupsListFromInstance(zidStr)
	if err != nil {
		response.DatabaseError(c, "获取组列表失败: "+err.Error())
		return
	}
	response.SuccessWithPage(c, hs, cnt)
}

// GetHostsByGroupID 根据组ID获取主机列表
func GetHostsByGroupID(c *gin.Context) {
	GroupID := c.Param("id")
	hs, err := model.GetHostsByGroupID(GroupID)
	if err != nil {
		response.DatabaseError(c, "获取主机列表失败: "+err.Error())
		return
	}
	response.Success(c, map[string]interface{}{
		"items": hs,
	})
}
