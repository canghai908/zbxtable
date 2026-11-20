package model

import (
	"bytes"
	"errors"
	"io"

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
	var plist []User
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

	// 获取加密密钥
	encryptionKey := GetEncryptionKey()

	for _, v := range plist {
		if v.WechatRobotKey != "" {
			// 解密企业微信群机器人Key
			decryptedKey := v.WechatRobotKey
			if v.WechatRobotKey != "" {
				key, err := utils.DecryptString(v.WechatRobotKey, encryptionKey)
				if err != nil {
					// 解密失败，可能是旧数据未加密，使用原值
					logger.Log.Warnf("解密用户 %s 的企业微信群机器人Key失败，使用原值: %v", v.Username, err)
				} else {
					decryptedKey = key
				}
			}

			// 使用解密后的Key发送消息
			SendWechatRobotAlert(v, event, body.String(), decryptedKey)
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
func SendWechatRobotAlert(user User, event *Event, content string, decryptedKey string) error {
	// 构建webhook URL
	webhookURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", decryptedKey)

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

	// A successful HTTP response can still be a rejected WeCom request.
	resp, sendErr := http.Post(webhookURL, "application/json", bytes.NewBuffer(messageBytes))
	if sendErr != nil {
		// http.Post errors contain the request URL, including the private webhook key.
		sendErr = errors.New("企业微信群机器人请求失败")
	} else {
		defer resp.Body.Close()
		switch {
		case resp.StatusCode != http.StatusOK:
			sendErr = fmt.Errorf("企业微信群机器人 HTTP status code: %d", resp.StatusCode)
		default:
			const responseLimit = 64 * 1024
			body, err := io.ReadAll(io.LimitReader(resp.Body, responseLimit+1))
			var result struct {
				ErrCode *int `json:"errcode"`
			}
			if err != nil || len(body) > responseLimit || json.Unmarshal(body, &result) != nil || result.ErrCode == nil {
				sendErr = errors.New("企业微信群机器人返回无效响应")
			} else if *result.ErrCode != 0 {
				sendErr = fmt.Errorf("企业微信群机器人拒绝消息，errcode: %d", *result.ErrCode)
			}
		}
	}
	elog := EventLog{AlarmID: int64(event.ID), EventID: event.EventID, Rule: event.Rule,
		Channel: "wechat_robot", User: user.Username, Account: user.WechatRobotKey,
		NotifyTime: time.Now(), NotifyContent: content, Status: strconv.Itoa(EventSuccess)}
	if sendErr != nil {
		elog.Status = strconv.Itoa(EventFailed)
		elog.NotifyError = sendErr.Error()
		logger.Log.Error(sendErr)
	}
	_, logErr := AddEventLog(&elog)
	if logErr != nil {
		logger.Log.Error(logErr)
	}
	return errors.Join(sendErr, logErr)
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
