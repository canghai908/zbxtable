package models

import (
	"github.com/astaxie/beego/logs"
	"sort"
	"strconv"
	"time"
)

// GetHistoryByItemID
func GetHistoryByItemID(itemid, history string, time_from, time_till int64) ([]History, error) {
	rep, err := API.CallWithError("history.get",
		Params{"output": "extend",
			"itemids":   itemid,
			"history":   history,
			"sortfield": "clock",
			"sortorder": "DESC",
			"time_from": time_from,
			"time_till": time_till})
	if err != nil {
		return []History{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []History{}, err
	}
	var hb []History
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []History{}, err
	}
	return hb, err
}

// GetHistoryByItemID
func GetHistoryByItemIDTTTT(itemid, history string, time_from, time_till int64) ([]History, error) {
	rep, err := API.CallWithError("history.get",
		Params{"output": "extend",
			"itemids":   itemid,
			"history":   history,
			"sortfield": "clock",
			"sortorder": "ASC",
			"time_from": time_from,
			"time_till": time_till})
	if err != nil {
		return []History{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []History{}, err
	}
	var hb []History
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []History{}, err
	}
	return hb, err
}

// GetHistoryByItemIDNew by id
func GetHistoryByItemIDNew(item Item, time_from, time_till int64) ([]History, error) {
	rep, err := API.Call("history.get",
		Params{"output": "extend",
			"itemids":   item.Itemid,
			"history":   item.ValueType,
			"sortfield": "clock",
			"sortorder": "DESC",
			"time_from": time_from,
			"time_till": time_till})
	if err != nil {
		return []History{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Error(err)
		return []History{}, err
	}

	var hb []History
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return []History{}, err
	}
	return hb, err
}

// GetHistoryByItemIDNewP bg
func GetHistoryByItemIDNewP(itemid, TimeFrom, TimeTill int64) ([]History, error) {
	rep, err := API.Call("history.get", Params{"output": "extend",
		"itemids": itemid, "history": "0", "sortfield": "clock",
		"sortorder": "DESC",
		"time_from": TimeFrom,
		"time_till": TimeTill})
	if err != nil {
		return []History{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Error(err)
		return []History{}, err
	}
	var hb []History
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return []History{}, err
	}
	return hb, err
}

// GetInterfaceGraphData 接口流量数据获取
func GetInterfaceGraphData(data InterfaceData) (series TrafficData, err error) {
	var timeFrom, timeTill int64
	loc, _ := time.LoadLocation("Asia/Shanghai")
	//时间处理
	if data.Begin != "" || data.End != "" {
		tBegin, err := time.ParseInLocation("2006-01-02 15:04:05", data.Begin, loc)
		if err != nil {
			timeFrom = time.Now().Add(-180 * time.Minute).Unix()
		} else {
			timeFrom = tBegin.Unix()
		}
		tEnd, err := time.ParseInLocation("2006-01-02 15:04:05", data.End, time.Local)
		if err != nil {
			timeTill = time.Now().Unix()
		} else {
			timeTill = tEnd.Unix()
		}
	}
	var itemList []string
	//traffic
	itemList = append(itemList, data.BitsReceivedItemId)
	itemList = append(itemList, data.BitsSentItemId)
	//discarded
	itemList = append(itemList, data.InDiscardedItemId)
	itemList = append(itemList, data.OutDiscardedItemId)
	//errors
	itemList = append(itemList, data.InErrorsItemId)
	itemList = append(itemList, data.OutErrorsItemId)
	//operation
	itemList = append(itemList, data.OperationalStatusItemId)
	rep, err := API.Call("history.get",
		Params{"output": "extend",
			"itemids":   itemList,
			"history":   data.BitsReceivedValueType,
			"sortfield": "clock",
			"sortorder": "ASC",
			"time_from": timeFrom,
			"time_till": timeTill})
	if err != nil {
		return TrafficData{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Error(err)
		return TrafficData{}, err
	}
	var hb []History
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return TrafficData{}, err
	}
	var date []string
	var BitsReceived, BitsSent, InDiscarded, OutDiscarded, InErrors,
		OutErrors, OperationalStatus []float64

	for _, v := range hb {
		switch v.Itemid {
		//BitsReceived
		case data.BitsReceivedItemId:
			tValue, _ := strconv.ParseFloat(v.Value, 64)
			BitsReceived = append(BitsReceived, tValue)
			t, _ := strconv.ParseInt(v.Clock, 10, 64)
			//时间统一
			date = append(date, time.Unix(t, 0).Format("01-02 15:04:05"))
			//BitsSent
		case data.BitsSentItemId:
			tValue, _ := strconv.ParseFloat(v.Value, 64)
			BitsSent = append(BitsSent, tValue)
		case data.InDiscardedItemId:
			tValue, _ := strconv.ParseFloat(v.Value, 64)
			InDiscarded = append(InDiscarded, tValue)
		case data.OutDiscardedItemId:
			tValue, _ := strconv.ParseFloat(v.Value, 64)
			OutDiscarded = append(OutDiscarded, tValue)
			//errors
		case data.InErrorsItemId:
			tValue, _ := strconv.ParseFloat(v.Value, 64)
			InErrors = append(InErrors, tValue)
		case data.OutErrorsItemId:
			tValue, _ := strconv.ParseFloat(v.Value, 64)
			OutErrors = append(OutErrors, tValue)
		//operation
		case data.OperationalStatusItemId:
			tValue, _ := strconv.ParseFloat(v.Value, 64)
			OperationalStatus = append(OperationalStatus, tValue)
		}
	}
	//空数据返回零
	if len(OperationalStatus) == 0 {
		length := len(BitsSent)
		for i := 0; i < length; {
			OperationalStatus = append(OperationalStatus, 0)
			i++
		}
	}
	var se TrafficData
	//流量数据
	var tfSeries TrafficSeries
	tfSeries.XAxis.Type = "category"
	tfSeries.XAxis.Data = date
	tfSeries.YAxis = append(tfSeries.YAxis, YAxis{
		Name: "接收流量",
		Data: BitsReceived,
		Type: "line",
		ItemStyle: ItemStyle{
			Opacity: 0.5,
			Color:   "rgb(0,204,0)",
		},
	}, YAxis{
		Name: "发送流量",
		Data: BitsSent,
		Type: "line",
		ItemStyle: ItemStyle{
			Opacity: 0.5,
			Color:   "rgb(0,0,225)",
		},
	})
	trafficLeg := []string{"接收流量", "发送流量"}
	//表格数据
	tfSeries.Legend.Data = trafficLeg
	avgRX, maxRx, minRx, perc95AvgRx, _ := GetFloat64ArrayData(BitsReceived)
	avgTx, maxTx, minTx, perc95AvgTx, _ := GetFloat64ArrayData(BitsSent)
	tfSeries.TrafficTable = append(tfSeries.TrafficTable, TrafficTable{
		Name:        "接收流量",
		Min:         minRx,
		Max:         maxRx,
		Avg:         avgRX,
		Th95PercAvg: perc95AvgRx,
		//Th95PercVal: perc95Val,
	}, TrafficTable{

		Name:        "发送流量",
		Min:         minTx,
		Max:         maxTx,
		Avg:         avgTx,
		Th95PercAvg: perc95AvgTx,
		//Th95PercVal: perc95Val,
	})
	se.TrafficSeries = tfSeries
	//丢弃包数据
	var disSeries DiscardedSeries
	disSeries.XAxis.Type = "category"
	disSeries.XAxis.Data = date
	disSeries.YAxis = append(disSeries.YAxis, YAxis{
		Name: "接收丢弃包",
		Data: InDiscarded,
		Type: "line",
	}, YAxis{
		Name: "发送丢弃包",
		Data: OutDiscarded,
		Type: "line",
	})
	disSeriesLeg := []string{"接收丢弃包", "发送丢弃包"}
	disSeries.Legend.Data = disSeriesLeg
	avgInDis, maxInDis, minInDis, _, _ := GetFloat64ArrayData(InDiscarded)
	avgOutDis, maxOutDis, minOutDis, _, _ := GetFloat64ArrayData(OutDiscarded)
	disSeries.Table = append(disSeries.Table, Table{
		Name: "接收丢弃包",
		Min:  minInDis,
		Max:  maxInDis,
		Avg:  avgInDis,
	}, Table{
		Name: "发送丢弃包",
		Min:  minOutDis,
		Max:  maxOutDis,
		Avg:  avgOutDis,
	})
	se.DiscardedSeries = disSeries
	//错包
	var errSeries ErrorsSeries
	errSeries.XAxis.Type = "category"
	errSeries.XAxis.Data = date
	errSeries.YAxis = append(errSeries.YAxis, YAxis{
		Name: "接收错包",
		Data: InErrors,
		Type: "line",
	}, YAxis{
		Name: "发送错包",
		Data: OutErrors,
		Type: "line",
	})
	errSeriesLeg := []string{"接收错包", "发送错包"}
	errSeries.Legend.Data = errSeriesLeg
	avgInErr, maxInErr, minInErr, _, _ := GetFloat64ArrayData(InErrors)
	avgOutErr, maxOutErr, minOutErr, _, _ := GetFloat64ArrayData(OutErrors)
	errSeries.Table = append(errSeries.Table, Table{
		Name: "接收错包",
		Min:  minInErr,
		Max:  maxInErr,
		Avg:  avgInErr,
	}, Table{
		Name: "发送错包",
		Min:  minOutErr,
		Max:  maxOutErr,
		Avg:  avgOutErr,
	})
	se.ErrorsSeries = errSeries
	//operation
	var opSeries OperationalStatusSeries
	opSeries.XAxis.Type = "category"
	opSeries.XAxis.Data = date
	opSeries.YAxis = append(opSeries.YAxis, YAxis{
		Name: "端口状态",
		Data: OperationalStatus,
		Type: "line",
		ItemStyle: ItemStyle{
			Opacity: 0.5,
			Color:   "rgb(0,255,0)",
		},
	})
	opSeriesLeg := []string{"端口状态"}
	opSeries.Legend.Data = opSeriesLeg
	avgOper, maxOper, minOper, _, _ := GetFloat64ArrayData(OperationalStatus)
	opSeries.Table = append(opSeries.Table, Table{
		Name: "端口状态",
		Min:  minOper,
		Max:  maxOper,
		Avg:  avgOper,
	})
	se.OperationalSeries = opSeries
	return se, nil
}

// GetFloat64ArrayData 计算float64数组的平均值，最大，最小，95%平均值，95%值，并保留3位小数
func GetFloat64ArrayData(data []float64) (avg, max, min, perc95Avg, perc95Val string) {
	//data拷贝到data1
	data1 := make([]float64, 0)
	data1 = append(data1, data[:]...)
	sort.Float64s(data1)
	var sum float64
	for _, v := range data1 {
		sum += v
	}
	avg = strconv.FormatFloat(sum/float64(len(data1)), 'f', 2, 64)
	max = strconv.FormatFloat(data1[len(data1)-1], 'f', 2, 64)
	min = strconv.FormatFloat(data1[0], 'f', 2, 64)
	perc95Avg = strconv.FormatFloat(sum/float64(len(data1))*0.95, 'f', 2, 64)
	perc95Val = strconv.FormatFloat(data1[int(float64(len(data1))*0.95)], 'f', 2, 64)
	return avg, max, min, perc95Avg, perc95Val
}
