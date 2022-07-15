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

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (c *WebSocketController) Get() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	ws, err := wsupgrader.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(c.Ctx.ResponseWriter, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		logs.Error("Cannot setup WebSocket connection:", err)
		return
	}
	defer ws.Close()

	for {
		val, err := models.GetTopologyById(id)
		if err != nil {
			logs.Debug(err)
			time.Sleep(time.Second * 10)
			continue
		}
		msg, _ := json.Marshal(val)
		err = ws.WriteMessage(1, msg)
		if err != nil {
			logs.Debug(err)
		}
		time.Sleep(time.Second * 10)
	}
}

//func HandleMessages() {
//	for {
//		msg := <-broadcast
//		fmt.Println(msg)
//		fmt.Println("clients len ", len(clients))
//		for client := range clients {
//			p, me, err := client.
//			if err != nil {
//				log.Println(p, string(me))
//				log.Println(msg)
//				log.Printf("client.WriteJSON error: %v", err)
//				client.Close()
//				delete(clients, client)
//				//	break
//			}
//		}
//	}
//}
