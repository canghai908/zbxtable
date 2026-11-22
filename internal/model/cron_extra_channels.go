package model

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"
)

const (
	defaultProblemTextTpl = `[{{.Status}}][{{.OccurTime.Format "15:04:05"}}]设备:{{.Hostname}}发生:{{.Message}}故障！
=========================================
Zabbix实例: {{.InstanceName}}
告警主机: {{.Hostname}}
主机IP：{{.HostsIP}}
主机分组: {{.Hgroup}}
告警时间: {{.OccurTime.Format "2006-01-02 15:04:05"}}
告警等级: {{.Level}}
告警信息: {{.ItemName}}
告警项目: {{.Hkey}}
问题详情: {{.Detail}}
当前状态: {{.Status}}
事件ID: {{.EventID}}`

	defaultRecoveryTextTpl = `[{{.Status}}][{{.OccurTime.Format "15:04:05"}}]设备:{{.Hostname}}发生:{{.Message}}已恢复！
=========================================
Zabbix实例: {{.InstanceName}}
告警主机: {{.Hostname}}
主机IP：{{.HostsIP}}
主机分组: {{.Hgroup}}
告警时间: {{.OccurTime.Format "2006-01-02 15:04:05"}}
告警等级: {{.Level}}
告警信息: {{.ItemName}}
告警项目: {{.Hkey}}
问题详情: {{.Detail}}
持续时长：{{.EventDuration}}
当前状态: {{.Status}}
事件ID: {{.EventID}}`
)

func decryptUserCredential(raw, encryptionKey string) string {
	if raw == "" {
		return ""
	}
	decrypted, err := utils.DecryptString(raw, encryptionKey)
	if err != nil {
		return raw
	}
	return decrypted
}

func renderAlertTextContent(event *Event) (string, error) {
	var tplname string
	fallback := defaultProblemTextTpl
	if event.Status == "0" {
		tplname = TplPath + "wechat_recovery.tpl"
		fallback = defaultRecoveryTextTpl
	} else {
		tplname = TplPath + "wechat_problem.tpl"
	}

	display := *event
	display.Level = utils.AlertSeverityTo(event.Level)
	display.Status = utils.AlertType(event.Status)

	tmpl, err := template.ParseFiles("./" + tplname)
	if err != nil {
		tmpl, err = template.New("alert_text").Parse(fallback)
		if err != nil {
			return "", err
		}
	}
	var body bytes.Buffer
	if err := tmpl.Execute(&body, &display); err != nil {
		return "", err
	}
	return body.String(), nil
}

func popAllFromChannelQueue(channel string) []*Event {
	ret := []*Event{}
	for {
		reply, err := CacheRPop(channel)
		if err != nil || reply == "" || reply == "nil" {
			break
		}
		var event Event
		if err := json.Unmarshal([]byte(reply), &event); err != nil {
			logger.Log.Error(err, reply)
			continue
		}
		ret = append(ret, &event)
	}
	return ret
}

func loadEventRecipientUsers(toUsers string) ([]User, error) {
	ids := strings.Split(toUsers, ",")
	var plist []User
	err := DB.Where("id IN ?", ids).
		Select("id", "username", "email", "wechat", "wechat_robot_key", "phone", "ding_talk", "ding_talk_secret", "feishu_robot_key", "feishu_secret", "webhook_url", "webhook_token").
		Find(&plist).Error
	return plist, err
}

func postJSONWebhook(targetURL string, headers map[string]string, payload any, errLabel string) (int, []byte, error) {
	messageBytes, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, err
	}
	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(messageBytes))
	if err != nil {
		return 0, nil, fmt.Errorf("%s请求创建失败", errLabel)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// 避免 http 错误中暴露含 token/key 的 URL
		return 0, nil, fmt.Errorf("%s请求失败", errLabel)
	}
	defer resp.Body.Close()
	const responseLimit = 64 * 1024
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, responseLimit+1))
	if readErr != nil || len(body) > responseLimit {
		return resp.StatusCode, nil, fmt.Errorf("%s响应读取失败或超出大小限制", errLabel)
	}
	return resp.StatusCode, body, nil
}

func recordChannelDeliveryLog(event *Event, channel, username, account, content string, sendErr error) error {
	elog := EventLog{
		AlarmID:       int64(event.ID),
		EventID:       event.EventID,
		Rule:          event.Rule,
		Channel:       channel,
		User:          username,
		Account:       account,
		NotifyTime:    time.Now(),
		NotifyContent: content,
		Status:        strconv.Itoa(EventSuccess),
	}
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

// ========== 钉钉群机器人 ==========

// ConsumeDingTalkRobot 消费钉钉群机器人消息队列
func ConsumeDingTalkRobot() {
	for {
		L := PopAllDingTalkRobot()
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendDingTalkRobotList(L)
	}
}

// PopAllDingTalkRobot 从队列中弹出所有钉钉机器人消息
func PopAllDingTalkRobot() []*Event {
	return popAllFromChannelQueue("dingtalk_robot")
}

// SendDingTalkRobotList 批量发送钉钉机器人消息
func SendDingTalkRobotList(L []*Event) {
	for _, v := range L {
		DingTalkRobotWorkerChan <- 1
		go SendDingTalkRobot(v)
	}
}

// SendDingTalkRobot 发送单条钉钉机器人消息
func SendDingTalkRobot(event *Event) {
	defer func() {
		<-DingTalkRobotWorkerChan
	}()
	plist, err := loadEventRecipientUsers(event.ToUsers)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	content, err := renderAlertTextContent(event)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	encryptionKey := GetEncryptionKey()
	for _, v := range plist {
		if strings.TrimSpace(v.DingTalk) == "" {
			logger.Log.Warningf("User %s does not have ding_talk configured", v.Username)
			continue
		}
		key := decryptUserCredential(v.DingTalk, encryptionKey)
		secret := decryptUserCredential(v.DingTalkSecret, encryptionKey)
		_ = SendDingTalkRobotAlert(v, event, content, key, secret)
	}
}

// BuildDingTalkWebhookURL 根据 token 或完整 URL 与可选加签密钥构造请求地址
func BuildDingTalkWebhookURL(keyOrURL, secret string, now time.Time) string {
	base := strings.TrimSpace(keyOrURL)
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = fmt.Sprintf("https://oapi.dingtalk.com/robot/send?access_token=%s", base)
	}
	if strings.TrimSpace(secret) == "" {
		return base
	}
	ts := now.UnixMilli()
	stringToSign := fmt.Sprintf("%d\n%s", ts, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(stringToSign))
	sign := url.QueryEscape(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	sep := "&"
	if !strings.Contains(base, "?") {
		sep = "?"
	}
	return fmt.Sprintf("%s%stimestamp=%d&sign=%s", base, sep, ts, sign)
}

func sendDingTalkRobotHTTP(content, decryptedKey, decryptedSecret string, now time.Time) error {
	webhookURL := BuildDingTalkWebhookURL(decryptedKey, decryptedSecret, now)
	payload := map[string]any{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}
	statusCode, body, sendErr := postJSONWebhook(webhookURL, nil, payload, "钉钉群机器人")
	if sendErr != nil {
		return sendErr
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("钉钉群机器人 HTTP status code: %d", statusCode)
	}
	var result struct {
		ErrCode *int `json:"errcode"`
	}
	if json.Unmarshal(body, &result) != nil || result.ErrCode == nil {
		return errors.New("钉钉群机器人返回无效响应")
	}
	if *result.ErrCode != 0 {
		return fmt.Errorf("钉钉群机器人拒绝消息，errcode: %d", *result.ErrCode)
	}
	return nil
}

// SendDingTalkRobotAlert 发送钉钉群机器人告警消息并记录事件日志
func SendDingTalkRobotAlert(user User, event *Event, content, decryptedKey, decryptedSecret string) error {
	sendErr := sendDingTalkRobotHTTP(content, decryptedKey, decryptedSecret, time.Now())
	return recordChannelDeliveryLog(event, "dingtalk_robot", user.Username, user.DingTalk, content, sendErr)
}

// ========== 飞书群机器人 ==========

// ConsumeFeishuRobot 消费飞书群机器人消息队列
func ConsumeFeishuRobot() {
	for {
		L := PopAllFeishuRobot()
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendFeishuRobotList(L)
	}
}

// PopAllFeishuRobot 从队列中弹出所有飞书机器人消息
func PopAllFeishuRobot() []*Event {
	return popAllFromChannelQueue("feishu_robot")
}

// SendFeishuRobotList 批量发送飞书机器人消息
func SendFeishuRobotList(L []*Event) {
	for _, v := range L {
		FeishuRobotWorkerChan <- 1
		go SendFeishuRobot(v)
	}
}

// SendFeishuRobot 发送单条飞书机器人消息
func SendFeishuRobot(event *Event) {
	defer func() {
		<-FeishuRobotWorkerChan
	}()
	plist, err := loadEventRecipientUsers(event.ToUsers)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	content, err := renderAlertTextContent(event)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	encryptionKey := GetEncryptionKey()
	for _, v := range plist {
		if strings.TrimSpace(v.FeishuRobotKey) == "" {
			logger.Log.Warningf("User %s does not have feishu_robot_key configured", v.Username)
			continue
		}
		key := decryptUserCredential(v.FeishuRobotKey, encryptionKey)
		secret := decryptUserCredential(v.FeishuSecret, encryptionKey)
		_ = SendFeishuRobotAlert(v, event, content, key, secret)
	}
}

// BuildFeishuSign 计算飞书自定义机器人签名
func BuildFeishuSign(secret string, timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	mac := hmac.New(sha256.New, []byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func sendFeishuRobotHTTP(content, decryptedKey, decryptedSecret string, now time.Time) error {
	webhookURL := strings.TrimSpace(decryptedKey)
	if !strings.HasPrefix(webhookURL, "http://") && !strings.HasPrefix(webhookURL, "https://") {
		webhookURL = fmt.Sprintf("https://open.feishu.cn/open-apis/bot/v2/hook/%s", webhookURL)
	}
	payload := map[string]any{
		"msg_type": "text",
		"content": map[string]string{
			"text": content,
		},
	}
	if strings.TrimSpace(decryptedSecret) != "" {
		ts := now.Unix()
		payload["timestamp"] = strconv.FormatInt(ts, 10)
		payload["sign"] = BuildFeishuSign(decryptedSecret, ts)
	}
	statusCode, body, sendErr := postJSONWebhook(webhookURL, nil, payload, "飞书群机器人")
	if sendErr != nil {
		return sendErr
	}
	if statusCode != http.StatusOK {
		return fmt.Errorf("飞书群机器人 HTTP status code: %d", statusCode)
	}
	var result struct {
		Code       *int `json:"code"`
		StatusCode *int `json:"StatusCode"`
	}
	if json.Unmarshal(body, &result) != nil || (result.Code == nil && result.StatusCode == nil) {
		return errors.New("飞书群机器人返回无效响应")
	}
	if result.Code != nil && *result.Code != 0 {
		return fmt.Errorf("飞书群机器人拒绝消息，code: %d", *result.Code)
	}
	if result.Code == nil && result.StatusCode != nil && *result.StatusCode != 0 {
		return fmt.Errorf("飞书群机器人拒绝消息，StatusCode: %d", *result.StatusCode)
	}
	return nil
}

// SendFeishuRobotAlert 发送飞书群机器人告警消息并记录事件日志
func SendFeishuRobotAlert(user User, event *Event, content, decryptedKey, decryptedSecret string) error {
	sendErr := sendFeishuRobotHTTP(content, decryptedKey, decryptedSecret, time.Now())
	return recordChannelDeliveryLog(event, "feishu_robot", user.Username, user.FeishuRobotKey, content, sendErr)
}

// ========== 通用自定义 Webhook ==========

// ConsumeWebhookChannel 消费通用 Webhook 消息队列
func ConsumeWebhookChannel() {
	for {
		L := PopAllWebhookChannel()
		if len(L) == 0 {
			time.Sleep(time.Millisecond * 200)
			continue
		}
		SendWebhookChannelList(L)
	}
}

// PopAllWebhookChannel 从队列中弹出所有通用 Webhook 消息
func PopAllWebhookChannel() []*Event {
	return popAllFromChannelQueue("webhook")
}

// SendWebhookChannelList 批量发送通用 Webhook 消息
func SendWebhookChannelList(L []*Event) {
	for _, v := range L {
		WebhookWorkerChan <- 1
		go SendWebhookChannel(v)
	}
}

// SendWebhookChannel 发送单条通用 Webhook 消息
func SendWebhookChannel(event *Event) {
	defer func() {
		<-WebhookWorkerChan
	}()
	plist, err := loadEventRecipientUsers(event.ToUsers)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	content, err := renderAlertTextContent(event)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	encryptionKey := GetEncryptionKey()
	for _, v := range plist {
		if strings.TrimSpace(v.WebhookURL) == "" {
			logger.Log.Warningf("User %s does not have webhook_url configured", v.Username)
			continue
		}
		token := decryptUserCredential(v.WebhookToken, encryptionKey)
		_ = SendWebhookChannelAlert(v, event, content, v.WebhookURL, token)
	}
}

func sendWebhookChannelHTTP(event *Event, content, webhookURL, decryptedToken string) error {
	targetURL := strings.TrimSpace(webhookURL)
	payload := map[string]any{
		"alarm_id":       event.ID,
		"event_id":       event.EventID,
		"zid":            event.ZID,
		"instance":       event.Instance,
		"instance_name":  event.InstanceName,
		"host_id":        event.HostID,
		"host":           event.Host,
		"hostname":       event.Hostname,
		"host_ip":        event.HostsIP,
		"host_group":     event.Hgroup,
		"trigger_id":     event.TriggerID,
		"item_id":        event.ItemID,
		"item_name":      event.ItemName,
		"item_value":     event.ItemValue,
		"hkey":           event.Hkey,
		"level":          utils.AlertSeverityTo(event.Level),
		"status":         utils.AlertType(event.Status),
		"message":        event.Message,
		"detail":         event.Detail,
		"occur_time":     event.OccurTime.Format("2006-01-02 15:04:05"),
		"event_duration": event.EventDuration,
		"rule":           event.Rule,
		"content":        content,
	}
	headers := map[string]string{}
	if strings.TrimSpace(decryptedToken) != "" {
		headers["Authorization"] = "Bearer " + decryptedToken
		headers["X-Webhook-Token"] = decryptedToken
	}
	statusCode, _, sendErr := postJSONWebhook(targetURL, headers, payload, "自定义Webhook")
	if sendErr != nil {
		return sendErr
	}
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("自定义Webhook HTTP status code: %d", statusCode)
	}
	return nil
}

// SendWebhookChannelAlert 发送通用自定义 Webhook 告警通知并记录事件日志
func SendWebhookChannelAlert(user User, event *Event, content, webhookURL, decryptedToken string) error {
	targetURL := strings.TrimSpace(webhookURL)
	sendErr := sendWebhookChannelHTTP(event, content, targetURL, decryptedToken)
	return recordChannelDeliveryLog(event, "webhook", user.Username, targetURL, content, sendErr)
}
