package handler

import (
	"io"
	"net/http"
	"strconv"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

func GetAllAssetGroups(c *gin.Context) {
	list, err := model.GetAllAssetGroups()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, list)
}

func GetAssetGroupByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	g, err := model.GetAssetGroupByID(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, g)
}

func CreateAssetGroup(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}
	name := gjson.Get(string(body), "name").String()
	groupKey := gjson.Get(string(body), "group_key").String()
	if name == "" || groupKey == "" {
		response.BadRequest(c, "名称和分组标识不能为空")
		return
	}
	g := &model.AssetGroup{
		Name:      name,
		GroupKey:  groupKey,
		Icon:      gjson.Get(string(body), "icon").String(),
		SortOrder: int(gjson.Get(string(body), "sort_order").Int()),
		IsBuiltin: false,
	}
	if err := model.CreateAssetGroup(g); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", g)
}

func UpdateAssetGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}
	name := gjson.Get(string(body), "name").String()
	if name == "" {
		response.BadRequest(c, "名称不能为空")
		return
	}
	g := &model.AssetGroup{
		ID:        id,
		Name:      name,
		Icon:      gjson.Get(string(body), "icon").String(),
		SortOrder: int(gjson.Get(string(body), "sort_order").Int()),
	}
	if err := model.UpdateAssetGroup(g); err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "更新成功", g)
}

func DeleteAssetGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的ID")
		return
	}
	if err := model.DeleteAssetGroup(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
