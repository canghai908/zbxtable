package handler

import (
	"io"
	"net/http"
	"strconv"
	"time"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

// GetGraphByHostID 根据Hostid查看主机图形
func GetGraphByHostID(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	var v model.GraphListQuery
	var GraphRes model.GraphList
	err = jsoniter.Unmarshal(body, &v)
	if err != nil {
		GraphRes.Code = 200
		GraphRes.Message = err.Error()
		c.JSON(http.StatusOK, GraphRes)
		return
	}

	var Start, End int64
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	if len(v.Period) == 0 || v.Period[0] == "" || v.Period[1] == "" {
		tEnd := time.Now()
		End = tEnd.Unix()
		Start = tEnd.Add(-168 * time.Hour).Unix()
	} else {
		st, _ := time.ParseInLocation(timeLayout, v.Period[0], loc)
		en, _ := time.ParseInLocation(timeLayout, v.Period[1], loc)
		Start = st.Unix()
		End = en.Unix()
	}

	id, _ := strconv.Atoi(v.Hostid)
	vv, count, err := model.GetGraphByHostID(id, Start, End)
	if err != nil {
		GraphRes.Code = 500
		GraphRes.Message = "获取图形数据错误"
		c.JSON(http.StatusOK, GraphRes)
		return
	}
	GraphRes.Code = 200
	GraphRes.Message = "获取数据成功"
	GraphRes.Data.Items = vv
	GraphRes.Data.Total = count
	c.JSON(http.StatusOK, GraphRes)
}
