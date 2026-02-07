package model

import (
	"bytes"
	"fmt"
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
	var plist []User
	err := DB.Where("id IN ?", ids).
		Select("id", "username", "email", "wechat", "phone", "ding_talk").
		Find(&plist).Error
	if err != nil {
		logger.Log.Error(err)
	}
	var tplname string
	if event.Status == "0" {
		tplname = TplPath + "wechat_recovery.tpl"
	} else {
		tplname = TplPath + "/wechat_problem.tpl"
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
func SendWechatAlert(user User, event *Event, content string) error {
	// 检查企业微信是否已初始化
	if WeApp == nil {
		err := fmt.Errorf("企业微信未初始化，请检查配置")
		logger.Log.Error(err)
		// 记录失败日志
		elog := EventLog{
			AlarmID:       int64(event.ID),
			EventID:       event.EventID,
			Rule:          event.Rule,
			Channel:       "wechat",
			User:          user.Username,
			Account:       user.Wechat,
			NotifyTime:    time.Now(),
			NotifyContent: content,
			Status:        strconv.Itoa(EventFailed),
			NotifyError:   err.Error(),
		}
		AddEventLog(&elog)
		return err
	}

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

// SendTestWechat 发送测试企业微信消息
func SendTestWechat(userID string) error {
	// 获取企业微信配置（注意：键名要与数据库中的一致）
	corpID := GetConfigValueByKey("wechat_corpid", "")
	agentID := GetConfigValueByKey("wechat_agentid", "")
	secret := GetConfigValueByKey("wechat_secret", "")

	// 验证必填配置
	if corpID == "" || agentID == "" || secret == "" {
		return fmt.Errorf("企业微信配置不完整，请检查企业ID、应用ID和Secret配置")
	}

	// 检查 WeApp 是否已初始化
	if WeApp == nil {
		return fmt.Errorf("企业微信应用未初始化，请检查配置并重启服务")
	}

	// 构建测试消息
	testContent := fmt.Sprintf(`【ZbxTable 企业微信配置测试】

✓ 配置测试成功！

您的企业微信配置已经正常工作，ZbxTable 可以正常发送企业微信通知了。

📋 配置信息：
企业ID：%s
应用ID：%s
测试用户：%s

⏰ 测试时间：%s

此消息由 ZbxTable 监控系统自动发送。`,
		corpID,
		agentID,
		userID,
		time.Now().Format("2006-01-02 15:04:05"))

	// 发送测试消息
	tos := workwx.Recipient{
		UserIDs: []string{userID},
	}

	err := WeApp.SendTextMessage(&tos, testContent, false)
	if err != nil {
		return fmt.Errorf("发送失败: %v", err)
	}

	return nil
}
