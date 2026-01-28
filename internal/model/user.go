package models

import (
	"errors"
	"strconv"
	"time"
	"zbxtable/pkg/utils"
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
	ID             int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username       string    `gorm:"column:username;size:255;uniqueIndex" json:"username"`
	Password       string    `gorm:"column:password;size:255" json:"password,omitempty"`
	Avatar         string    `gorm:"column:avatar;size:255" json:"avatar"`
	Status         int64     `gorm:"column:status" json:"status"`
	Role           string    `gorm:"column:role;size:255" json:"role"`
	Operation      string    `gorm:"column:operation;size:255" json:"operation"`
	Email          string    `gorm:"column:email;size:255" json:"email"`
	Wechat         string    `gorm:"column:wechat;size:255" json:"wechat"`
	WechatRobotKey string    `gorm:"column:wechat_robot_key;size:255" json:"wechat_robot_key"`
	Phone          string    `gorm:"column:phone;size:255" json:"phone"`
	DingTalk       string    `gorm:"column:ding_talk;size:255" json:"ding_talk"`
	Created        time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Updated        time.Time `gorm:"column:updated;autoUpdateTime" json:"updated_at"`
}

// TableName string
func (t *Manager) TableName() string {
	return TableName("manager")
}

// AddManager insert a new Manager into database and returns
// last inserted Id on success.
func AddUser(m *Manager) (id int64, err error) {
	// 检查用户是否已存在
	var existing Manager
	result := DB.Where("username = ?", m.Username).First(&existing)
	if result.Error == nil {
		return 0, errors.New("用户已存在")
	}
	// 插入
	result = DB.Create(m)
	if result.Error != nil {
		return 0, result.Error
	}
	return int64(m.ID), nil
}

// UpdateUser 更新用户信息
func UpdateUser(m *Manager, tuser string) error {
	//role检查
	var p Manager
	err := DB.Where("username = ?", tuser).First(&p).Error
	if err != nil {
		return err
	}
	if p.Role != "admin" && m.Role == "admin" {
		return errors.New("no permission")
	}
	
	//密码不为空更新密码
	if m.Password != "" {
		updates := map[string]interface{}{
			"password":         m.Password,
			"role":             m.Role,
			"email":            m.Email,
			"phone":            m.Phone,
			"wechat":           m.Wechat,
			"wechat_robot_key": m.WechatRobotKey,
			"ding_talk":        m.DingTalk,
			"status":           m.Status,
		}
		err = DB.Model(&Manager{}).Where("id = ?", m.ID).Updates(updates).Error
		if err != nil {
			return err
		}
		return nil
	}
	//更新其他字段
	updates := map[string]interface{}{
		"role":             m.Role,
		"email":            m.Email,
		"phone":            m.Phone,
		"wechat":           m.Wechat,
		"wechat_robot_key": m.WechatRobotKey,
		"ding_talk":        m.DingTalk,
	}
	err = DB.Model(&Manager{}).Where("id = ?", m.ID).Updates(updates).Error
	if err != nil {
		return err
	}
	return nil
}

// udpate user
func UpdateUserStatus(m *Manager, tuser string) error {
	//user
	var v Manager
	err := DB.First(&v, m.ID).Error
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
	err = DB.Model(&Manager{}).Where("id = ?", m.ID).Update("status", m.Status).Error
	if err != nil {
		return err
	}
	return nil
}

// GetAllAlarm retrieves all Alarm matches certain condition. Returns empty list if
// no records exist
func GetUser(page, limit, tuser, username, status string) (cnt int64, userlist []Manager, err error) {
	var users []Manager
	var countUsers []Manager
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	
	// 构建查询
	query := DB.Model(&Manager{})
	
	// 管理员角色检查
	var p Manager
	err = DB.Where("username = ?", tuser).First(&p).Error
	if err != nil {
		return 0, []Manager{}, err
	}
	if p.Role != "admin" {
		query = query.Where("username = ?", tuser)
	}
	
	// 条件过滤
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	// 计数
	err = query.Find(&countUsers).Error
	if err != nil {
		return 0, []Manager{}, err
	}
	cnt = int64(len(countUsers))
	
	// 分页查询
	offset := (pages - 1) * limits
	err = query.Select("id", "username", "role", "avatar", "email", "ding_talk",
		"phone", "created", "status", "wechat", "wechat_robot_key").
		Limit(limits).Offset(offset).Find(&users).Error
	if err != nil {
		return 0, []Manager{}, err
	}
	
	return cnt, users, nil
}

// GetManagerByID retrieves Manager by Id. Returns error if
// Id doesn't exist
func GetManagerByID(id int) (v *Manager, err error) {
	v = &Manager{}
	err = DB.First(v, id).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetManagerByName retrieves User by Username. Returns error if
// Id doesn't exist
func GetManagerByName(username string) (v *Manager, err error) {
	v = &Manager{}
	err = DB.Where("username = ?", username).First(v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Chanagepwd func
func Chanagepwd(old, new string) (err error) {
	v, err := GetManagerByName("admin")
	if err != nil {
		return err
	}
	if v.Username != "admin" || v.Password != utils.Md5([]byte(old)) {
		return errors.New("账号或密码错误")
	}
	err = DB.Model(&Manager{}).Where("id = ?", v.ID).Update("password", utils.Md5([]byte(new))).Error
	if err != nil {
		return errors.New("更新密码出错")
	}
	return nil
}

func DeleteUser(id int, tuser string) (err error) {
	//role检查
	var p Manager
	err = DB.Where("username = ?", tuser).First(&p).Error
	if err != nil {
		return err
	}
	//not admin role return err
	if p.Role != "admin" {
		return errors.New("no permission")
	}
	//admin not delete
	if id == 1 {
		return errors.New("admin user cannot delete ")
	}
	// ascertain id exists in the database
	var v Manager
	err = DB.First(&v, id).Error
	if err == nil {
		if v.Username == tuser {
			return errors.New("cannot delete myself")
		}
		err = DB.Delete(&Manager{}, id).Error
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
