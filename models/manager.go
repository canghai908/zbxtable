package models

import (
	"errors"
	"time"

	"github.com/astaxie/beego/orm"
	"zbxtable/utils"
)

//Auth struct
type Auth struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

//Chpwd struct aa
type Chpwd struct {
	Name   string `json:"name"`
	Oldpwd string `json:"oldpwd"`
	Pwd1   string `json:"pwd1"`
	Pwd2   string `json:"pwd2"`
}

//ManagerInfo struct
type ManagerInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID       int       `json:"id"`
		Username string    `json:"username"`
		Avatar   string    `json:"avatar"`
		Status   int64     `json:"status"`
		Role     string    `json:"role"`
		Created  time.Time `json:"created"`
	} `json:"data"`
}

//Token struct
type Token struct {
	Token string `json:"token"`
}

//Manager struct
type Manager struct {
	ID       int       `orm:"column(id);auto" json:"id"`
	Username string    `orm:"column(username);size(255)" json:"username"`
	Password string    `orm:"column(password);size(255)" json:"password"`
	Avatar   string    `orm:"column(avatar);size(255)" json:"avatar"`
	Status   int64     `orm:"column(status)" json:"status"`
	Role     string    `orm:"column(role);size(255)" json:"role"`
	Created  time.Time `orm:"column(created);type(datetime)" json:"created"`
}

//TableName string
func (t *Manager) TableName() string {
	return TableName("manager")
}

// AddManager insert a new Manager into database and returns
// last inserted Id on success.
func AddManager(m *Manager) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
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

//Chanagepwd func
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
