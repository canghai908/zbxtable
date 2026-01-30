package model

import (
	"time"
)

// TableName alarm
func (t *Egress) TableName() string {
	return TableName("egress")
}

type Egress struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	NameOne   string    `gorm:"column:name_one;size:255" json:"name_one"`
	InOne     string    `gorm:"column:in_one;size:255" json:"in_one"`
	OutOne    string    `gorm:"column:out_one;size:255" json:"out_one"`
	NameTwo   string    `gorm:"column:name_two;size:255" json:"name_two"`
	InTwo     string    `gorm:"column:in_two;size:255" json:"in_two"`
	OutTwo    string    `gorm:"column:out_two;size:255" json:"out_two"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	Status    int       `gorm:"column:status" json:"status"`
}

// get id
func GetEgress() (v *Egress, err error) {
	v = &Egress{}
	err = DB.Where("id = ?", 1).First(v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// get all
func UpdateEgress(m *Egress) (err error) {
	var v Egress
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err != nil {
		return err
	}
	m.CreatedAt = v.CreatedAt
	m.Status = 1
	err = DB.Model(&Egress{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"name_one":   m.NameOne,
		"in_one":     m.InOne,
		"out_one":    m.OutOne,
		"name_two":   m.NameTwo,
		"in_two":     m.InTwo,
		"out_two":    m.OutTwo,
		"created_at": m.CreatedAt,
		"status":     m.Status,
	}).Error
	if err != nil {
		return err
	}
	return nil
}
