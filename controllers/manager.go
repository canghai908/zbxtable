package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	jwtbeego "github.com/canghai908/jwt-beego"
	"github.com/canghai908/zbxtable/models"
)

// ManagerController operations for Manager
type ManagerController struct {
	BaseController
}

// URLMapping ...
func (c *ManagerController) URLMapping() {
	c.Mapping("Info", c.Info)
	c.Mapping("Chpwd", c.Chpwd)
}

// GetOne ...
// @Title Get One
// @Description get Manager by id
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.Manager
// @Failure 403 :id is empty
// @router /:id [get]
func (c *ManagerController) GetOne() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v, err := models.GetManagerByID(id)
	if err != nil {
		c.Data["json"] = err.Error()
	} else {
		c.Data["json"] = v
	}
	c.ServeJSON()
}

// Info for manager controller
// @Title Manager info
// @Description Logs Manaager into the system
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	body    	body 	models.Token	true		"The Token"
// @Success 200 login success
// @Failure 403 manager not exist
// @router /info [post]
func (c *ManagerController) Info() {
	var t models.Token
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &t); err == nil {

		et := jwtbeego.EasyToken{}
		valid, iss, _ := et.ValidateToken(t.Token)
		if !valid {
			type resp struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Data    struct {
					ID       int       `json:"id"`
					Username string    `json:"username"`
					Avatar   string    `json:"avatar"`
					Status   int       `json:"status"`
					Role     string    `json:"role"`
					Created  time.Time `json:"created"`
				} `json:"data"`
			}
			c.Ctx.Output.SetStatus(401)
			var res resp
			res.Code = 50014
			res.Message = "token过期或非法的token"
			c.Data["json"] = res

		} else {
			v, err := models.GetManagerByName(iss)
			if err != nil {
				c.Ctx.Output.SetStatus(401)
				res.Code = 50014
				res.Message = "token过期或非法的token"
				c.Data["json"] = res
			} else {
				c.Ctx.Output.SetStatus(200)
				res.Code = 200
				res.Message = "登录成功"
				res.Data.ID = v.ID
				res.Data.Username = v.Username
				res.Data.Avatar = v.Avatar
				res.Data.Role = v.Role
				res.Data.Status = v.Status
				res.Data.Created = v.Created
				c.Data["json"] = res

			}

		}
	} else {
		c.Data["json"] = err.Error()
	}
	c.ServeJSON()
}

// Chpwd for manager controller
// @Title Chpwd
// @Change Manager Password
// @Param	X-Token		header  string			true		"x-token in header"
// @Param	body		body 	models.Chpwd	true		"body for Manager content"
// @Success 201 {int} models.Chpwd
// @Failure 403 body is empty
// @router /chpwd [post]
func (c *ManagerController) Chpwd() {
	var v models.Chpwd
	type resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			ID       int       `json:"id"`
			Username string    `json:"username"`
			Avatar   string    `json:"avatar"`
			Status   int       `json:"status"`
			Role     string    `json:"role"`
			Created  time.Time `json:"created"`
		} `json:"data"`
	}
	var res resp
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		fmt.Print(&v)
		if err := models.Chanagepwd(&v); err == nil {
			c.Ctx.Output.SetStatus(201)
			res.Code = 200
			res.Message = "修改密码成功"
		} else {
			res.Code = 50014
			res.Message = err.Error()
		}
	} else {
		res.Code = 50014
		res.Message = err.Error()
	}
	c.Data["json"] = res
	c.ServeJSON()
}
