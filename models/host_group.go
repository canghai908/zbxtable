package models

import (
	"encoding/json"
	"log"
	"math"
	"strconv"
	"strings"
)

//GetAllHostGroups func
func GetAllHostGroups(page, limit, groups string) ([]HostGroups, int64, error) {
	rep, err := API.Call("hostgroup.get", Params{"output": "extend",
		"selectHosts": "count"})
	if err != nil {
		return []HostGroups{}, 0, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []HostGroups{}, 0, err
	}
	var hb []HostGroups
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []HostGroups{}, 0, err
	}
	var dt []HostGroups
	var d HostGroups
	if groups != "" {
		for _, v := range hb {
			if strings.Contains(v.Name, groups) {
				d.GroupID = v.GroupID
				d.Name = v.Name
				d.Hosts = v.Hosts
				d.Internal = v.Internal
				dt = append(dt, d)
			}
		}
	} else {
		for _, v := range hb {
			if strings.Contains(v.Name, groups) {
				d.GroupID = v.GroupID
				d.Name = v.Name
				d.Hosts = v.Hosts
				d.Internal = v.Internal
				dt = append(dt, d)
			}
		}
	}
	//页数
	IntPage, err := strconv.Atoi(page)
	if err != nil {
		IntPage = 1
	}
	//每页数量
	IntLimit, err := strconv.Atoi(limit)
	if err != nil {
		IntLimit = 10
	}
	// //fmt.Println(hb.Host)
	//如果dt为空直接返回
	if len(dt) == 0 {
		return dt, int64(len(dt)), err
	}
	//分页
	nums := len(dt)

	//page总数
	totalpages := int(math.Ceil(float64(nums) / float64(IntLimit)))
	if IntPage >= totalpages {
		IntPage = totalpages
	}
	if IntPage <= 0 {
		IntPage = 1
	}
	//结束页数据
	var end int
	//begin 开始页数据

	begin := (IntPage - 1) * IntLimit
	if IntPage == totalpages {
		end = nums
	}
	if IntPage < totalpages {
		end = IntPage * IntLimit
	} else {
		end = nums
	}
	var newgroups []HostGroups
	for i := begin; i < end; i++ {
		newgroups = append(newgroups, dt[i])
	}
	return newgroups, int64(len(dt)), err
}

//GetAllHostGroupsList func
func GetAllHostGroupsList() ([]HostTree, int64, error) {
	selectHosts := []string{"hostid", "name", "status"}
	rep, err := API.Call("hostgroup.get", Params{"output": "extend",
		"selectHosts": selectHosts})
	if err != nil {
		return []HostTree{}, 0, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []HostTree{}, 0, err
	}
	var hb []HostTree

	err = json.Unmarshal(hba, &hb)
	if err != nil {
		log.Println(err)
		return []HostTree{}, 0, err
	}
	return hb, int64(len(hb)), err
}

//GetHostsInfoByGroupID func
func GetHostsInfoByGroupID(GroupID string) ([]HostGroupBYGroupID, error) {
	output := []string{"groupid", "name"}
	selectHosts := []string{"hostid", "name", "status"}
	selectInterfaces := []string{"main", "port"}
	rep, err := API.Call("hostgroup.get", Params{"output": output,
		"groupids": GroupID, "selectHosts": selectHosts,
		"selectInterfaces": selectInterfaces})

	if err != nil {
		log.Fatalln(err)
		return []HostGroupBYGroupID{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		log.Fatalln(err)
		return []HostGroupBYGroupID{}, err
	}
	var hb []HostGroupBYGroupID
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		log.Fatalln(err)
		return []HostGroupBYGroupID{}, err
	}
	return hb, err
}

//GetHostsByGroupID func
func GetHostsByGroupID(GroupID string) ([]HostGroupBYGroupID, error) {
	output := []string{"groupid", "name"}
	selectHosts := []string{"hostid", "name", "status"}
	rep, err := API.Call("hostgroup.get", Params{"output": output,
		"groupids": GroupID, "selectHosts": selectHosts})

	if err != nil {
		log.Fatalln(err)
		return []HostGroupBYGroupID{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		log.Fatalln(err)
		return []HostGroupBYGroupID{}, err
	}

	var hb []HostGroupBYGroupID

	err = json.Unmarshal(hba, &hb)
	if err != nil {
		log.Fatalln(err)
		return []HostGroupBYGroupID{}, err
	}
	return hb, err
}

//GetHostsByGroupIDList func
func GetHostsByGroupIDList(GroupID string) ([]Hosts, error) {
	output := []string{"groupid", "name"}
	selectHosts := []string{"hostid", "name", "status"}
	rep, err := API.Call("hostgroup.get", Params{"output": output,
		"groupids": GroupID, "selectHosts": selectHosts})
	if err != nil {
		return []Hosts{}, err
	}
	if err != nil {
		return []Hosts{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []Hosts{}, err
	}
	var hb []HostGroupBYGroupID
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []Hosts{}, err
	}
	var b Hosts
	var list []Hosts
	for _, v := range hb {
		for _, vv := range v.Hosts {
			b.HostID = vv.HostID
			b.Name = vv.Name
			b.Status = vv.Status
			list = append(list, b)
		}
	}
	return list, nil
}
