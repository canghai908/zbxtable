package handler

import (
	"zbxtable/pkg/response"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"zbxtable/internal/model"
	"zbxtable/pkg/logger"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

// ExportTrend 导出趋势数据
func ExportTrend(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	var v model.ListQueryAll
	var ExpRes model.ExpList
	err = jsoniter.Unmarshal(body, &v)
	if err != nil {
		ExpRes.Code = 500
		ExpRes.Message = err.Error()
		c.JSON(http.StatusOK, ExpRes)
		return
	}
	var Start, End int64
	if len(v.Period) == 0 || v.Period[0] == "" || v.Period[1] == "" {
		tEnd := time.Now()
		End = tEnd.Unix()
		Start = tEnd.Add(-168 * time.Hour).Unix()
	}
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	st, _ := time.ParseInLocation(timeLayout, v.Period[0], loc)
	en, _ := time.ParseInLocation(timeLayout, v.Period[1], loc)
	Start = st.Unix()
	End = en.Unix()
	iodata, err := model.GetTrenDataFileName(v, Start, End)
	if err != nil {
		ExpRes.Code = 200
		ExpRes.Message = err.Error()
		c.JSON(http.StatusOK, ExpRes)
		return
	}
	oldfilename := v.Host.Name + "_" + v.Item.Name + "_trend" + ".xlsx"
	filename := url.QueryEscape(oldfilename)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment;filename="+filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Data(http.StatusOK, "application/octet-stream", iodata)
}

// ExportHistory 导出历史数据
func ExportHistory(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	var v model.ListQueryAll
	var ExpRes model.ExpList
	err = jsoniter.Unmarshal(body, &v)
	if err != nil {
		ExpRes.Code = 500
		ExpRes.Message = err.Error()
		c.JSON(http.StatusOK, ExpRes)
		return
	}
	var Start, End int64
	if len(v.Period) == 0 || v.Period[0] == "" || v.Period[1] == "" {
		tEnd := time.Now()
		End = tEnd.Unix()
		Start = tEnd.Add(-168 * time.Hour).Unix()
	}

	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	st, _ := time.ParseInLocation(timeLayout, v.Period[0], loc)
	en, _ := time.ParseInLocation(timeLayout, v.Period[1], loc)
	Start = st.Unix()
	End = en.Unix()

	iodata, err := model.GetHistoryDataFileName(v, Start, End)
	if err != nil {
		ExpRes.Code = 200
		ExpRes.Message = err.Error()
		c.JSON(http.StatusOK, ExpRes)
		return
	}
	oldfilename := v.Host.Name + "_" + v.Item.Name + "_history" + ".xlsx"
	filename := url.QueryEscape(oldfilename)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Data(http.StatusOK, "application/octet-stream", iodata)
}

// ExportInspect 导出巡检报告
func ExportInspect(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	var v model.HostGroupsPlist
	var ExpRes model.ExpList
	err = jsoniter.Unmarshal(body, &v)
	if err != nil {
		ExpRes.Code = 200
		ExpRes.Message = err.Error()
		c.JSON(http.StatusOK, ExpRes)
		return
	}
	hostdata, err := model.GetHostsByGroupIDList(v.GroupID)
	if err != nil {
		ExpRes.Code = 200
		ExpRes.Message = err.Error()
		c.JSON(http.StatusOK, ExpRes)
		return
	}
	type ss struct {
		V1 float64 `json:"v1"`
		V2 float64 `json:"v2"`
		V3 float64 `json:"v3"`
		V4 float64 `json:"v4"`
	}
	tt := make([]ss, len(hostdata))
	yy := make([]model.Insp, len(hostdata))
	for kk, v := range hostdata {
		b, err := model.GetItemByKey(v.HostID, "system.cpu.util[,idle]")
		if err != nil {
			logger.Log.Error(err)
			ExpRes.Code = 200
			ExpRes.Message = err.Error()
			c.JSON(http.StatusOK, ExpRes)
			return
		}
		mem1, err := model.GetItemByKey(v.HostID, "vm.memory.size[total]")
		if err != nil {
			logger.Log.Error(err)
			ExpRes.Code = 200
			ExpRes.Message = err.Error()
			c.JSON(http.StatusOK, ExpRes)
			return
		}
		mem2, err := model.GetItemByKey(v.HostID, "vm.memory.size[available]")
		if err != nil {
			logger.Log.Error(err)
			ExpRes.Code = 200
			ExpRes.Message = err.Error()
			c.JSON(http.StatusOK, ExpRes)
			return
		}
		for _, v := range mem1 {
			vint64, _ := strconv.ParseFloat(v.Lastvalue, 64)
			tt[kk].V1 = vint64
		}
		for _, v := range mem2 {
			vint64, _ := strconv.ParseFloat(v.Lastvalue, 64)
			tt[kk].V2 = vint64
		}
		for _, v := range b {
			vint64, _ := strconv.ParseFloat(v.Lastvalue, 64)
			if vint64 != 0 {
				tt[kk].V3 = model.Round(100-vint64, 2)
			} else {
				tt[kk].V3 = 0
			}
			if tt[kk].V1 != 0 {
				tt[kk].V4 = model.Round(tt[kk].V2/tt[kk].V1, 2)
			} else {
				tt[kk].V4 = 0
			}
		}
		yy[kk].HostName = v.Name
		yy[kk].CPULoad = tt[kk].V3
		yy[kk].MemPct = tt[kk].V4
	}
	ByteData, err := model.ExpInspect(v.Name, yy)
	if err != nil {
		ExpRes.Code = 200
		ExpRes.Message = err.Error()
		c.JSON(http.StatusOK, ExpRes)
		return
	}
	oldfilename := v.Name + ".xlsx"
	filename := url.QueryEscape(oldfilename)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Data(http.StatusOK, "application/octet-stream", ByteData)
}

// ExportHosts 导出设备列表
func ExportHosts(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	var v model.ExportHosts
	var HostRes model.HostList
	err = jsoniter.Unmarshal(body, &v)
	if err != nil {
		HostRes.Code = 500
		HostRes.Message = err.Error()
		c.JSON(http.StatusOK, HostRes)
		return
	}
	hs, err := model.GetHostList(v.Hosttype, v.Hosts, v.Model, v.Ip, v.Available)
	if err != nil {
		HostRes.Code = 500
		HostRes.Message = err.Error()
		c.JSON(http.StatusOK, HostRes)
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=host_list.xlsx")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Data(http.StatusOK, "application/octet-stream", hs)
}

// ExportInventory 导出资产列表
func ExportInventory(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	var v model.ExportInventory
	var HostRes model.HostList
	err = jsoniter.Unmarshal(body, &v)
	if err != nil {
		HostRes.Code = 500
		HostRes.Message = err.Error()
		c.JSON(http.StatusOK, HostRes)
		return
	}
	hs, err := model.GetInventoryInfo(v.HostType)
	if err != nil {
		HostRes.Code = 500
		HostRes.Message = err.Error()
		c.JSON(http.StatusOK, HostRes)
		return
	}
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=inventoryExport.xlsx")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Data(http.StatusOK, "application/octet-stream", hs)
}
