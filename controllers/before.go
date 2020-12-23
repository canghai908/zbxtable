package controllers

import (
	"github.com/astaxie/beego/logs"
	"io/ioutil"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	jwtbeego "github.com/canghai908/jwt-beego"
	"github.com/canghai908/zbxtable/models"
	"github.com/canghai908/zbxtable/utils"
)

// BeforeUserController sd
type BeforeUserController struct {
	beego.Controller
}

//Manager var
var Manager = new(models.Manager)

//manager var
var manager models.Manager

//res var
//var res *models.ManagerInfo = new(models.ManagerInfo)
var res = &models.ManagerInfo{}

// Login controller
// @Title Login
// @Description Logs user into the system
// @Param	body		body 	models.Userlogin	true	"body for user content"
// @Success 200 login success
// @Failure 403 user not exist
// @router /login [post]
func (u *BeforeUserController) Login() {
	err := json.Unmarshal(u.Ctx.Input.RequestBody, &Manager)
	if err != nil {
		u.Data["json"] = map[string]int64{"code": 110}
		u.ServeJSON()
		return
	}
	var SessionTimeout int
	SessionTimeout, err = beego.AppConfig.Int("timeout")
	if err != nil {
		logs.Error(err)
		SessionTimeout = 12
	}
	o := orm.NewOrm()
	err = o.QueryTable(Manager).Filter("username", Manager.Username).Filter("password", utils.Md5([]byte(Manager.Password))).One(Manager)
	var res models.Auth
	if err != nil {
		res.Code = 400
		res.Message = "用户名或密码错误"
	} else {
		et := jwtbeego.EasyToken{
			Username: Manager.Username,
			Expires:  time.Now().Add(time.Hour * time.Duration(SessionTimeout)).Unix(),
		}
		tokenString, _ := et.GetToken()
		res.Code = 200
		res.Message = "登录成功"
		res.Data.Token = tokenString
	}
	u.Data["json"] = res
	u.ServeJSON()
}

// Logout controller
// @Title Logout
// @Description Logs out current logged in user session
// @Param	body		body 	models.Userlogin	true	"body for user content"
// @Success 200 {string} logout success
// @router /logout [post]
func (u *BeforeUserController) Logout() {
	Res.Code = 200
	Res.Message = "用户注销成功"
	u.Data["json"] = Res
	u.ServeJSON()
}

// Receive message controller
// @Title Recevie
// @Description Logs out current logged in user session
// @Param	body		body 	models.Userlogin	true	"body for user content"
// @Success 200 {string} logout success
// @router /receive [post]
func (u *BeforeUserController) Receive() {
	type Re struct {
		ID  int64  `json:"id"`
		Msg string `json:"msg"`
	}
	var res Re
	if !u.Ctx.Input.IsPost() {
		res.ID = 0
		res.Msg = "method is not allowed for the requested url."
		u.Data["json"] = res
		u.ServeJSON()
		return
	}
	token := u.Ctx.Request.Header.Get("Token")
	tenantid := u.Ctx.Request.Header.Get("ZBX-TenantID")
	mess := u.Ctx.Request.Body
	defer mess.Close()
	body, err := ioutil.ReadAll(mess)
	if err != nil {
		res.ID = 0
		res.Msg = err.Error()
		u.Data["json"] = res
		u.ServeJSON()
		return
	}
	id, err := models.MsAdd(token, tenantid, body)
	if err != nil {
		res.ID = 0
		res.Msg = "Please run 'zbxtable ua' to update action."
		u.Data["json"] = res
		u.ServeJSON()
		return
	}
	res.ID = id
	res.Msg = "successed"
	u.Data["json"] = res
	u.ServeJSON()
	return
}

// Webhook receive message controller
// @Title Recevie
// @Description Logs out current logged in user session
// @Param	body		body 	models.Userlogin	true	"body for user content"
// @Success 200 {string} logout success
// @router /webhook [post]
func (u *BeforeUserController) Webhook() {
	if !u.Ctx.Input.IsPost() {
		u.Data["json"] = "method is not allowed for the requested url."
		u.ServeJSON()
		return
	}
	tok := u.Ctx.Request.Header.Get("Token")
	if tok != beego.AppConfig.String("token") {
		u.Data["json"] = "Token Error!"
		u.ServeJSON()
		return
	}
	mess := u.Ctx.Request.Body
	defer mess.Close()
	body, err := ioutil.ReadAll(mess)
	if err != nil {
		u.Data["json"] = err.Error()
		u.ServeJSON()

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
		u.Data["json"] = "message format is error!"
		u.ServeJSON()
		return
	}
	type Re struct {
		ID  int64  `json:"id"`
		Msg string `json:"msg"`
	}
	var res Re
	res.ID = 132
	res.Msg = "ok"
	u.Data["json"] = res
	u.ServeJSON()
}
