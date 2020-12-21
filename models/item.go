package models

import (
	"github.com/astaxie/beego/logs"
)

//GetItemByKey bye key
func GetItemByKey(hostid, key string) (item []Item, err error) {
	par := make(map[string]string)
	par["key_"] = key
	rep, err := API.Call("item.get", Params{"output": "extend",
		"sortfield": "name", "limit": "1", "hostids": hostid, "search": par})

	if err != nil {
		logs.Error(err)
		return []Item{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Error(err)
		return []Item{}, err
	}

	var hb []Item

	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return []Item{}, err
	}
	// //fmt.Println(hb.Host)

	return hb, err
}

//GetAllItemByHostID func
func GetAllItemByHostID(hostid string) (item []Item, count int64, err error) {
	output := []string{"itemid", "name", "key_", "value_type", "units"}
	rep, err := API.Call("item.get", Params{"output": output, "sortfield": "name",
		"hostids": hostid})

	if err != nil {
		logs.Error(err)
		return []Item{}, 0, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Error(err)
		return []Item{}, 0, err
	}

	var hb []Item
	//	log.Println(string(hba))
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return []Item{}, 0, err
	}
	return hb, int64(len(hb)), err
}
