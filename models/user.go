package models

import (
	"errors"
	"strconv"
	"time"
	"zbxtable/utils"

	"github.com/astaxie/beego/orm"
)

// Auth struct
type Auth struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string    `json:"token"`
		User  LoginUser `json:"user"`
		Roles []Roles   `json:"roles"`
	} `json:"data"`
}

type LoginUser struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Avatar  string    `json:"avatar"`
	Status  int64     `json:"status"`
	Role    string    `json:"role"`
	Created time.Time `json:"created"`
}

type Roles struct {
	ID        string `json:"id"`
	Operation string `json:"operation"`
}

// Chpwd struct aa
type Chpwd struct {
	Name   string `json:"name"`
	Oldpwd string `json:"oldpwd"`
	Pwd1   string `json:"pwd1"`
	Pwd2   string `json:"pwd2"`
}

// ManagerInfo struct
type UserResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

// ManagerInfo struct
type ManagerInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID        int       `json:"id"`
		Username  string    `json:"username"`
		Avatar    string    `json:"avatar"`
		Status    int64     `json:"status"`
		Role      string    `json:"role"`
		Operation string    `json:"operation"`
		Created   time.Time `json:"created"`
	} `json:"data"`
}

// Token struct
type Token struct {
	Token string `json:"token"`
}

// Manager struct
type Manager struct {
	ID        int       `orm:"column(id);auto" json:"id"`
	Username  string    `orm:"column(username);size(255)" json:"username"`
	Password  string    `orm:"column(password);size(255)" json:"password,omitempty"`
	Avatar    string    `orm:"column(avatar);size(255)" json:"avatar"`
	Status    int64     `orm:"column(status)" json:"status"`
	Role      string    `orm:"column(role);size(255)" json:"role"`
	Operation string    `orm:"column(operation);size(255)" json:"operation"`
	Email     string    `orm:"column(email);size(255)" json:"email"`
	Wechat    string    `orm:"column(wechat);size(255)" json:"wechat"`
	Phone     string    `orm:"column(phone);size(255)" json:"phone"`
	DingTalk  string    `orm:"column(ding_talk);size(255)" json:"ding_talk"`
	Created   time.Time `orm:"column(created);type(datetime);auto_now_add" json:"created"`
	Updated   time.Time `orm:"column(updated);type(datetime);auto_now" json:"updated_at"`
}

// TableName string
func (t *Manager) TableName() string {
	return TableName("manager")
}

// AddManager insert a new Manager into database and returns
// last inserted Id on success.
func AddUser(m *Manager) (id int64, err error) {
	o := orm.NewOrm()
	//用户是否已存在
	p := Manager{Username: m.Username}
	err = o.Read(&p, "username")
	if err == nil {
		return 0, errors.New("用户已存在")
	}
	//插入
	_, err = o.Insert(m)
	if err != nil {
		return 0, err
	}
	return
}

// UpdateUser 更新用户信息
func UpdateUser(m *Manager, tuser string) error {
	o := orm.NewOrm()
	//role检查
	p := Manager{Username: tuser}
	err := o.Read(&p, "username")
	if err != nil {
		return err
	}
	if p.Role != "admin" && m.Role == "admin" {
		return errors.New("no permission")
	}
	//
	v := Manager{ID: m.ID}
	err = o.Read(&v)
	if err != nil {
		return err
	}

	//密码不为空更新密码
	if m.Password != "" {
		v.Password = m.Password
		v.Role = m.Role
		v.Email = m.Email
		v.Phone = m.Phone
		v.Wechat = m.Wechat
		v.DingTalk = m.DingTalk
		v.Status = m.Status
		_, err = o.Update(m, "Password", "Role", "Email", "Wechat", "Phone", "DingTalk")
		if err != nil {
			return err
		}
		return nil
	}
	//更新其他字段
	v.Role = m.Role
	v.Email = m.Email
	v.Phone = m.Phone
	v.Wechat = m.Wechat
	v.DingTalk = m.DingTalk
	_, err = o.Update(m, "Role", "Email", "Wechat", "Phone", "DingTalk")
	if err != nil {
		return err
	}
	return nil
}

// udpate user
func UpdateUserStatus(m *Manager, tuser string) error {
	o := orm.NewOrm()
	//user
	v := Manager{ID: m.ID}
	err := o.Read(&v)
	if err != nil {
		return err
	}
	//admin user
	if v.Username == "admin" && v.Status == 0 {
		return errors.New("cannot disable admin")
	}
	//selft
	if v.Username == tuser {
		return errors.New("cannot disable self")
	}
	v.Status = m.Status
	_, err = o.Update(m, "Status")
	if err != nil {
		return err
	}
	return nil
}

// GetAllAlarm retrieves all Alarm matches certain condition. Returns empty list if
// no records exist
func GetUser(page, limit, tuser, username, status string) (cnt int64, userlist []Manager, err error) {
	o := orm.NewOrm()
	var users []Manager
	var CountUsers []Manager
	al := new(Manager)
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	//count alarms
	cond := orm.NewCondition()
	if username != "" {
		cond = cond.And("username__icontains", username)
	}
	if status != "" {
		cond = cond.And("status", status)
	}
	//管理员角色
	p := Manager{Username: tuser}
	err = o.Read(&p, "username")
	if err != nil {
		return 0, []Manager{}, err
	}
	if p.Role != "admin" {
		cond = cond.And("username", tuser)
	}
	_, err = o.QueryTable(al).SetCond(cond).
		All(&CountUsers)
	_, err = o.QueryTable(al).
		Limit(limits, (pages-1)*limits).SetCond(cond).
		All(&users, "id", "username", "role", "avatar", "email", "ding_talk",
			"phone", "created", "status", "wechat")
	if err != nil {
		return 0, []Manager{}, err
	}
	cnt = int64(len(CountUsers))
	return cnt, users, nil
}

// GetManagerByID retrieves Manager by Id. Returns error if
// Id doesn't exist
func GetManagerByID(id int) (v *Manager, err error) {
	o := orm.NewOrm()
	v = &Manager{ID: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetManagerByName retrieves User by Username. Returns error if
// Id doesn't exist
func GetManagerByName(username string) (v *Manager, err error) {
	o := orm.NewOrm()
	v = &Manager{Username: username}
	if err = o.Read(v, "Username"); err == nil {
		return v, nil
	}
	return nil, err
}

// Chanagepwd func
func Chanagepwd(old, new string) (err error) {
	o := orm.NewOrm()
	v, err := GetManagerByName("admin")
	if err != nil {
		return err
	}
	if v.Username != "admin" || v.Password != utils.Md5([]byte(old)) {
		return errors.New("账号或密码错误")
	}
	v.Password = utils.Md5([]byte(new))
	_, err = o.Update(v)
	if err != nil {
		return errors.New("更新密码出错")
	}
	return nil
}

func DeleteUser(id int, tuser string) (err error) {
	o := orm.NewOrm()
	//role检查
	p := Manager{Username: tuser}
	err = o.Read(&p, "username")
	if err != nil {
		return err
	}
	//not admin role return err
	if p.Role != "admin" {
		return errors.New("no permission")
	}
	v := Manager{ID: id}
	//admin not delete
	if id == 1 {
		return errors.New("admin user cannot delete ")
	}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		if v.Username == tuser {
			return errors.New("cannot delete myself")
		}
		_, err = o.Delete(&Manager{ID: id})
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
