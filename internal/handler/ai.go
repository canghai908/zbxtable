package handler

import (
	"zbxtable/pkg/response"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

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
		Message string `json:"message"`
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

	// 获取 AI 类型配置
	aiType := model.GetConfigValueByKey("ai_type", "ollama")
	if aiType == "" {
		aiType = "ollama"
	}

	// 根据 AI 类型调用不同的服务
	switch aiType {
	case "deepseek":
		handleDeepseekChat(c, userReq.Message)
	case "ollama":
		handleOllamaChat(c, userReq.Message)
	default:
		// 默认使用 Ollama
		handleOllamaChat(c, userReq.Message)
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
		Model    string        `json:"model"`
		Messages []interface{} `json:"messages"`
		Stream   bool          `json:"stream"`
	}{
		Model:  ollamaModel,
		Stream: true,
		Messages: []interface{}{
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
	// 从配置获取 Deepseek 参数
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
		if strings.HasPrefix(lineStr, "data: ") {
			lineStr = strings.TrimPrefix(lineStr, "data: ")
		}

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
