package models

import (
	"github.com/astaxie/beego/logs"
)

//GetHistoryByItemID by id
func GetHistoryByItemID(itemid, history string, limit string) ([]History, error) {
	// par := make(map[string]string)
	// par["key_"] = key
	rep, err := API.Call("history.get", Params{"output": "extend",
		"itemids": itemid, "history": history, "sortfield": "clock",
		"sortorder": "DESC", "limit": limit})

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

//GetHistoryByItemIDNew by id
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

//GetHistoryByItemIDNewP bg
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
