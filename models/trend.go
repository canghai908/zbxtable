package models

import (
	"github.com/astaxie/beego/logs"
)

//GetTrendByItemID by itemid limit
func GetTrendByItemID(itemid string, limit string) ([]Trend, error) {
	par := []string{"itemid", "clock", "num", "value_min", "value_avg", "value_max"}
	par1 := []string{itemid}
	rep, err := API.Call("trend.get", Params{"output": par,
		"itemids": par1, "limit": limit})
	if err != nil {
		logs.Error(err)
		return []Trend{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Error(err)
		return []Trend{}, err
	}
	var hb []Trend
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return []Trend{}, err
	}
	return hb, err
}

//GetTrendData by itemid limit
func GetTrendData(itemid, timefrom, timetill string) ([]Trend, error) {
	output := []string{"itemid", "clock", "num", "value_min", "value_avg", "value_max"}
	itemids := []string{itemid}
	rep, err := API.Call("trend.get", Params{"output": output, "time_from": timefrom,
		"time_till": timetill, "itemids": itemids})
	if err != nil {
		logs.Error(err)
		return []Trend{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Error(err)
		return []Trend{}, err
	}
	var hb []Trend
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return []Trend{}, err
	}
	return hb, err
}

//GetTrendDataByItemid by itemid limit
func GetTrendDataByItemid(item Item, time_from, time_till int64) ([]Trend, error) {
	output := []string{"itemid", "clock", "num", "value_min", "value_avg", "value_max"}
	itemids := []string{item.Itemid}
	rep, err := API.CallWithError("trend.get",
		Params{"output": output,
			"time_from": time_from,
			"time_till": time_till,
			"itemids":   itemids})
	if err != nil {
		logs.Error(err)
		return []Trend{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Error(err)
		return []Trend{}, err
	}
	var hb []Trend
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return []Trend{}, err
	}
	return hb, err
}
