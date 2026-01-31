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

// ReceiveGin 接收消息（Gin版本）
func ReceiveGin(c *gin.Context) {
	type Re struct {
		ID  int64  `json:"id"`
		Msg string `json:"msg"`
	}
	var res Re

	if c.Request.Method != "POST" {
		res.ID = 0
		res.Msg = "method is not allowed for the requested url."
		response.Success(c, res)
		return
	}

	tenantid := c.GetHeader("ZBX-TenantID")
	token := c.GetHeader("Token")
	fmt.Println(c.Request.Header)
	// 多租户 token 校验：优先使用租户绑定表；未配置绑定时回退到全局 token（兼容旧逻辑）
	zabbixInstanceID := 0
	fmt.Println(tenantid)
	if binding, err := model.GetZabbixTenantByTenantID(tenantid); err == nil && binding != nil {
		fmt.Println("aaaa")
		if !binding.Enabled {
			res.ID = 0
			res.Msg = "Tenant Disabled!"
			logger.Log.Error("Tenant Disabled!")
			response.Success(c, res)
			return
		}
		if binding.Token != "" && token != binding.Token {
			res.ID = 0
			res.Msg = "Token Error!"
			response.Success(c, res)
			return
		}
		zabbixInstanceID = binding.ID
	} else {
		if token != model.GetConfKey("token") {
			res.ID = 0
			res.Msg = "Token Error!"
			logger.Log.Error("Token Error!")
			response.Success(c, res)
			return
		}
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		res.ID = 0
		res.Msg = err.Error()
		logger.Log.Error(err.Error())
		response.Success(c, res)
		return
	}
	fmt.Println(string(body))

	id, err := model.MsAdd(tenantid, zabbixInstanceID, body)
	if err != nil {
		res.ID = 0
		res.Msg = err.Error()
		response.Success(c, res)
		return
	}

	res.ID = id
	res.Msg = "successed"
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
