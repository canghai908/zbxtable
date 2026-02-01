package model

import (
	"errors"
	"strconv"
	"zbxtable/pkg/utils"
)

const (
	RuleCust = iota + 1
	RuleDefault
)

// AddRule one
func AddRule(m *Rule) (id int64, err error) {
	m.Sweek = utils.VAarToStr(m.Sweek)
	m.ZID = utils.VAarToStr(m.ZID)
	m.Channel = utils.VAarToStr(m.Channel)
	m.UserIds = utils.VAarToStr(m.UserIds)
	m.GroupIds = utils.VAarToStr(m.GroupIds)
	err = DB.Create(m).Error
	if err != nil {
		return 0, err
	}
	return int64(m.ID), nil
}

// GetRuleByID one
func GetRuleByID(id int) (v *Rule, err error) {
	v = &Rule{}
	err = DB.Where("id = ?", id).First(v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

func GetRule(page, limit, name, zid, m_type, status string) (cnt int64, userlist []Rule, err error) {
	var rules []Rule
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	if limits == 0 {
		limits = 10
	}
	if pages == 0 {
		pages = 1
	}

	query := DB.Model(&Rule{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if zid != "" {
		query = query.Where("zid = ?", zid)
	}
	if m_type == "" {
		query = query.Where("m_type IN ?", []string{"1", "2"})
	} else {
		query = query.Where("m_type = ?", m_type)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	err = query.Count(&cnt).Error
	if err != nil {
		return 0, []Rule{}, err
	}

	// 获取分页数据
	offset := (pages - 1) * limits
	err = query.Select("id", "name", "conditions", "zid", "note",
		"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
		Limit(limits).Offset(offset).Find(&rules).Error
	if err != nil {
		return 0, []Rule{}, err
	}
	return cnt, rules, nil
}

// UpdateRuleStatus rule
func UpdateRuleStatus(m *Rule, tuser string) error {
	//role检查
	var p Manager
	err := DB.Where("username = ?", tuser).First(&p).Error
	if err != nil {
		return err
	}
	//not admin role return err
	if p.Role != "admin" {
		return errors.New("no permission")
	}
	//update
	err = DB.Model(&Rule{}).Where("id = ?", m.ID).Update("status", m.Status).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdateRule rule
func UpdateRule(m *Rule, tuser string) error {
	//role检查
	var p Manager
	err := DB.Where("username = ?", tuser).First(&p).Error
	if err != nil {
		return err
	}
	//not admin role return err
	if p.Role != "admin" {
		return errors.New("no permission")
	}
	//vue array to str
	m.ZID = utils.VAarToStr(m.ZID)
	m.Sweek = utils.VAarToStr(m.Sweek)
	m.Channel = utils.VAarToStr(m.Channel)
	m.UserIds = utils.VAarToStr(m.UserIds)
	m.GroupIds = utils.VAarToStr(m.GroupIds)
	//判断是不是修改默认规则，如果是修改默认规则，则默认规则类型不变,表达式，条件配置
	var mType, mConditions, mInstanceID string
	if m.MType == "2" {
		mType = "2"
		mConditions = ""
		mInstanceID = "*"
	} else {
		mType = m.MType
		mConditions = m.Conditions
		mInstanceID = m.ZID
	}
	err = DB.Model(&Rule{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"name":       m.Name,
		"zid":        mInstanceID,
		"conditions": mConditions,
		"s_week":     m.Sweek,
		"m_type":     mType,
		"s_time":     m.Stime,
		"e_time":     m.Etime,
		"channel":    m.Channel,
		"user_ids":   m.UserIds,
		"group_ids":  m.GroupIds,
		"note":       m.Note,
		"status":     m.Status,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// DeleteRule status
func DeleteRule(id int, tuser string) (err error) {
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
	var v Rule
	err = DB.Where("id = ?", id).First(&v).Error
	if err == nil {
		//default 规则不能删除
		if v.MType == "2" {
			return errors.New("默认规则不能删除")
		}
		err = DB.Delete(&Rule{}, id).Error
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
