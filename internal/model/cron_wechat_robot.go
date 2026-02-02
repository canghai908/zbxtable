package model

import (
	"bytes"

	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"
)

// ConsumeWechatRobot 消费企业微信群机器人消息队列
func ConsumeWechatRobot() {
	for {
		L := PopAllWechatRobot()
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendWechatRobotList(L)
	}
}

// SendWechatRobotList 批量发送企业微信群机器人消息
func SendWechatRobotList(L []*Event) {
	for _, v := range L {
		WechatRobotWorkerChan <- 1
		go SendWechatRobot(v)
	}
}

// SendWechatRobot 发送企业微信群机器人消息
func SendWechatRobot(event *Event) {
	defer func() {
		<-WechatRobotWorkerChan
	}()
	ids := strings.Split(event.ToUsers, ",")
	var plist []Manager
	err := DB.Where("id IN ?", ids).
		Select("id", "username", "email", "wechat", "wechat_robot_key", "phone", "ding_talk").
		Find(&plist).Error
	if err != nil {
		logger.Log.Error(err)
		return
	}
	var tplname string
	if event.Status == "0" {
		tplname = TplPath + "wechat_recovery.tpl"
	} else {
		tplname = TplPath + "wechat_problem.tpl"
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
		if v.WechatRobotKey != "" {
			SendWechatRobotAlert(v, event, body.String())
		} else {
			logger.Log.Warningf("User %s does not have wechat_robot_key configured", v.Username)
		}
	}
}

// WechatRobotMessage 企业微信群机器人消息结构
type WechatRobotMessage struct {
	MsgType  string   `json:"msgtype"`
	Text     Text     `json:"text,omitempty"`
	Markdown Markdown `json:"markdown,omitempty"`
}

type Text struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

type Markdown struct {
	Content string `json:"content"`
}

// SendWechatRobotAlert 发送企业微信群机器人告警消息
func SendWechatRobotAlert(user Manager, event *Event, content string) error {
	// 构建webhook URL
	webhookURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", user.WechatRobotKey)

	// 构建消息体
	message := WechatRobotMessage{
		MsgType: "text",
		Text: Text{
			Content: content,
		},
	}

	// 序列化消息
	messageBytes, err := json.Marshal(message)
	if err != nil {
		logger.Log.Error("Failed to marshal wechat robot message:", err)
		return err
	}

	// 发送HTTP POST请求
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(messageBytes))
	if err != nil {
		logger.Log.Error("Failed to send wechat robot message:", err)
		var elog EventLog
		elog = EventLog{AlarmID: int64(event.ID), EventID: event.EventID,
			Rule: event.Rule, Channel: "wechat_robot", User: user.Username, Account: user.WechatRobotKey,
			NotifyTime: time.Now(), NotifyContent: content,
			Status: strconv.Itoa(EventFailed), NotifyError: err.Error(),
		}
		_, err = AddEventLog(&elog)
		if err != nil {
			logger.Log.Error(err)
		}
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	var elog EventLog
	if resp.StatusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("HTTP status code: %d", resp.StatusCode)
		logger.Log.Error("Wechat robot API returned error:", errorMsg)
		elog = EventLog{AlarmID: int64(event.ID), EventID: event.EventID,
			Rule: event.Rule, Channel: "wechat_robot", User: user.Username, Account: user.WechatRobotKey,
			NotifyTime: time.Now(), NotifyContent: content,
			Status: strconv.Itoa(EventFailed), NotifyError: errorMsg,
		}
	} else {
		elog = EventLog{AlarmID: int64(event.ID), EventID: event.EventID,
			Rule: event.Rule, Channel: "wechat_robot", User: user.Username, Account: user.WechatRobotKey,
			NotifyTime: time.Now(), NotifyContent: content,
			Status: strconv.Itoa(EventSuccess), NotifyError: "",
		}
	}

	// 添加事件日志
	_, err = AddEventLog(&elog)
	if err != nil {
		logger.Log.Error(err)
	}

	return nil
}

// PopAllWechatRobot 从内存队列中弹出所有企业微信群机器人消息
func PopAllWechatRobot() []*Event {
	ret := []*Event{}
	for {
		reply, err := CacheRPop("wechat_robot")
		if err != nil {
			break
		}
		if reply == "" || reply == "nil" {
			break
		}

		var event Event
		err = json.Unmarshal([]byte(reply), &event)
		if err != nil {
			logger.Log.Error(err, reply)
			continue
		}
		ret = append(ret, &event)
	}
	return ret
}
