package models

import (
	"github.com/astaxie/beego/logs"
)

//GetApplicationByHostid st
func GetApplicationByHostid(hostid string) ([]Application, int64, error) {
	exp := []string{"applicationid", "name"}
	output := []string{"itemid", "name", "value_type", "units"}
	rep, err := API.Call("application.get", Params{"output": exp,
		"hostids": hostid, "selectItems": output})
	if err != nil {
		return []Application{}, 0, err
	}
	//log.Println(rep.Result)
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Error(err)
		return []Application{}, 0, err
	}

	var hb []Application
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return []Application{}, 0, err
	}
	return hb, int64(len(hb)), err
}
