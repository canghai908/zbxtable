package model

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
type User struct {
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
	DingTalkSecret string    `gorm:"column:ding_talk_secret;size:255" json:"ding_talk_secret"`
	FeishuRobotKey string    `gorm:"column:feishu_robot_key;size:255" json:"feishu_robot_key"`
	FeishuSecret   string    `gorm:"column:feishu_secret;size:255" json:"feishu_secret"`
	WebhookURL     string    `gorm:"column:webhook_url;size:500" json:"webhook_url"`
	WebhookToken   string    `gorm:"column:webhook_token;size:255" json:"webhook_token"`
	Theme          string    `gorm:"column:theme;type:text" json:"theme,omitempty"`
	Created        time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Updated        time.Time `gorm:"column:updated;autoUpdateTime" json:"updated_at"`
}

// TableName string
func (t *User) TableName() string {
	return TableName("user")
}

func encryptUserSecrets(m *User) error {
	encryptionKey := GetEncryptionKey()
	targets := []struct {
		label string
		val   *string
	}{
		{"企业微信群机器人Key", &m.WechatRobotKey},
		{"钉钉机器人Key", &m.DingTalk},
		{"钉钉加签密钥", &m.DingTalkSecret},
		{"飞书群机器人Key", &m.FeishuRobotKey},
		{"飞书签名密钥", &m.FeishuSecret},
		{"Webhook令牌", &m.WebhookToken},
	}
	for _, t := range targets {
		if *t.val == "" {
			continue
		}
		encrypted, err := utils.EncryptString(*t.val, encryptionKey)
		if err != nil {
			return errors.New("加密" + t.label + "失败: " + err.Error())
		}
		*t.val = encrypted
	}
	return nil
}

// AddManager insert a new Manager into database and returns
// last inserted Id on success.
func AddUser(m *User) (id int64, err error) {
	// 检查用户是否已存在
	var existing User
	result := DB.Where("username = ?", m.Username).First(&existing)
	if result.Error == nil {
		return 0, errors.New("用户已存在")
	}

	if err := encryptUserSecrets(m); err != nil {
		return 0, err
	}

	// 插入
	result = DB.Create(m)
	if result.Error != nil {
		return 0, result.Error
	}
	return int64(m.ID), nil
}

// UpdateUserTheme 更新指定用户名的主题配置（仅允许改自己的 theme）。
func UpdateUserTheme(username, theme string) error {
	if username == "" {
		return errors.New("username is empty")
	}
	if theme == "" {
		return errors.New("theme is empty")
	}
	return DB.Model(&User{}).Where("username = ?", username).Update("theme", theme).Error
}

// UpdateUser 更新用户信息（仅管理员可改密码 / 角色 / 他人资料）
func UpdateUser(m *User, tuser string) error {
	//role检查
	var p User
	err := DB.Where("username = ?", tuser).First(&p).Error
	if err != nil {
		return err
	}
	if p.Role != "admin" {
		return errors.New("no permission")
	}

	// 构建更新字段映射，只更新非空字段
	updates := map[string]interface{}{}

	// 密码字段
	if m.Password != "" {
		updates["password"] = m.Password
	}

	// 角色字段
	if m.Role != "" {
		updates["role"] = m.Role
	}

	// 联系方式字段 - 允许更新为空字符串（清空）
	if m.Email != "" {
		updates["email"] = m.Email
	}
	if m.Phone != "" {
		updates["phone"] = m.Phone
	}
	if m.Wechat != "" {
		updates["wechat"] = m.Wechat
	}
	encryptionKey := GetEncryptionKey()
	secretColumns := []struct {
		label  string
		column string
		value  string
	}{
		{"企业微信群机器人Key", "wechat_robot_key", m.WechatRobotKey},
		{"钉钉机器人Key", "ding_talk", m.DingTalk},
		{"钉钉加签密钥", "ding_talk_secret", m.DingTalkSecret},
		{"飞书群机器人Key", "feishu_robot_key", m.FeishuRobotKey},
		{"飞书签名密钥", "feishu_secret", m.FeishuSecret},
		{"Webhook令牌", "webhook_token", m.WebhookToken},
	}
	for _, sc := range secretColumns {
		if sc.value == "" || sc.value == "********" {
			continue
		}
		encryptedVal, err := utils.EncryptString(sc.value, encryptionKey)
		if err != nil {
			return errors.New("加密" + sc.label + "失败: " + err.Error())
		}
		updates[sc.column] = encryptedVal
	}
	if m.WebhookURL != "" {
		updates["webhook_url"] = m.WebhookURL
	}

	// 主题字段
	if m.Theme != "" {
		updates["theme"] = m.Theme
	}

	// 如果没有任何字段需要更新，直接返回
	if len(updates) == 0 {
		return nil
	}

	err = DB.Model(&User{}).Where("id = ?", m.ID).Updates(updates).Error
	if err != nil {
		return err
	}
	return nil
}

// udpate user
func UpdateUserStatus(m *User, tuser string) error {
	//user
	var v User
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
	err = DB.Model(&User{}).Where("id = ?", m.ID).Update("status", m.Status).Error
	if err != nil {
		return err
	}
	return nil
}

// GetAllAlarm retrieves all Alarm matches certain condition. Returns empty list if
// no records exist
func GetUser(page, limit, tuser, username, status string) (cnt int64, userlist []User, err error) {
	var users []User
	var countUsers []User
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)

	// 构建查询
	query := DB.Model(&User{})

	// 管理员角色检查
	var p User
	err = DB.Where("username = ?", tuser).First(&p).Error
	if err != nil {
		return 0, []User{}, err
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
		return 0, []User{}, err
	}
	cnt = int64(len(countUsers))

	// 分页查询
	offset := (pages - 1) * limits
	err = query.Select("id", "username", "role", "avatar", "email", "ding_talk",
		"ding_talk_secret", "feishu_robot_key", "feishu_secret", "webhook_url", "webhook_token",
		"phone", "created", "status", "wechat", "wechat_robot_key").
		Limit(limits).Offset(offset).Find(&users).Error
	if err != nil {
		return 0, []User{}, err
	}

	// 对敏感字段进行脱敏处理
	for i := range users {
		if users[i].WechatRobotKey != "" {
			users[i].WechatRobotKey = "********"
		}
		if users[i].DingTalk != "" {
			users[i].DingTalk = "********"
		}
		if users[i].DingTalkSecret != "" {
			users[i].DingTalkSecret = "********"
		}
		if users[i].FeishuRobotKey != "" {
			users[i].FeishuRobotKey = "********"
		}
		if users[i].FeishuSecret != "" {
			users[i].FeishuSecret = "********"
		}
		if users[i].WebhookToken != "" {
			users[i].WebhookToken = "********"
		}
	}

	return cnt, users, nil
}

// GetManagerByID retrieves Manager by Id. Returns error if
// Id doesn't exist
func GetManagerByID(id int) (v *User, err error) {
	v = &User{}
	err = DB.First(v, id).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetManagerByName retrieves User by Username. Returns error if
// Id doesn't exist
func GetManagerByName(username string) (v *User, err error) {
	v = &User{}
	err = DB.Where("username = ?", username).First(v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Chanagepwd func
func Chanagepwd(old, new string) (err error) {
	return ChangePasswordForUser("admin", old, new)
}

func ChangePasswordForUser(username, oldPassword, newPassword string) error {
	v, err := GetManagerByName(username)
	if err != nil {
		return err
	}
	if err := utils.ComparePass(v.Password, oldPassword); err != nil {
		return errors.New("账号或密码错误")
	}
	password, err := utils.PasswordHash(newPassword)
	if err != nil {
		return errors.New("更新密码出错")
	}
	if err := DB.Model(&User{}).Where("id = ?", v.ID).Update("password", password).Error; err != nil {
		return errors.New("更新密码出错")
	}
	return nil
}

func DeleteUser(id int, tuser string) (err error) {
	//role检查
	var p User
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
	var v User
	err = DB.First(&v, id).Error
	if err == nil {
		if v.Username == tuser {
			return errors.New("cannot delete myself")
		}
		err = DB.Delete(&User{}, id).Error
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
