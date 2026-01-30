package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
	models "zbxtable/internal/model"
	"zbxtable/pkg/logger"
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
		http.Error(c.Writer, "Not a websocket handshake", 400)
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
			val, err := models.GetTopologyById(id)
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
			err = models.UpdateEdgeDataById(id)
			if err != nil {
				continue
			}
		}
		time.Sleep(time.Second * 10)
	}
}

// LoginGin 登录（Gin版本）
func LoginGin(c *gin.Context) {
	var res models.Auth
	var manager models.Manager

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		res.Code = 400
		res.Message = "请求体读取失败"
		c.JSON(http.StatusOK, res)
		return
	}

	err = json.Unmarshal(body, &manager)
	if err != nil {
		res.Code = 400
		res.Message = "用户名或密码错误"
		c.JSON(http.StatusOK, res)
		return
	}

	// session timeout
	var SessionTimeout int64
	SessionTimeout, err = strconv.ParseInt(models.GetConfKey("timeout"), 10, 32)
	if err != nil {
		logger.Log.Error(err)
		SessionTimeout = 12
	}

	// 使用 GORM 查询
	var Manager models.Manager
	err = models.GetDB().Where("username = ?", manager.Username).First(&Manager).Error
	if err != nil {
		res.Code = 400
		res.Message = "用户名或密码错误"
		c.JSON(http.StatusOK, res)
		return
	}

	if Manager.Status == 1 {
		res.Code = 403
		res.Message = "用户已被禁用"
		c.JSON(http.StatusOK, res)
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
		res.Code = 200
		res.Message = "登录成功"
		res.Data.Token = tokenString
		res.Data.User.ID = Manager.ID
		res.Data.User.Name = Manager.Username
		res.Data.User.Avatar = Manager.Avatar
		res.Data.User.Role = Manager.Role
		res.Data.User.Created = Manager.Created
		res.Data.Roles = []models.Roles{{ID: Manager.Role, Operation: Manager.Operation}}
		c.JSON(http.StatusOK, res)
		return
	}

	// md5 encrypt
	if utils.Md5([]byte(manager.Password)) == Manager.Password {
		et := jwtbeego.EasyToken{
			Username: Manager.Username,
			Expires:  time.Now().Add(time.Hour * time.Duration(SessionTimeout)).Unix(),
		}
		tokenString, _ := et.GetToken()
		res.Code = 200
		res.Message = "登录成功"
		res.Data.Token = tokenString
		res.Data.User.Name = Manager.Username
		res.Data.User.Avatar = Manager.Avatar
		res.Data.User.Created = Manager.Created
		res.Data.Roles = []models.Roles{{ID: Manager.Role, Operation: Manager.Operation}}
		c.JSON(http.StatusOK, res)
		return
	}

	res.Code = 400
	res.Message = "用户名或密码错误"
	c.JSON(http.StatusOK, res)
}

// LogoutGin 注销（Gin版本）
func LogoutGin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "用户注销成功",
	})
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
		c.JSON(http.StatusOK, res)
		return
	}

	tenantid := c.GetHeader("ZBX-TenantID")
	token := c.GetHeader("Token")
	fmt.Println(c.Request.Header)
	// 多租户 token 校验：优先使用租户绑定表；未配置绑定时回退到全局 token（兼容旧逻辑）
	zabbixInstanceID := 0
	fmt.Println(tenantid)
	if binding, err := models.GetZabbixTenantByTenantID(tenantid); err == nil && binding != nil {
		fmt.Println("aaaa")
		if !binding.Enabled {
			res.ID = 0
			res.Msg = "Tenant Disabled!"
			logger.Log.Error("Tenant Disabled!")
			c.JSON(http.StatusOK, res)
			return
		}
		if binding.Token != "" && token != binding.Token {
			res.ID = 0
			res.Msg = "Token Error!"
			c.JSON(http.StatusOK, res)
			return
		}
		zabbixInstanceID = binding.ID
	} else {
		if token != models.GetConfKey("token") {
			res.ID = 0
			res.Msg = "Token Error!"
			logger.Log.Error("Token Error!")
			c.JSON(http.StatusOK, res)
			return
		}
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		res.ID = 0
		res.Msg = err.Error()
		logger.Log.Error(err.Error())
		c.JSON(http.StatusOK, res)
		return
	}
	fmt.Println(string(body))

	id, err := models.MsAdd(tenantid, zabbixInstanceID, body)
	if err != nil {
		res.ID = 0
		res.Msg = err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.ID = id
	res.Msg = "successed"
	c.JSON(http.StatusOK, res)
}

// WebhookGin Webhook（Gin版本）
func WebhookGin(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.JSON(http.StatusOK, "method is not allowed for the requested url.")
		return
	}

	tok := c.GetHeader("Token")
	if tok != models.GetConfKey("token") {
		c.JSON(http.StatusOK, "Token Error!")
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, err.Error())
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
		c.JSON(http.StatusOK, "message format is error!")
		return
	}

	type Re struct {
		ID  int64  `json:"id"`
		Msg string `json:"msg"`
	}
	var res Re
	res.ID = 132
	res.Msg = "ok"
	c.JSON(http.StatusOK, res)
}
