package models

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	redis "github.com/go-redis/redis/v8"
	"github.com/jordan-wright/email"
	"html/template"
	"net/smtp"
	"strconv"
	"strings"
	"time"
	template2 "zbxtable/utils"
)

//copy from open-facon-plus
func ConsumeMail() {
	for {
		L := PopAllMail()
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendMailList(L)
	}
}
func SendMailList(L []*Event) {
	for _, mail := range L {
		MailWorkerChan <- 1
		go SendMail(mail)
	}
}

//
func SendMail(mail *Event) {
	defer func() {
		<-MailWorkerChan
	}()
	o := orm.NewOrm()
	ids := strings.Split(mail.ToUsers, ",")
	var user Manager
	var plist []Manager
	_, err := o.QueryTable(user).Filter("id__in", ids).
		All(&plist, "id", "username", "email", "wechat", "phone", "ding_talk")
	if err != nil {
		logs.Error(err)
	}
	mail.Level = template2.AlertSeverityTo(mail.Level)
	mail.Status = template2.AlertType(mail.Status)
	for _, v := range plist {
		SendEmailAlert(mail, v)
	}
}

func PopAllMail() []*Event {
	ret := []*Event{}
	for {
		var ctx = context.Background()
		reply, err := RDB.RPop(ctx, "mail").Result()
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

func SendEmailAlert(event *Event, user Manager) error {
	//发邮件
	var err error
	var tplname string
	if event.Status == "0" {
		tplname = "./template/mail_recovery.tpl"
	} else {
		tplname = "./template/mail_problem.tpl"
	}
	tmpl, err := template.ParseFiles("./" + tplname)
	if err != nil {
		logs.Error(err)
		return err
	}
	var body bytes.Buffer
	err = tmpl.Execute(&body, event) //将str的值合成到tmpl模版的{{.}}中，并将合成得到的文本输入到os.Stdout,返回hello, world
	if err != nil {
		logs.Error(err)
		return err
	}
	//tos := strings.Split(m.Emails, ",")
	from := beego.AppConfig.String("email_from")
	nickname := beego.AppConfig.String("email_nickname")
	secret := beego.AppConfig.String("email_secret")
	host := beego.AppConfig.String("email_host")
	port, _ := beego.AppConfig.Int("email_port")
	isSSL, _ := beego.AppConfig.Bool("email_isSSl")
	auth := smtp.PlainAuth("", from, secret, host)
	at := smtp.CRAMMD5Auth(from, secret)
	e := email.NewEmail()
	if nickname != "" {
		e.From = fmt.Sprintf("%s <%s>", nickname, from)
	} else {
		e.From = from
	}
	e.To = []string{user.Email}
	e.Subject = "[" + event.Status + "]" + "[" + event.Message + "]" + event.Hostname + "(" + event.HostsIP + ")"
	e.HTML = body.Bytes()
	hostAddr := fmt.Sprintf("%s:%d", host, port)
	if isSSL {
		err = e.SendWithTLS(hostAddr, auth, &tls.Config{ServerName: host})
	} else {
		err = e.Send(hostAddr, at)
	}
	var elog EventLog
	if err != nil {
		elog = EventLog{AlarmID: int64(event.ID), EventID: event.EventID,
			Rule: event.Rule, Channel: "mail", User: user.Username, Account: user.Email,
			NotifyTime: time.Now(), NotifyContent: string(body.Bytes()),
			Status: strconv.Itoa(EventFailed), NotifyError: err.Error(),
		}
	} else {
		elog = EventLog{AlarmID: int64(event.ID), EventID: event.EventID,
			Rule: event.Rule, Channel: "mail", User: user.Username, Account: user.Email,
			NotifyTime: time.Now(), NotifyContent: string(body.Bytes()),
			Status: strconv.Itoa(EventSuccess), NotifyError: "",
		}
	}
	//add event log
	_, err = AddEventLog(&elog)
	if err != nil {
		logs.Error(err)
	}
	return nil
}
