package models

import (
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/canghai908/zbxtable/utils"
	"strings"
)

func MsAdd(token, tenantid string, message []byte) (int64, error) {
	if token != GetConfKey("token") {
		return 0, errors.New("Token Error!")
	}
	//replace "
	p1 := strings.ReplaceAll(string(message), `"`, `\"`)
	p2 := strings.ReplaceAll(p1, `¦`, `"`)
	var mes EventTpl
	err := json.Unmarshal([]byte(p2), &mes)
	if err != nil {
		logs.Error(err)
		return 0, err
	}
	occtime, err := utils.ParTime(mes.EventTime)
	if err != nil {
		logs.Error(err)
		return 0, err
	}
	var meal = Alarm{
		TenantID:  tenantid,
		HostID:    mes.HostsID,
		Hostname:  mes.Hostname,
		Host:      mes.HostHost,
		HostsIP:   mes.HostsIP,
		TriggerID: mes.TriggerID,
		ItemID:    mes.ItemID,
		ItemName:  mes.ItemName,
		ItemValue: mes.ItemValue,
		Hgroup:    mes.HostGroup,
		Occurtime: occtime,
		Level:     mes.Severity,
		Message:   mes.TriggerName,
		Hkey:      mes.TriggerKey,
		Detail:    mes.ItemName + ":" + mes.ItemValue,
		Status:    mes.TriggerValue,
		EventID:   mes.EventID,
	}
	id, err := AddAlarm(&meal)
	if err != nil {
		logs.Error(err)
		return 0, err
	}
	return id, nil
}
