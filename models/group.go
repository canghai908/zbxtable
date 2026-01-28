package models

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type UserGroup struct {
	ID      int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name    string    `gorm:"column:name;size:255" json:"name"`
	Member  string    `gorm:"column:member;size:1000" json:"member"`
	Note    string    `gorm:"column:note;size:255" json:"note"`
	Created time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Updated time.Time `gorm:"column:updated;autoUpdateTime" json:"updated_at"`
}

//ManagerInfo struct
type GroupResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

//TableName string
func (t *UserGroup) TableName() string {
	return TableName("user_group")
}

// AddManager insert a new Manager into database and returns
// last inserted Id on success.
func AddUserGroup(m *UserGroup) (id int64, err error) {
	//用户是否已存在
	var p UserGroup
	err = DB.Where("name = ?", m.Name).First(&p).Error
	if err == nil {
		return 0, errors.New("用户组存在")
	}
	//插入
	err = DB.Create(m).Error
	if err != nil {
		return 0, err
	}
	return int64(m.ID), nil
}

// GetAllAlarm retrieves all Alarm matches certain condition. Returns empty list if
// no records exist
func GetGroup(page, limit, tuser, name string) (cnt int64, userlist []UserGroup, err error) {
	var groups []UserGroup
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	if limits == 0 {
		limits = 10
	}
	if pages == 0 {
		pages = 1
	}

	//管理员角色
	var p Manager
	err = DB.Where("username = ?", tuser).First(&p).Error
	if err != nil {
		return 0, []UserGroup{}, err
	}
	if p.Role != "admin" {
		return 0, []UserGroup{}, errors.New("no permission")
	}

	query := DB.Model(&UserGroup{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// 获取总数
	err = query.Count(&cnt).Error
	if err != nil {
		return 0, []UserGroup{}, err
	}

	// 获取分页数据
	offset := (pages - 1) * limits
	err = query.Select("id", "name", "member", "note", "created", "updated").
		Limit(limits).Offset(offset).Find(&groups).Error
	if err != nil {
		return 0, []UserGroup{}, err
	}
	return cnt, groups, nil
}

//udpate user
func UpdateUserGroup(m *UserGroup, tuser string) error {
	//role检查
	var p Manager
	err := DB.Where("username = ?", tuser).First(&p).Error
	if err != nil {
		return err
	}
	if p.Role != "admin" {
		return errors.New("no permission")
	}
	err = DB.Model(&UserGroup{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"name": m.Name,
		"note": m.Note,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

//UpdateGroupMember user
func UpdateGroupMember(m *UserGroup, tuser string) error {
	//role检查
	var p Manager
	err := DB.Where("username = ?", tuser).First(&p).Error
	if err != nil {
		return err
	}
	if p.Role != "admin" {
		return errors.New("no permission")
	}
	//更新字段
	new1 := strings.Replace(m.Member, "[", "", -1)
	new2 := strings.Replace(new1, "]", "", -1)
	new3 := strings.TrimSuffix(strings.Replace(new2, `"`, ``, -1), `,`)
	m.Member = new3
	err = DB.Model(&UserGroup{}).Where("id = ?", m.ID).Update("member", m.Member).Error
	if err != nil {
		return err
	}
	return nil
}

func DeleteGroup(id int, tuser string) (err error) {
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
	err = DB.Delete(&UserGroup{}, id).Error
	if err != nil {
		return err
	}
	return nil
}
