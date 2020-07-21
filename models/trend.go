package models

import (
	"encoding/json"
	"log"
)

//GetTrendByItemID by itemid limit
func GetTrendByItemID(itemid string, limit string) ([]Trend, error) {
	par := []string{"itemid", "clock", "num", "value_min", "value_avg", "value_max"}
	par1 := []string{itemid}
	rep, err := API.Call("trend.get", Params{"output": par,
		"itemids": par1, "limit": limit})
	if err != nil {
		log.Fatalln(err)
		return []Trend{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		log.Fatalln(err)
		return []Trend{}, err
	}
	var hb []Trend
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
		return []Trend{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		log.Fatalln(err)
		return []Trend{}, err
	}
	var hb []Trend
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		log.Fatalln(err)
		return []Trend{}, err
	}
	return hb, err
}

//GetTrendDataByItemid by itemid limit
func GetTrendDataByItemid(itemid, timefrom, timetill string) ([]Trend, error) {
	output := []string{"itemid", "clock", "num", "value_min", "value_avg", "value_max"}
	itemids := []string{itemid}
	rep, err := API.CallWithError("trend.get", Params{"output": output, "time_from": timefrom,
		"time_till": timetill, "itemids": itemids})
	if err != nil {
		log.Fatalln(err)
		return []Trend{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		log.Fatalln(err)
		return []Trend{}, err
	}
	var hb []Trend
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		log.Fatalln(err)
		return []Trend{}, err
	}
	return hb, err
}
