package controllers

import (
	"github.com/astaxie/beego"
	jwtbeego "github.com/canghai908/jwt-beego"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

var json = jsoniter.Config{
	EscapeHTML:             false,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}.Froze()

// BaseController use
type BaseController struct {
	beego.Controller
}

func (c *BaseController) ServeJSON() {
	c.Ctx.Output.Header("Content-Type", "application/json; charset=utf-8")
	var err error
	data := c.Data["json"]
	encoder := json.NewEncoder(c.Ctx.Output.Context.ResponseWriter)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(data)
	if err != nil {
		http.Error(c.Ctx.Output.Context.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Resp all
type Resp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Items  string `json:"items"`
		Totals int    `json:"totals"`
	} `json:"data"`
}

// Res var
var Res = &Resp{}

// Tuser is userinfo
var Tuser string

// Prepare login
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
		var res resp
		res.Code = 50014
		res.Message = "token过期或非法的token"
		c.Data["json"] = res
		c.ServeJSON()
	}
	Tuser = iss
	return
}
