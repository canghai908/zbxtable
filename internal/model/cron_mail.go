package model

import (
	"bytes"
	"crypto/tls"

	"fmt"
	"html/template"
	"net/smtp"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"

	"github.com/jordan-wright/email"
)

// copy from open-facon-plus
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

func SendMail(mail *Event) {
	defer func() {
		<-MailWorkerChan
	}()
	ids := strings.Split(mail.ToUsers, ",")
	var plist []User
	err := DB.Where("id IN ?", ids).
		Select("id", "username", "email", "wechat", "phone", "ding_talk").
		Find(&plist).Error
	if err != nil {
		logger.Log.Error(err)
	}
	mail.Level = utils.AlertSeverityTo(mail.Level)
	mail.Status = utils.AlertType(mail.Status)
	for _, v := range plist {
		SendEmailAlert(mail, v)
	}
}

func PopAllMail() []*Event {
	ret := []*Event{}
	for {
		reply, err := CacheRPop("mail")
		if err != nil {
			logger.Log.Error(err)
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

func SendEmailAlert(event *Event, user User) error {
	//发邮件
	var err error
	var tplname string
	if event.Status == "0" {
		tplname = TplPath + "mail_recovery.tpl"
	} else {
		tplname = TplPath + "mail_problem.tpl"
	}
	tmpl, err := template.ParseFiles("./" + tplname)
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	var body bytes.Buffer
	err = tmpl.Execute(&body, event) //将str的值合成到tmpl模版的{{.}}中，并将合成得到的文本输入到os.Stdout,返回hello, world
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	// 邮件配置优先从系统配置表读取，其次回退到 app.conf
	from := GetConfigValueByKey("email_from", GetConfKey("email_from"))
	nickname := GetConfigValueByKey("email_nickname", GetConfKey("email_nickname"))
	secret := GetConfigValueByKey("email_secret", GetConfKey("email_secret"))
	host := GetConfigValueByKey("email_host", GetConfKey("email_host"))
	portStr := GetConfigValueByKey("email_port", GetConfKey("email_port"))
	if portStr == "" {
		portStr = "465"
	}
	port, _ := strconv.Atoi(portStr)
	isSSlStr := GetConfigValueByKey("email_isSSl", GetConfKey("email_isSSl"))
	isSSL := strings.ToLower(isSSlStr) == "true"
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
			NotifyTime: time.Now(), NotifyContent: body.String(),
			Status: strconv.Itoa(EventFailed), NotifyError: err.Error(),
		}
	} else {
		elog = EventLog{AlarmID: int64(event.ID), EventID: event.EventID,
			Rule: event.Rule, Channel: "mail", User: user.Username, Account: user.Email,
			NotifyTime: time.Now(), NotifyContent: body.String(),
			Status: strconv.Itoa(EventSuccess), NotifyError: "",
		}
	}
	//add event log
	_, err = AddEventLog(&elog)
	if err != nil {
		logger.Log.Error(err)
	}
	return nil
}
