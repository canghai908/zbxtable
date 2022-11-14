package models

import (
	"errors"
	"github.com/astaxie/beego/orm"
	"strconv"
	"strings"
	"time"
)

type UserGroup struct {
	ID      int       `orm:"column(id);auto" json:"id"`
	Name    string    `orm:"column(name);size(255)" json:"name"`
	Member  string    `orm:"column(member);size(1000)" json:"member"`
	Note    string    `orm:"column(note);size(255)" json:"note"`
	Created time.Time `orm:"column(created);type(datetime);auto_now_add" json:"created"`
	Updated time.Time `orm:"column(updated);type(datetime);auto_now" json:"updated_at"`
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
	o := orm.NewOrm()
	//用户是否已存在
	p := UserGroup{Name: m.Name}
	err = o.Read(&p, "name")
	if err == nil {
		return 0, errors.New("用户组存在")
	}
	//插入
	_, err = o.Insert(m)
	if err != nil {
		return 0, err
	}
	return
}

// GetAllAlarm retrieves all Alarm matches certain condition. Returns empty list if
// no records exist
func GetGroup(page, limit, tuser, name string) (cnt int64, userlist []UserGroup, err error) {
	o := orm.NewOrm()
	var groups []UserGroup
	var CountGroups []UserGroup
	al := new(UserGroup)
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	//count alarms
	cond := orm.NewCondition()
	if name != "" {
		cond = cond.And("name__icontains", name)
	}
	//管理员角色
	p := Manager{Username: tuser}
	err = o.Read(&p, "username")
	if err != nil {
		return 0, []UserGroup{}, err
	}
	if p.Role != "admin" {
		return 0, []UserGroup{}, err
	}
	_, err = o.QueryTable(al).SetCond(cond).
		All(&CountGroups)
	_, err = o.QueryTable(al).
		Limit(limits, (pages-1)*limits).SetCond(cond).
		All(&groups, "id", "name", "member", "note", "created", "updated")
	if err != nil {
		return 0, []UserGroup{}, err
	}
	cnt = int64(len(CountGroups))
	return cnt, groups, nil
}

//udpate user
func UpdateUserGroup(m *UserGroup, tuser string) error {
	o := orm.NewOrm()
	//role检查
	p := Manager{Username: tuser}
	err := o.Read(&p, "username")
	if err != nil {
		return err
	}
	if p.Role != "admin" {
		return errors.New("no permission")
	}
	//
	v := UserGroup{ID: m.ID}
	err = o.Read(&v)
	if err != nil {
		return err
	}
	//更新其他字段
	v.Name = m.Name
	v.Note = m.Note
	_, err = o.Update(m, "name", "note")
	if err != nil {
		return err
	}
	return nil
}

//UpdateGroupMember user
func UpdateGroupMember(m *UserGroup, tuser string) error {
	o := orm.NewOrm()
	//role检查
	p := Manager{Username: tuser}
	err := o.Read(&p, "username")
	if err != nil {
		return err
	}
	if p.Role != "admin" {
		return errors.New("no permission")
	}
	//
	v := UserGroup{ID: m.ID}
	err = o.Read(&v)
	if err != nil {
		return err
	}
	//更新字段
	new1 := strings.Replace(m.Member, "[", "", -1)
	new2 := strings.Replace(new1, "]", "", -1)
	new3 := strings.TrimSuffix(strings.Replace(new2, `"`, ``, -1), `,`)
	m.Member = new3
	v.Member = m.Member
	_, err = o.Update(m, "member")
	if err != nil {
		return err
	}
	return nil
}

func DeleteGroup(id int, tuser string) (err error) {
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
	v := UserGroup{ID: id}
	//admin not delete
	//if id == 1 {
	//	return errors.New("admin user cannot delete ")
	//}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		//if v.Username == tuser {
		//	return errors.New("cannot delete myself")
		//}
		_, err = o.Delete(&UserGroup{ID: id})
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
