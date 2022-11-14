package models

import (
	"bytes"
	"context"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	redis "github.com/go-redis/redis/v8"
	"github.com/xen0n/go-workwx"
	"html/template"
	"strconv"
	"strings"
	"time"
	template2 "zbxtable/utils"
)

//copy from open-facon-plus
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

//
func SendWechat(event *Event) {
	defer func() {
		<-WechatWorkerChan
	}()
	o := orm.NewOrm()
	ids := strings.Split(event.ToUsers, ",")
	var user Manager
	var plist []Manager
	_, err := o.QueryTable(user).Filter("id__in", ids).
		All(&plist, "id", "username", "email", "wechat", "phone", "ding_talk")
	if err != nil {
		logs.Error(err)
	}
	var tplname string
	if event.Status == "0" {
		tplname = "./template/wechat_recovery.tpl"
	} else {
		tplname = "./template/wechat_problem.tpl"
	}
	event.Level = template2.AlertSeverityTo(event.Level)
	event.Status = template2.AlertType(event.Status)
	tmpl, err := template.ParseFiles("./" + tplname)
	if err != nil {
		logs.Error(err)
	}
	var body bytes.Buffer
	err = tmpl.Execute(&body, event) //将str的值合成到tmpl模版的{{.}}中，并将合成得到的文本输入到os.Stdout,返回hello, world
	if err != nil {
		panic(err)
	}
	if err != nil {
		logs.Error(err)
	}
	for _, v := range plist {
		SendWechatAlert(v, event, string(body.Bytes()))
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
		logs.Error(err)
	}
	//update alalrm status
	return nil
}

func PopAllWechat() []*Event {
	ret := []*Event{}
	for {
		var ctx = context.Background()
		reply, err := RDB.RPop(ctx, "wechat").Result()
		if err != nil {
			if err != redis.Nil {
				logs.Error(err)
			}
			break
		}
		if reply == "" || reply == "nil" {
			continue
		}

		var mail Event
		err = json.Unmarshal([]byte(reply), &mail)
		if err != nil {
			logs.Error(err, reply)
			continue
		}
		ret = append(ret, &mail)
	}
	return ret
}
