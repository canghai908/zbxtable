package controllers

import (
	"github.com/astaxie/beego"
	jwtbeego "github.com/canghai908/jwt-beego"
)

//BaseController use
type BaseController struct {
	beego.Controller
}

//Resp Resp all
type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Items  string `json:"items"`
		Totals int    `json:"totals"`
	} `json:"data"`
}

//Res var
var Res = &Resp{}

//API for zabbix
//var API = &zabbix.API{}

//Tuser is userinfo
var Tuser string

//Prepare Prepare login
func (c *BaseController) Prepare() {
	tokenString := c.Ctx.Request.Header.Get("X-Token")
	et := jwtbeego.EasyToken{}
	valid, iss, _ := et.ValidateToken(tokenString)
	if !valid {
		type resp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Items string `json:"items"`
				Total int    `json:"total"`
			} `json:"data"`
		}
		//this.Ctx.Output.SetStatus(401)
		var res resp
		res.Code = 50014
		res.Message = "token过期或非法的token"
		c.Data["json"] = res
		c.ServeJSON()
	}
	Tuser = iss
	return
}
