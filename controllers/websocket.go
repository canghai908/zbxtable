package controllers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"time"
	"zbxtable/models"
)

type WebSocketController struct {
	beego.Controller
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *WebSocketController) Get() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	ws, err := wsUpgrader.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(c.Ctx.ResponseWriter, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		logs.Error("Cannot setup WebSocket connection:", err)
		return
	}
	defer ws.Close()
	for {
		//读取数据
		_, ms, err := ws.ReadMessage()
		if err != nil {
			logs.Debug(err)
			break
		}
		//发送数据
		if string(ms) == "success" {
			//查询数据
			val, err := models.GetTopologyById(id)
			if err != nil {
				logs.Debug(err)
				continue
			}
			//write
			msg, _ := json.Marshal(val)
			err = ws.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				logs.Debug(err)
				continue
			}
			//更新数据
			err = models.UpdateEdgeDataById(id)
			if err != nil {
				continue
			}

		}
		time.Sleep(time.Second * 10)
	}
}
