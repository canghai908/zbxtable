package handler

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

var alarmPromptVarPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

func extractAlarmContext(rawMessage string) (string, string) {
	parts := strings.SplitN(rawMessage, "\n\n", 2)
	if len(parts) == 2 {
		ctx := strings.TrimSpace(parts[0])
		if strings.Contains(ctx, "告警上下文") || strings.Contains(ctx, "Alarm Context") {
			return ctx, strings.TrimSpace(parts[1])
		}
	}
	return "", strings.TrimSpace(rawMessage)
}

type AlarmContextPayload struct {
	Hostname string `json:"hostname"`
	HostIP   string `json:"host_ip"`
	Message  string `json:"message"`
	Detail   string `json:"detail"`
	Level    string `json:"level"`
	Status   string `json:"status"`
}

func parseContextFields(alarmContext string) map[string]string {
	fields := map[string]string{}
	if alarmContext == "" {
		return fields
	}
	for _, line := range strings.Split(alarmContext, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		kv := strings.SplitN(line, ":", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		switch k {
		case "设备名称", "Device Name":
			fields["hostname"] = v
		case "IP":
			fields["host_ip"] = v
		case "告警描述", "Description":
			fields["message"] = v
		case "告警详情", "Detail":
			fields["detail"] = v
		}
	}
	return fields
}

func renderAlarmPromptTemplate(template string, rawMessage string, alarmCtx *AlarmContextPayload) string {
	t := strings.TrimSpace(template)
	if t == "" {
		return rawMessage
	}

	alarmContext, userMessage := extractAlarmContext(rawMessage)
	ctxFields := parseContextFields(alarmContext)
	if alarmCtx != nil {
		if alarmCtx.Hostname != "" {
			ctxFields["hostname"] = alarmCtx.Hostname
		}
		if alarmCtx.HostIP != "" {
			ctxFields["host_ip"] = alarmCtx.HostIP
		}
		if alarmCtx.Message != "" {
			ctxFields["message"] = alarmCtx.Message
		}
		if alarmCtx.Detail != "" {
			ctxFields["detail"] = alarmCtx.Detail
		}
		if alarmCtx.Level != "" {
			ctxFields["level"] = alarmCtx.Level
		}
		if alarmCtx.Status != "" {
			ctxFields["status"] = alarmCtx.Status
		}
	}
	if alarmContext == "" && len(ctxFields) > 0 {
		alarmContext = strings.TrimSpace(
			fmt.Sprintf("告警上下文:\n设备名称: %s\nIP: %s\n告警描述: %s\n告警详情: %s",
				ctxFields["hostname"], ctxFields["host_ip"], ctxFields["message"], ctxFields["detail"]),
		)
	}

	mapping := map[string]string{
		"hostname":      ctxFields["hostname"],
		"host_ip":       ctxFields["host_ip"],
		"message":       ctxFields["message"],
		"detail":        ctxFields["detail"],
		"level":         ctxFields["level"],
		"status":        ctxFields["status"],
		"alarm_context": alarmContext,
	}

	rendered := alarmPromptVarPattern.ReplaceAllStringFunc(t, func(token string) string {
		matched := alarmPromptVarPattern.FindStringSubmatch(token)
		if len(matched) < 2 {
			return token
		}
		key := matched[1]
		if val, ok := mapping[key]; ok {
			return val
		}
		return token
	})

	if userMessage != "" {
		rendered = strings.TrimSpace(rendered + "\n\n" + userMessage)
	}
	return rendered
}

// AIChat 流式聊天接口
func AIChat(c *gin.Context) {
	// 设置响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 解析请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, "请求体读取失败")
		return
	}

	var userReq struct {
		Message      string               `json:"message"`
		AlarmContext *AlarmContextPayload `json:"alarm_context"`
	}
	if err := jsoniter.Unmarshal(body, &userReq); err != nil {
		var AIRes struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		}
		AIRes.Code = 500
		AIRes.Message = err.Error()
		c.JSON(http.StatusOK, AIRes)
		return
	}

	// 告警分析提示词（后端兜底渲染，避免仅依赖前端）
	alarmPromptTpl := model.GetConfigValueByKey("alarm_analysis_prompt", "")
	finalMessage := renderAlarmPromptTemplate(alarmPromptTpl, userReq.Message, userReq.AlarmContext)

	// 获取 AI 类型配置
	aiType := model.GetConfigValueByKey("ai_type", "ollama")
	if aiType == "" {
		aiType = "ollama"
	}

	// 根据 AI 类型调用不同的服务
	switch aiType {
	case "deepseek":
		handleDeepseekChat(c, finalMessage)
	case "ollama":
		handleOllamaChat(c, finalMessage)
	default:
		// 默认使用 Ollama
		handleOllamaChat(c, finalMessage)
	}
}

// handleOllamaChat 处理 Ollama AI 请求
func handleOllamaChat(c *gin.Context, message string) {
	// 构造系统提示词
	systemPrompt := struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{
		Role:    "system",
		Content: "你是一个专业的运维分析师，请分析用户提供的问题并给出专业的建议。",
	}

	// 从配置获取 Ollama 地址
	ollamaHost := model.GetConfigValueByKey("ollama_host", model.GetConfKey("ollama_host"))
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}
	ollamaModel := model.GetConfigValueByKey("ollama_model", model.GetConfKey("ollama_model"))
	if ollamaModel == "" {
		ollamaModel = "deepseek-r1:32b"
	}

	// 构造 Ollama 请求
	ollamaReq := struct {
		Model    string `json:"model"`
		Messages []any  `json:"messages"`
		Stream   bool   `json:"stream"`
	}{
		Model:  ollamaModel,
		Stream: true,
		Messages: []any{
			systemPrompt,
			struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			}{
				Role:    "user",
				Content: message,
			},
		},
	}

	jsonData, err := jsoniter.Marshal(ollamaReq)
	if err != nil {
		var AIRes struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		}
		AIRes.Code = 500
		AIRes.Message = err.Error()
		c.JSON(http.StatusOK, AIRes)
		return
	}

	ollamaURL := fmt.Sprintf("%s/api/chat", ollamaHost)

	// 发送请求到 Ollama
	resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		var AIRes struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		}
		AIRes.Code = 500
		AIRes.Message = fmt.Sprintf("Ollama 服务连接失败: %v", err)
		c.JSON(http.StatusOK, AIRes)
		return
	}
	defer resp.Body.Close()

	// 创建 reader
	reader := bufio.NewReader(resp.Body)

	// 流式读取响应
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}

		// 解析响应
		var response struct {
			Model     string `json:"model"`
			CreatedAt string `json:"created_at"`
			Message   struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			Done bool `json:"done"`
		}
		if err := jsoniter.Unmarshal(line, &response); err != nil {
			continue
		}

		// 发送数据
		if response.Message.Content != "" {
			c.Writer.WriteString(response.Message.Content)
			c.Writer.Flush()
		}

		// 如果响应完成，退出循环
		if response.Done {
			break
		}
	}
}

// handleDeepseekChat 处理 Deepseek AI 请求
func handleDeepseekChat(c *gin.Context, message string) {
	// 从配置获取 Deepseek 参数（会自动解密）
	apiKey := model.GetConfigValueByKey("deepseek_api_key", "")
	if apiKey == "" {
		var AIRes struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		}
		AIRes.Code = 500
		AIRes.Message = "Deepseek API Key 未配置"
		c.JSON(http.StatusOK, AIRes)
		return
	}

	deepseekModel := model.GetConfigValueByKey("deepseek_model", "deepseek-chat")
	if deepseekModel == "" {
		deepseekModel = "deepseek-chat"
	}

	baseURL := model.GetConfigValueByKey("deepseek_base_url", "https://api.deepseek.com")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	// 构造 Deepseek 请求
	deepseekReq := struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Stream bool `json:"stream"`
	}{
		Model:  deepseekModel,
		Stream: true,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "system",
				Content: "你是一个专业的运维分析师，请分析用户提供的问题并给出专业的建议。",
			},
			{
				Role:    "user",
				Content: message,
			},
		},
	}

	jsonData, err := jsoniter.Marshal(deepseekReq)
	if err != nil {
		var AIRes struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		}
		AIRes.Code = 500
		AIRes.Message = err.Error()
		c.JSON(http.StatusOK, AIRes)
		return
	}

	// 构造 Deepseek API URL
	deepseekURL := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(baseURL, "/"))

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", deepseekURL, bytes.NewBuffer(jsonData))
	if err != nil {
		var AIRes struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		}
		AIRes.Code = 500
		AIRes.Message = err.Error()
		c.JSON(http.StatusOK, AIRes)
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// 发送请求到 Deepseek
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		var AIRes struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		}
		AIRes.Code = 500
		AIRes.Message = fmt.Sprintf("Deepseek 服务连接失败: %v", err)
		c.JSON(http.StatusOK, AIRes)
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		var AIRes struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		}
		AIRes.Code = resp.StatusCode
		AIRes.Message = fmt.Sprintf("Deepseek API 错误: %s", string(bodyBytes))
		c.JSON(http.StatusOK, AIRes)
		return
	}

	// 创建 reader
	reader := bufio.NewReader(resp.Body)

	// 流式读取响应
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}

		// 跳过空行
		lineStr := strings.TrimSpace(string(line))
		if lineStr == "" {
			continue
		}

		// Deepseek 使用 SSE 格式，需要去掉 "data: " 前缀
		lineStr = strings.TrimPrefix(lineStr, "data: ")

		// 检查是否是结束标记
		if lineStr == "[DONE]" {
			break
		}

		// 解析响应
		var response struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			Model   string `json:"model"`
			Choices []struct {
				Index int `json:"index"`
				Delta struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := jsoniter.Unmarshal([]byte(lineStr), &response); err != nil {
			continue
		}

		// 发送数据
		if len(response.Choices) > 0 && response.Choices[0].Delta.Content != "" {
			c.Writer.WriteString(response.Choices[0].Delta.Content)
			c.Writer.Flush()
		}

		// 检查是否完成
		if len(response.Choices) > 0 && response.Choices[0].FinishReason != "" {
			break
		}
	}
}
