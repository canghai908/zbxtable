package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	model "zbxtable/internal/model"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/response"
	"zbxtable/pkg/utils"

	jwtbeego "github.com/canghai908/jwt-beego"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.Config{
	EscapeHTML:             false,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}.Froze()

// WebSocketHandlerGin WebSocket 处理（Gin版本）
func WebSocketHandlerGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	// 使用与 websocket.go 相同的 upgrader 配置
	upgrader := websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 10 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		response.BadRequest(c, "Not a websocket handshake")
		return
	} else if err != nil {
		logger.Log.Error("Cannot setup WebSocket connection:", err)
		return
	}
	defer ws.Close()

	for {
		// 读取数据
		_, ms, err := ws.ReadMessage()
		if err != nil {
			logger.Log.Debug(err)
			break
		}
		// 发送数据
		if string(ms) == "success" {
			// 查询数据
			val, err := model.GetTopologyById(id)
			if err != nil {
				logger.Log.Debug(err)
				continue
			}
			// write
			msg, _ := json.Marshal(val)
			err = ws.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				logger.Log.Debug(err)
				continue
			}
			// 更新数据
			err = model.UpdateEdgeDataById(id)
			if err != nil {
				continue
			}
		}
		time.Sleep(time.Second * 10)
	}
}

// PublicWebSocketHandlerGin 公开的 WebSocket 处理（无需认证，仅限已共享的拓扑）
func PublicWebSocketHandlerGin(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	logger.Log.Info("PublicWebSocketHandlerGin called, id:", idStr)
	logger.Log.Info("Request Headers:", c.Request.Header)
	logger.Log.Info("Request URL:", c.Request.URL.String())
	logger.Log.Info("Request Method:", c.Request.Method)

	// 检查拓扑是否已共享
	topo, err := model.GetTopologyById(id)
	if err != nil {
		logger.Log.Error("GetTopologyById failed:", err)
		// 返回 HTTP 错误，不进行 WebSocket 升级
		// c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "拓扑不存在"})
		return
	}
	logger.Log.Info("Topology found, status:", topo.Status)

	if topo.Status != "1" {
		logger.Log.Error("Topology not shared, status:", topo.Status)
		// 返回 HTTP 错误，不进行 WebSocket 升级
		//	c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "该拓扑未共享，无法公开访问"})
		return
	}

	logger.Log.Info("Starting WebSocket upgrade...")

	// 使用与 websocket.go 相同的 upgrader 配置
	upgrader := websocket.Upgrader{
		ReadBufferSize:   10240,
		WriteBufferSize:  10240,
		HandshakeTimeout: 10 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			logger.Log.Info("CheckOrigin called, Origin:", r.Header.Get("Origin"))
			return true
		},
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		logger.Log.Error("WebSocket handshake error:", err)
		return
	} else if err != nil {
		logger.Log.Error("Cannot setup WebSocket connection:", err)
		return
	}
	defer ws.Close()

	logger.Log.Info("WebSocket connection established successfully")

	for {
		// 读取数据
		_, ms, err := ws.ReadMessage()
		if err != nil {
			logger.Log.Debug("ReadMessage error:", err)
			break
		}
		logger.Log.Info("Received message:", string(ms))
		fmt.Println("AAAAAAAAAA", string(ms))
		// 发送数据
		if string(ms) == "success" {
			// 查询数据
			val, err := model.GetTopologyById(id)
			if err != nil {
				logger.Log.Debug("GetTopologyById in loop error:", err)
				continue
			}
			fmt.Println(val)
			// 再次检查状态
			if val.Status != "1" {
				logger.Log.Debug("拓扑已取消共享")
				break
			}
			// write
			msg, _ := json.Marshal(val)
			logger.Log.Info("Sending topology data, size:", len(msg))
			err = ws.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				logger.Log.Debug("WriteMessage error:", err)
				continue
			}
			logger.Log.Info("Topology data sent successfully")

			// 更新数据
			err = model.UpdateEdgeDataById(id)
			if err != nil {
				logger.Log.Debug("UpdateEdgeDataById error:", err)
				continue
			}
		}
		time.Sleep(time.Second * 10)
	}

	logger.Log.Info("WebSocket connection closed")
}

// LoginGin 登录（Gin版本）
func LoginGin(c *gin.Context) {
	var manager model.Manager

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "请求体读取失败")
		return
	}

	err = json.Unmarshal(body, &manager)
	if err != nil {
		response.BadRequest(c, "用户名或密码错误")
		return
	}

	// session timeout
	var SessionTimeout int64
	SessionTimeout, err = strconv.ParseInt(model.GetConfKey("timeout"), 10, 32)
	if err != nil {
		logger.Log.Error(err)
		SessionTimeout = 12
	}

	// 使用 GORM 查询
	var Manager model.Manager
	err = model.GetDB().Where("username = ?", manager.Username).First(&Manager).Error
	if err != nil {
		response.BadRequest(c, "用户名或密码错误")
		return
	}

	if Manager.Status == 1 {
		response.Forbidden(c, "用户已被禁用")
		return
	}

	// bcrypt encrypt
	err = utils.ComparePass(Manager.Password, manager.Password)
	if err == nil {
		et := jwtbeego.EasyToken{
			Username: Manager.Username,
			Expires:  time.Now().Add(time.Hour * time.Duration(SessionTimeout)).Unix(),
		}
		tokenString, _ := et.GetToken()
		response.SuccessWithMessage(c, "登录成功", gin.H{
			"token": tokenString,
			"user": gin.H{
				"id":      Manager.ID,
				"name":    Manager.Username,
				"avatar":  Manager.Avatar,
				"role":    Manager.Role,
				"created": Manager.Created,
			},
			"roles": []gin.H{
				{
					"id":        Manager.Role,
					"operation": Manager.Operation,
				},
			},
		})
		return
	}

	// md5 encrypt
	if utils.Md5([]byte(manager.Password)) == Manager.Password {
		et := jwtbeego.EasyToken{
			Username: Manager.Username,
			Expires:  time.Now().Add(time.Hour * time.Duration(SessionTimeout)).Unix(),
		}
		tokenString, _ := et.GetToken()
		response.SuccessWithMessage(c, "登录成功", gin.H{
			"token": tokenString,
			"user": gin.H{
				"name":    Manager.Username,
				"avatar":  Manager.Avatar,
				"created": Manager.Created,
			},
			"roles": []gin.H{
				{
					"id":        Manager.Role,
					"operation": Manager.Operation,
				},
			},
		})
		return
	}

	response.BadRequest(c, "用户名或密码错误")
}

// LogoutGin 注销（Gin版本）
func LogoutGin(c *gin.Context) {
	response.SuccessWithMessage(c, "用户注销成功", nil)
}

// ReceiveGin 接收消息告警消息
func ReceiveGin(c *gin.Context) {
	type Re struct {
		ID  int64  `json:"id"`
		Msg string `json:"msg"`
	}
	var res Re
	if c.Request.Method != "POST" {
		res.ID = 0
		res.Msg = "method is not allowed for the requested url."
		response.BadRequest(c, res.Msg)
		return
	}
	//老系统使用的是ZBX-TenantID
	TenantID := c.GetHeader("ZBX-TenantID")

	token := c.GetHeader("X-Token")
	var InstanceID string
	//如果为空可能是新webhook
	if TenantID == "" {
		//新系统使用的是X-Instance
		InstanceID = c.GetHeader("X-Instance")
	}
	if InstanceID == "" {
		response.BadRequest(c, "instanceID not found")
	}
	// instance的token校验
	//查询instanceid和token,可以一起查询，无需多次查询
	//
	instance, err := model.GetZabbixInstanceByInstanceID(InstanceID)
	if err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	//实例是否启用
	if !instance.Enabled {
		res.ID = 0
		res.Msg = "Instance Disabled!"
		response.Success(c, res)
		return
	}
	//token是否正确
	if token == "" || (instance.Token != "" && token != instance.Token) {
		res.ID = 0
		res.Msg = "Token Error!"
		response.Success(c, res)
		return

	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		res.ID = 0
		res.Msg = err.Error()
		logger.Log.Error(err.Error())
		response.Success(c, res)
		return
	}
	id, err := model.MsAdd(instance.ID, body)
	if err != nil {
		res.ID = 0
		res.Msg = err.Error()
		response.Success(c, res)
		return
	}

	res.ID = id
	res.Msg = "success"
	response.Success(c, res)
}

// WebhookGin Webhook（Gin版本）
func WebhookGin(c *gin.Context) {
	if c.Request.Method != "POST" {
		response.BadRequest(c, "method is not allowed for the requested url.")
		return
	}

	tok := c.GetHeader("Token")
	if tok != model.GetConfKey("token") {
		response.Unauthorized(c, "Token Error!")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	type Message struct {
		Hostname    string `json:"Hostname"`
		Group       string `json:"Group"`
		EventTime   string `json:"EventTime"`
		Severity    string `json:"Severity"`
		TriggerName string `json:"TriggerName"`
		TriggerKey  string `json:"TriggerKey"`
		ItemName    string `json:"ItemName"`
		Status      string `json:"Status"`
		EventID     string `json:"EventID"`
	}
	var b Message
	err = json.Unmarshal(body, &b)
	if err != nil {
		response.BadRequest(c, "message format is error!")
		return
	}

	type Re struct {
		ID  int64  `json:"id"`
		Msg string `json:"msg"`
	}
	var res Re
	res.ID = 132
	res.Msg = "ok"
	response.Success(c, res)
}
