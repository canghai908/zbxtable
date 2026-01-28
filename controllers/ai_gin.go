package controllers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"zbxtable/models"

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
		c.JSON(http.StatusOK, gin.H{"code": 500, "message": "请求体读取失败"})
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

	// 构造系统提示词
	systemPrompt := struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}{
		Role:    "system",
		Content: "你是一个专业的运维分析师，请分析用户提供的问题并给出专业的建议。",
	}

	// 从配置或系统配置表获取Ollama地址
	ollamaHost := models.GetConfigValueByKey("ollama_host", models.GetConfKey("ollama_host"))
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}
	ollamaModel := models.GetConfigValueByKey("ollama_model", models.GetConfKey("ollama_model"))
	if ollamaModel == "" {
		ollamaModel = "deepseek-r1:32b"
	}

	// 构造Ollama请求
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
				Content: userReq.Message,
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

	// 发送请求到Ollama
	resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(jsonData))
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
	defer resp.Body.Close()

	// 创建reader
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
