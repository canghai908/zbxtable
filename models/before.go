package models

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/canghai908/zbxtable/utils"
)

func MsFormat(token, version, tenantid string, message []byte) error {
	if token != beego.AppConfig.String("token") {
		return errors.New("Token Error!")
	}
	switch version {
	case "":

	case "v2", "V2":
		var mes EventTpl
		err := json.Unmarshal(message, &mes)
		if err != nil {
			logs.Error(err)
			return err
		}
		occtime, err := utils.ParTime(mes.EventTime)
		if err != nil {
			logs.Error(err)
			return err
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
			return err
		}
		logs.Info(mes)
		logs.Info(id)
	}
	return nil
}
