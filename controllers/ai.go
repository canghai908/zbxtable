package controllers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/astaxie/beego"
	"zbxtable/models"
)

type AIController struct {
	BaseController
}

type ChatRequest struct {
	Message string `json:"message"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
}

// AIRes 统一响应格式
var AIRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

// URLMapping ...
func (c *AIController) URLMapping() {
	c.Mapping("Chat", c.Chat)
}

// Chat ...
// @Title 流式聊天接口
// @Description 调用Ollama API进行流式对话
// @Param   message    body    string  true    "聊天消息内容"
// @Success 200 {object} ChatResponse
// @router /chat [post]
func (c *AIController) Chat() {
	// 设置响应头
	c.Ctx.ResponseWriter.Header().Set("Content-Type", "text/event-stream")
	c.Ctx.ResponseWriter.Header().Set("Cache-Control", "no-cache")
	c.Ctx.ResponseWriter.Header().Set("Connection", "keep-alive")

	// 解析请求体
	var userReq ChatRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &userReq); err != nil {
		AIRes.Code = 500
		AIRes.Message = err.Error()
		c.Data["json"] = AIRes
		c.ServeJSON()
		return
	}

	// 构造系统提示词
	systemPrompt := Message{
		Role:    "system",
		Content: "你是一个专业的运维分析师，请分析用户提供的问题并给出专业的建议。",
	}

	// 构造Ollama请求
	ollamaReq := struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
		Stream   bool      `json:"stream"`
	}{
		Model:  models.GetConfigValueByKey("ollama_model", beego.AppConfig.DefaultString("ollama_model", "deepseek-r1:32b")),
		Stream: true,
		Messages: []Message{
			systemPrompt,
			{
				Role:    "user",
				Content: userReq.Message,
			},
		},
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		AIRes.Code = 500
		AIRes.Message = err.Error()
		c.Data["json"] = AIRes
		c.ServeJSON()
		return
	}

	// 从配置或系统配置表获取Ollama地址
	ollamaHost := models.GetConfigValueByKey("ollama_host", beego.AppConfig.DefaultString("ollama_host", "http://localhost:11434"))
	ollamaURL := fmt.Sprintf("%s/api/chat", ollamaHost)

	// 发送请求到Ollama
	resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		AIRes.Code = 500
		AIRes.Message = err.Error()
		c.Data["json"] = AIRes
		c.ServeJSON()
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
		var response ChatResponse
		if err := json.Unmarshal(line, &response); err != nil {
			continue
		}

		// 发送数据
		if response.Message.Content != "" {
			c.Ctx.WriteString(response.Message.Content)
			c.Ctx.ResponseWriter.Flush()
		}

		// 如果响应完成，退出循环
		if response.Done {
			break
		}
	}
}
