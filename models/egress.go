package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

//TableName alarm
func (t *Egress) TableName() string {
	return TableName("egress")
}

type Egress struct {
	ID        int64     `orm:"column(id);auto" json:"id"`
	NameOne   string    `orm:"column(name_one);size(255)" json:"name_one"`
	InOne     string    `orm:"column(in_one);size(255)" json:"in_one"`
	OutOne    string    `orm:"column(out_one);size(255)" json:"out_one"`
	NameTwo   string    `orm:"column(name_two);size(255)" json:"name_two"`
	InTwo     string    `orm:"column(in_two);size(255)" json:"in_two"`
	OutTwo    string    `orm:"column(out_two);size(255)" json:"out_two"`
	CreatedAt time.Time `orm:"column(created_at);type(datetime);auto_now_add" json:"created_at"`
	Status    int       `orm:"column(status);size(200)" json:"status"`
}

//get id
func GetEgress() (v *Egress, err error) {
	o := orm.NewOrm()
	v = &Egress{ID: 1}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

//get all
func UpdateEgress(m *Egress) (err error) {
	o := orm.NewOrm()
	v := Egress{ID: m.ID}
	err = o.Read(&v)
	if err != nil {
		return err
	}
	v.NameOne = m.NameOne
	v.InOne = m.InOne
	v.OutOne = m.OutOne
	v.NameTwo = m.NameTwo
	v.InTwo = m.InTwo
	v.OutTwo = m.OutTwo
	m.CreatedAt = v.CreatedAt
	v.Status = 1
	_, err = o.Update(m, "NameOne", "InOne", "OutOne",
		"NameTwo", "InTwo", "OutTwo",
		"CreatedAt", "Status")
	if err != nil {
		return err
	}
	return nil
}
