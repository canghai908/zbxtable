package handler

import (
	"io"
	"net/http"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// GetTopoDataByID 获取拓扑数据
func GetTopoDataByID(c *gin.Context) {
	idStr := c.Param("id")
	var TopologyRes model.TopologyInfo
	TopologyRes.Code = 200
	TopologyRes.Message = idStr
	c.JSON(http.StatusOK, TopologyRes)
}

// CreateTopoData 创建数据采集点
func CreateTopoData(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}

	pid := gjson.Get(string(body), "pid").String()
	vtype := gjson.Get(string(body), "v_type").String()
	ptype := gjson.Get(string(body), "type").String()
	tid := gjson.Get(string(body), "tid").String()
	var TopologyRes model.TopologyList
	v := model.TopologyData{PID: pid, Type: ptype, VType: vtype, TID: tid}
	_, err = model.AddTopoData(&v)
	if err != nil {
		TopologyRes.Code = 500
		TopologyRes.Message = err.Error()
	} else {
		TopologyRes.Code = 200
		TopologyRes.Message = "创建成功"
	}
	c.JSON(http.StatusOK, TopologyRes)
}
