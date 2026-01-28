package models

import (
	"time"
)

//GetGraphByHostID by id
func GetGraphByHostID(hostid int, start, end int64) ([]GraphInfo, int64, error) {
	rep, err := API.CallWithError("graph.get", Params{"output": "extend",
		"hostids": hostid, "sortfiled": "name"})
	if err != nil {
		return []GraphInfo{}, 0, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []GraphInfo{}, 0, err
	}
	EndTime := time.Unix(end, 0).Format("2006-01-02 15:04:05")
	StartTime := time.Unix(start, 0).Format("2006-01-02 15:04:05")
	var hb []GraphInfo
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []GraphInfo{}, 0, err
	}
	var bb GraphInfo
	var cc []GraphInfo
	for _, v := range hb {
		bb.GraphID = "/v1/images/" + v.GraphID + "?from=" + StartTime + "?to=" + EndTime
		bb.Name = v.Name
		cc = append(cc, bb)
	}
	return cc, int64(len(hb)), err
}
