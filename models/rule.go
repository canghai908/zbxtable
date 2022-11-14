package models

import (
	"errors"
	"github.com/astaxie/beego/orm"
	"strconv"
	"zbxtable/utils"
)

const (
	RuleCust = iota + 1
	RuleDefault
)

// AddRule one
func AddRule(m *Rule) (id int64, err error) {
	o := orm.NewOrm()
	m.Sweek = utils.VAarToStr(m.Sweek)
	m.TenantID = utils.VAarToStr(m.TenantID)
	m.Channel = utils.VAarToStr(m.Channel)
	m.UserIds = utils.VAarToStr(m.UserIds)
	m.GroupIds = utils.VAarToStr(m.GroupIds)
	id, err = o.Insert(m)
	if err != nil {
		return 0, err
	}
	return id, nil
}

//GetRuleByID one
func GetRuleByID(id int) (v *Rule, err error) {
	o := orm.NewOrm()
	v = &Rule{ID: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

func GetRule(page, limit, name, tenant_id, m_type, status string) (cnt int64, userlist []Rule, err error) {
	o := orm.NewOrm()
	var rules []Rule
	var CountRules []Rule
	al := new(Rule)
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	//count alarms
	cond := orm.NewCondition()
	if name != "" {
		cond = cond.And("name__icontains", name)
	}
	if tenant_id != "" {
		cond = cond.And("tenant_id", tenant_id)
	}
	if m_type == "" {
		cond = cond.And("m_type", "1").Or("m_type", "2")
	} else {
		cond = cond.And("m_type", m_type)
	}
	if status != "" {
		cond = cond.And("status", status)
	}
	_, err = o.QueryTable(al).SetCond(cond).
		All(&CountRules)
	_, err = o.QueryTable(al).
		Limit(limits, (pages-1)*limits).SetCond(cond).
		All(&rules, "id", "name", "conditions", "tenant_id", "note",
			"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created")
	if err != nil {
		return 0, []Rule{}, err
	}
	cnt = int64(len(CountRules))
	return cnt, rules, nil
}

//UpdateRuleStatus rule
func UpdateRuleStatus(m *Rule, tuser string) error {
	o := orm.NewOrm()
	//role检查
	p := Manager{Username: tuser}
	err := o.Read(&p, "username")
	if err != nil {
		return err
	}
	//not admin role return err
	if p.Role != "admin" {
		return errors.New("no permission")
	}
	//update
	v := Rule{ID: m.ID}
	err = o.Read(&v)
	if err != nil {
		return err
	}
	v.Status = m.Status
	_, err = o.Update(m, "Status")
	if err != nil {
		return err
	}
	return nil
}

//UpdateRule rule
func UpdateRule(m *Rule, tuser string) error {
	o := orm.NewOrm()
	//role检查
	p := Manager{Username: tuser}
	err := o.Read(&p, "username")
	if err != nil {
		return err
	}
	//not admin role return err
	if p.Role != "admin" {
		return errors.New("no permission")
	}
	//update
	v := Rule{ID: m.ID}
	err = o.Read(&v)
	if err != nil {
		return err
	}
	//vue array to str
	m.TenantID = utils.VAarToStr(m.TenantID)
	m.Sweek = utils.VAarToStr(m.Sweek)
	m.Channel = utils.VAarToStr(m.Channel)
	m.UserIds = utils.VAarToStr(m.UserIds)
	m.GroupIds = utils.VAarToStr(m.GroupIds)
	_, err = o.Update(m, "name", "tenant_id", "conditions", "s_week", "m_type",
		"s_time", "e_time", "channel", "user_ids", "group_ids", "note", "status", "status")
	if err != nil {
		return err
	}
	return nil
}

//DeleteRule status
func DeleteRule(id int, tuser string) (err error) {
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
	v := Rule{ID: id}

	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		//default 规则不能删除
		if v.MType == "2" {
			return errors.New("默认规则不能删除")
		}
		_, err = o.Delete(&Rule{ID: id})
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
