package models

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"strings"
	"time"
)

//item link to inventory
//link https://www.zabbix.com/documentation/current/en/manual/api/reference/host/object#host-inventory
const (
	CPUCore             = 16
	CPUUtilizationID    = 18
	MemoryUtilizationID = 19
	MemoryTotalID       = 20
	UptimeID            = 22
	Model               = 29
)

//TableName alarm
func (t *System) TableName() string {
	return TableName("system")
}

//get id
func GetSystemByID(id int64) (v *System, err error) {
	o := orm.NewOrm()
	v = &System{ID: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

//get all
func GetALlSystem() (cnt int64, system []System, err error) {
	o := orm.NewOrm()
	var sys []System
	al := new(System)
	_, err = o.QueryTable(al).All(&sys)
	if err != nil {
		return 0, []System{}, err
	}
	cnt = int64(len(sys))
	return cnt, sys, nil
}

//get all
func UpdateSystem(m *System) (err error) {
	o := orm.NewOrm()
	v := System{ID: m.ID}
	err = o.Read(&v)
	if err != nil {
		return err
	}
	v.CPUCore = m.CPUCore
	v.CPUUtilizationID = m.CPUUtilizationID
	v.GroupID = m.GroupID
	v.MemoryTotalID = m.MemoryTotalID
	v.MemoryUsedID = m.MemoryUsedID
	v.MemoryUtilizationID = m.MemoryUtilizationID
	v.UptimeID = m.UptimeID
	v.Model = m.Model
	m.UpdatedAt = time.Now()
	m.CreatedAt = v.CreatedAt
	_, err = o.Update(m, "CPUCore", "CPUUtilizationID", "GroupID", "MemoryTotalID",
		"MemoryUsedID", "MemoryUtilizationID", "UpdatedAt", "CreatedAt", "UptimeID", "Model")
	if err != nil {
		return err
	}
	return nil
}

//system init
func SystemInit(id int64) error {
	o := orm.NewOrm()
	v := &System{ID: id}
	err := o.Read(v)
	if err != nil {
		return err
	}
	list := strings.Split(v.GroupID, ",")
	err = HostTypeSet(v, list)
	if err != nil {
		return err
	}
	vt := System{ID: id, Status: 1, InitedAt: time.Now()}
	_, err = o.Update(&vt, "status", "inited_at")
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
func HostTypeSet(s *System, groupid []string) error {
	//根据groupid获取host
	OutputPar := []string{"hostid"}
	rep, err := API.CallWithError("host.get", Params{
		"output":   OutputPar,
		"groupids": groupid})
	if err != nil {
		return err
	}
	type hostdata struct {
		HostID string `json:"hostid"`
	}
	var p []hostdata
	resByre, resByteErr := json.Marshal(rep.Result)
	if resByteErr != nil {
		return err
	}
	err = json.Unmarshal(resByre, &p)
	if err != nil {
		return err
	}
	//根据id主机类型
	var htype string
	switch s.ID {
	case 1:
		htype = "VM_LIN"
	case 2:
		htype = "VM_WIN"
	case 3:
		htype = "HW_NET"
	case 4:
		htype = "HW_SRV"
	default:
		htype = "VM_LIN"
	}
	//inventory
	InventoryPara := make(map[string]string)
	InventoryPara["type"] = htype
	_, err = API.CallWithError("host.massupdate", Params{
		"hosts":          p,
		"inventory_mode": 1,
		"inventory":      InventoryPara})
	if err != nil {
		return err
	}
	//CPUCore
	ItemToInventory(s.UptimeID, UptimeID)
	ItemToInventory(s.CPUCore, CPUCore)
	ItemToInventory(s.CPUUtilizationID, CPUUtilizationID)
	ItemToInventory(s.MemoryUtilizationID, MemoryUtilizationID)
	ItemToInventory(s.MemoryTotalID, MemoryTotalID)
	ItemToInventory(s.Model, Model)
	return nil
}

func ItemToInventory(list string, inventoryid int) error {
	itemids := strings.Split(list, ",")
	for _, v := range itemids {
		_, err := API.CallWithError("item.update", Params{
			"itemid":         v,
			"inventory_link": inventoryid})
		if err != nil {
			return err
		}
	}
	return nil
}
