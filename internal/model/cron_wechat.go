package models

import (
	"bytes"
	"html/template"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"

	"github.com/xen0n/go-workwx"
)

// copy from open-facon-plus
func ConsumeWechat() {
	for {
		L := PopAllWechat()
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendWechatList(L)
	}
}
func SendWechatList(L []*Event) {
	for _, v := range L {
		WechatWorkerChan <- 1
		go SendWechat(v)
	}
}

func SendWechat(event *Event) {
	defer func() {
		<-WechatWorkerChan
	}()
	ids := strings.Split(event.ToUsers, ",")
	var plist []Manager
	err := DB.Where("id IN ?", ids).
		Select("id", "username", "email", "wechat", "phone", "ding_talk").
		Find(&plist).Error
	if err != nil {
		logger.Log.Error(err)
	}
	var tplname string
	if event.Status == "0" {
		tplname = "./template/wechat_recovery.tpl"
	} else {
		tplname = "./template/wechat_problem.tpl"
	}
	event.Level = utils.AlertSeverityTo(event.Level)
	event.Status = utils.AlertType(event.Status)
	tmpl, err := template.ParseFiles("./" + tplname)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	var body bytes.Buffer
	err = tmpl.Execute(&body, event)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	for _, v := range plist {
		SendWechatAlert(v, event, body.String())
	}
}
func SendWechatAlert(user Manager, event *Event, content string) error {
	tos := workwx.Recipient{
		UserIDs: []string{user.Wechat},
	}
	err := WeApp.SendTextMessage(&tos, content, false)
	var elog EventLog

	if err != nil {
		elog = EventLog{AlarmID: int64(event.ID), EventID: event.EventID,
			Rule: event.Rule, Channel: "wechat", User: user.Username, Account: user.Wechat,
			NotifyTime: time.Now(), NotifyContent: content,
			Status: strconv.Itoa(EventFailed), NotifyError: err.Error(),
		}

	} else {
		elog = EventLog{AlarmID: int64(event.ID), EventID: event.EventID,
			Rule: event.Rule, Channel: "wechat", User: user.Username, Account: user.Wechat,
			NotifyTime: time.Now(), NotifyContent: content,
			Status: strconv.Itoa(EventSuccess), NotifyError: "",
		}
	}
	//add event log
	_, err = AddEventLog(&elog)
	if err != nil {
		logger.Log.Error(err)
	}
	//update alalrm status
	return nil
}

func PopAllWechat() []*Event {
	ret := []*Event{}
	for {
		reply, err := CacheRPop("wechat")
		if err != nil {
			break
		}
		if reply == "" || reply == "nil" {
			break
		}

		var mail Event
		err = json.Unmarshal([]byte(reply), &mail)
		if err != nil {
			logger.Log.Error(err, reply)
			continue
		}
		ret = append(ret, &mail)
	}
	return ret
}
