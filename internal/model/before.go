package model

import (
	"fmt"
	"strings"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"
)

func MsAdd(zid int, message []byte) (int64, error) {
	//replace " \
	p0 := strings.Replace(string(message), `\`, `\\`, -1)
	//替换"
	p1 := strings.Replace(p0, `"`, `\"`, -1)
	p2 := strings.ReplaceAll(p1, `¦`, `"`)
	var mes EventTpl
	err := json.Unmarshal([]byte(p2), &mes)
	if err != nil {
		fmt.Println("CCCC")
		logger.Log.Error(err)
		return 0, err
	}
	occurTime, err := utils.ParTime(mes.EventTime)
	if err != nil {
		logger.Log.Error(err)
		return 0, err
	}

	// 创建告警记录，只保存 ZID
	var meal = Alarm{
		ZID:           zid,
		HostID:        mes.HostsID,
		Hostname:      mes.Hostname,
		Host:          mes.HostHost,
		HostsIP:       mes.HostsIP,
		TriggerID:     mes.TriggerID,
		ItemID:        mes.ItemID,
		ItemName:      mes.ItemName,
		ItemValue:     mes.ItemValue,
		Hgroup:        mes.HostGroup,
		OccurTime:     occurTime,
		Level:         mes.Severity,
		Message:       mes.TriggerName,
		Hkey:          mes.TriggerKey,
		Detail:        mes.ItemName + ":" + mes.ItemValue,
		Status:        mes.TriggerValue,
		EventID:       mes.EventID,
		EventDuration: mes.EventDuration,
	}
	id, err := AddAlarm(&meal)
	if err != nil {
		logger.Log.Error(err)
		return 0, err
	}
	//alert gen
	GenAlert(&meal)
	return id, nil
}
