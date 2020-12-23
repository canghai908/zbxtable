package models

import (
	"strconv"
)

//GetCountHost func
func GetCountHost() (info IndexInfo, err error) {
	hosts, err := API.CallWithError("host.get", Params{"output": "extend",
		"countOutput": true})
	if err != nil {
		return IndexInfo{}, err
	}
	items, err := API.CallWithError("item.get", Params{"output": "extend",
		"countOutput": true})
	if err != nil {
		return IndexInfo{}, err
	}
	problems, err := API.CallWithError("problem.get", Params{"output": "extend",
		"countOutput": true,
	})
	if err != nil {
		return IndexInfo{}, err
	}
	triggers, err := API.CallWithError("trigger.get", Params{"output": "extend",
		"countOutput": true})
	if err != nil {
		return IndexInfo{}, err
	}
	d := IndexInfo{}
	hostsint64, _ := strconv.ParseInt(hosts.Result.(string), 10, 64)
	d.Hosts = hostsint64
	itemsint64, _ := strconv.ParseInt(items.Result.(string), 10, 64)
	d.Items = itemsint64
	problemsint64, _ := strconv.ParseInt(problems.Result.(string), 10, 64)
	d.Problems = problemsint64
	triggersint64, _ := strconv.ParseInt(triggers.Result.(string), 10, 64)
	d.Triggers = triggersint64
	return d, nil
}
