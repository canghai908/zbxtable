package models

import (
	"context"
	"github.com/astaxie/beego/logs"
	"strings"
	"time"
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
	return hb, err
}

//GetAllItemByHostID func
func GetAllItemByHostID(hostid string) (item []Item, count int64, err error) {
	output := []string{"itemid", "name", "key_", "value_type", "units", "hostid"}
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

//GetAllItemByHostID func
func GetAllTrafficeItemByHostID(hostid string) (item []interface{}, count int64, err error) {
	var ItemList []interface{}
	if ZBX_V {
		ItemsOutput := []string{"itemid", "tags", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
		selectTags := []string{"tag", "value"}
		Search2Par := make(map[string]string, 1)
		Search2Par["tag"] = "interface"
		Search2Par["value"] = ""
		Par := make(map[int]interface{})
		Par[0] = Search2Par
		rep1, err := API.CallWithError("item.get", Params{
			"output":     ItemsOutput,
			"hostids":    hostid,
			"selectTags": selectTags,
			"sortfield":  "name",
			"tags":       Par})
		if err != nil {
			return ItemList, 0, err
		}
		ApplicationResByte, err := json.Marshal(rep1.Result)
		if err != nil {
			return ItemList, 0, err
		}
		var ts []MonIts
		err = json.Unmarshal(ApplicationResByte, &ts)
		if err != nil {
			return ItemList, 0, err
		}
		for _, v := range ts {
			for _, vv := range v.Tags {
				if vv.Tag == "interface" {
					switch {
					case strings.Contains(v.Name, "Bits received"):
						ItemList = append(ItemList, v)
					case strings.Contains(v.Name, "Bits sent"):
						ItemList = append(ItemList, v)
					}
				}
			}
		}
		return ItemList, int64(len(ItemList)), nil
	}
	ItemsOutput := []string{"itemid", "tags", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
	rep, err := API.Call("item.get", Params{"output": ItemsOutput, "" +
		"sortfield": "name",
		"hostids": hostid,
	})
	if err != nil {
		logs.Error(err)
		return ItemList, 0, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Error(err)
		return ItemList, 0, err
	}
	var hb []Item
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return ItemList, 0, err
	}
	for _, v := range hb {
		switch {
		case strings.Contains(v.Name, "Bits received"):
			ItemList = append(ItemList, v)
		case strings.Contains(v.Name, "Bits sent"):
			ItemList = append(ItemList, v)
		}
	}
	return ItemList, int64(len(ItemList)), nil
}

//GetAllItemByHostID func
func GetReceiveTrafficeItemByHostID(hostid string) (item []interface{}, count int64, err error) {
	var ItemList []interface{}
	if ZBX_V {
		ItemsOutput := []string{"itemid", "tags", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
		selectTags := []string{"tag", "value"}
		Search2Par := make(map[string]string, 1)
		Search2Par["tag"] = "interface"
		Search2Par["value"] = ""
		Par := make(map[int]interface{})
		Par[0] = Search2Par
		rep1, err := API.CallWithError("item.get", Params{
			"output":     ItemsOutput,
			"hostids":    hostid,
			"selectTags": selectTags,
			"sortfield":  "name",
			"tags":       Par})
		if err != nil {
			return ItemList, 0, err
		}
		ApplicationResByte, err := json.Marshal(rep1.Result)
		if err != nil {
			return ItemList, 0, err
		}
		var ts []MonIts
		err = json.Unmarshal(ApplicationResByte, &ts)
		if err != nil {
			return ItemList, 0, err
		}
		for _, v := range ts {
			for _, vv := range v.Tags {
				if vv.Tag == "interface" {
					switch {
					case strings.Contains(v.Name, "Bits received"):
						v.Name = vv.Value
						ItemList = append(ItemList, v)
						//case strings.Contains(v.Name, "Bits sent"):
						//	ItemList = append(ItemList, v)
					}
				}
			}
		}
		return ItemList, int64(len(ItemList)), nil
	}
	//old version
	ItemsOutput := []string{"itemid", "tags", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
	rep, err := API.Call("item.get", Params{"output": ItemsOutput, "" +
		"sortfield": "name",
		"hostids": hostid,
	})
	if err != nil {
		logs.Error(err)
		return ItemList, 0, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Error(err)
		return ItemList, 0, err
	}
	var hb []Item
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return ItemList, 0, err
	}
	for _, v := range hb {
		switch {
		case strings.Contains(v.Name, "Bits received"):
			ItemList = append(ItemList, v)
		}
	}
	return ItemList, int64(len(ItemList)), nil
}

//GetFlowItemByHostID func
func GetFlowItemByHostID(hostid string) (item []Item, count int64, err error) {
	output := []string{"itemid", "name", "key_", "value_type", "units"}
	rep, err := API.Call("item.get", Params{"output": output, "" +
		"sortfield": "name",
		"hostids": hostid,
	})
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

//GetAllItemByHostID func
func GetItemByID(itemid string) (item []Item, err error) {
	if itemid == "" {
		return []Item{}, err
	}
	output := []string{"itemid", "name", "key_", "value_type", "units", "lastvalue", "lastclock", "hostid"}
	rep, err := API.Call("item.get", Params{"output": output, "sortfield": "name",
		"itemids": itemid})
	if err != nil {
		logs.Debug(err)
		return []Item{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Debug(err)
		return []Item{}, err
	}

	var hb []Item
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Debug(err)
		return []Item{}, err
	}
	return hb, err
}

//GetValueMap
func GetValueMapByID(id, value string) (newvalue string, err error) {
	if id == "" {
		return "", err
	}
	var ctx = context.Background()
	newvalue, err = RDB.Get(ctx, "ValueMap_"+id+"_"+value).Result()
	if err != nil {
		rep, err := API.Call("valuemap.get", Params{
			"output":         "extend",
			"selectMappings": "extend",
			"valuemapids":    id})
		if err != nil {
			logs.Debug(err)
			return "", err
		}
		hba, err := json.Marshal(rep.Result)
		if err != nil {
			logs.Debug(err)
			return "", err
		}
		var hb []ValueMap
		err = json.Unmarshal(hba, &hb)
		if err != nil {
			logs.Debug(err)

		}
		for _, v := range hb[0].Mappings {
			err = RDB.Set(ctx, "ValueMap_"+id+"_"+v.Value, v.Newvalue, -1*time.Second).Err()
			if err != nil {
				logs.Error(err)
				continue
			}
		}
		newvalue, err = RDB.Get(ctx, "ValueMap_"+id+"_"+value).Result()
		if err != nil {
			return "", err
		}
		return newvalue, nil
	}
	return newvalue, nil

}

//GetItemByItemids
func GetItemByIDS(itemids []string) (item []Item, err error) {
	if len(itemids) == 0 {
		return []Item{}, err
	}
	output := []string{"itemid", "name", "key_", "value_type", "units", "lastvalue", "lastclock", "hostid"}
	rep, err := API.Call("item.get", Params{"output": output, "sortfield": "name",
		"itemids": itemids})
	if err != nil {
		logs.Debug(err)
		return []Item{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Debug(err)
		return []Item{}, err
	}

	var hb []Item
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Debug(err)
		return []Item{}, err
	}
	return hb, err
}
