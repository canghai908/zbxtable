package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

// TableName alarm
func (t *Config) TableName() string {
	return TableName("config")
}

type Config struct {
	ID        int64     `orm:"column(id);auto" json:"id"`
	Name      string    `orm:"column(name);size(255)" json:"name"`
	Key       string    `orm:"column(key);size(255)" json:"key"`
	Value     string    `orm:"column(value);size(255)" json:"value"`
	Comment   string    `orm:"column(comment);size(255)" json:"comment"`
	CreatedAt time.Time `orm:"column(created_at);type(datetime);auto_now_add" json:"created_at"`
	UpdatedAt time.Time `orm:"column(updated_at);type(datetime);auto_now" json:"updated_at"`
}

// GetConfigList 获取系统配置
func GetConfigList() ([]Config, error) {
	o := orm.NewOrm()
	var v []Config
	_, err := o.QueryTable(Config{}).OrderBy("ID").All(&v)
	if err != nil {
		return nil, err
	}
	return v, err
}

// GetConfigOne 获取某一个配置
func GetConfigOne(id int64) (v *Config, err error) {
	o := orm.NewOrm()
	v = &Config{ID: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// UpdateConfig 更新系统配置
func UpdateConfig(m *Config) (err error) {
	o := orm.NewOrm()
	v := Config{ID: m.ID}
	//数据面板独立配置
	err = o.Read(&v)
	if err != nil {
		return err
	}
	v.Value = m.Value
	_, err = o.Update(m, "Value")
	if err != nil {
		return err
	}
	//dash处理
	if m.ID == 1 {
		err = updateZbxDash(m)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

// 数据面板控制
func updateZbxDash(m *Config) (err error) {
	o := orm.NewOrm()
	menu := Menu{Router: "dash"}
	if o.Read(&menu, "Router") == nil {
		switch m.Value {
		case "0":
			menu.IsAvailable = true
		case "1":
			menu.IsAvailable = false
		}
	}
	_, err = o.Update(&menu, "IsAvailable")
	if err != nil {
		return err
	}
	return nil
}
