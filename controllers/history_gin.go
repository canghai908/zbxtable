package controllers

import (
	"io"
	"net/http"
	"time"
	"zbxtable/models"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

// GetHistoryByItemID 获取历史数据
func GetHistoryByItemID(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
		return
	}
	
	var v models.HistoryQuery
	var HistoryRes models.HistoryList
	err = jsoniter.Unmarshal(body, &v)
	if err != nil {
		HistoryRes.Code = 500
		HistoryRes.Message = err.Error()
		c.JSON(http.StatusOK, HistoryRes)
		return
	}
	var Start, End int64
	if len(v.Period) == 0 || v.Period[0] == "" || v.Period[1] == "" {
		tEnd := time.Now()
		End = tEnd.Unix()
		Start = tEnd.Add(-10 * time.Minute).Unix()
	}
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	st, _ := time.ParseInLocation(timeLayout, v.Period[0], loc)
	en, _ := time.ParseInLocation(timeLayout, v.Period[1], loc)
	Start = st.Unix()
	End = en.Unix()
	his, err := models.GetHistoryByItemID(v.Itemids, v.History, Start, End)
	if err != nil {
		HistoryRes.Code = 500
		HistoryRes.Message = err.Error()
		c.JSON(http.StatusOK, HistoryRes)
		return
	}
	HistoryRes.Code = 200
	HistoryRes.Message = "获取成功"
	HistoryRes.Data.Items = his
	c.JSON(http.StatusOK, HistoryRes)
}
