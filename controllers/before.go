package controllers

import (
	"github.com/astaxie/beego/logs"
	"io"
	"net/http/pprof"

	//
	"strconv"
	"time"
	"zbxtable/utils"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	jwtbeego "github.com/canghai908/jwt-beego"
	"zbxtable/models"
)

// BeforeUserController sd
type BeforeUserController struct {
	beego.Controller
}

// Manager var
var Manager = new(models.Manager)

// manager var
var manager models.Manager

// BeforeUserController  dd
// @Title 登录
// @Description 登录
// @Param	body	body 	models.Userlogin	true	"登录"
// @Success 200 login success
// @Failure 403 user not exist
// @router /login [post]
func (u *BeforeUserController) Login() {
	var res models.Auth
	err := json.Unmarshal(u.Ctx.Input.RequestBody, &manager)
	if err != nil {
		res.Code = 400
		res.Message = "用户名或密码错误"
		u.Data["json"] = res
		u.ServeJSON()
	}
	//session timeout
	var SessionTimeout int64
	SessionTimeout, err = strconv.ParseInt(models.GetConfKey("timeout"), 10, 32)
	//SessionTimeout=
	if err != nil {
		logs.Error(err)
		SessionTimeout = 12
	}
	o := orm.NewOrm()
	//find one
	err = o.QueryTable(Manager).Filter("username", manager.Username).One(Manager)
	if err != nil {
		res.Code = 400
		res.Message = "用户名或密码错误"
		u.Data["json"] = res
		u.ServeJSON()
	}
	if Manager.Status == 1 {
		res.Code = 403
		res.Message = "用户已被禁用"
		u.Data["json"] = res
		u.ServeJSON()
	}
	//bcrypt encrypt
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
		u.Data["json"] = res
		u.ServeJSON()
	}
	//md5 encrypt
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
		u.Data["json"] = res
		u.ServeJSON()
	}
	res.Code = 400
	res.Message = "用户名或密码错误"
	u.Data["json"] = res
	u.ServeJSON()
}

// Logout controller
// @Title 注销
// @Description 注销
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
	}
	token := u.Ctx.Request.Header.Get("Token")
	if token != models.GetConfKey("token") {
		res.ID = 0
		res.Msg = "Token Error!"
		u.Data["json"] = res
		u.ServeJSON()
	}
	tenantid := u.Ctx.Request.Header.Get("ZBX-TenantID")
	mess := u.Ctx.Request.Body
	defer mess.Close()
	body, err := io.ReadAll(mess)
	if err != nil {
		res.ID = 0
		res.Msg = err.Error()
		u.Data["json"] = res
		u.ServeJSON()
	}
	id, err := models.MsAdd(tenantid, body)
	if err != nil {
		res.ID = 0
		res.Msg = err.Error()
		u.Data["json"] = res
		u.ServeJSON()
	}
	res.ID = id
	res.Msg = "successed"
	u.Data["json"] = res
	u.ServeJSON()
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
	if tok != models.GetConfKey("token") {
		u.Data["json"] = "Token Error!"
		u.ServeJSON()
		return
	}
	mess := u.Ctx.Request.Body
	defer mess.Close()
	body, err := io.ReadAll(mess)
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

type ProfController struct {
	beego.Controller
}

func (this *ProfController) Get() {
	switch this.Ctx.Input.Param(":app") {
	default:
		pprof.Index(this.Ctx.ResponseWriter, this.Ctx.Request)
	case "":
		pprof.Index(this.Ctx.ResponseWriter, this.Ctx.Request)
	case "cmdline":
		pprof.Cmdline(this.Ctx.ResponseWriter, this.Ctx.Request)
	case "profile":
		pprof.Profile(this.Ctx.ResponseWriter, this.Ctx.Request)
	case "symbol":
		pprof.Symbol(this.Ctx.ResponseWriter, this.Ctx.Request)
	}
	this.Ctx.ResponseWriter.WriteHeader(200)
}
