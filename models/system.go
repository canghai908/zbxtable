package models

import (
	"github.com/astaxie/beego/orm"
	"strings"
	"time"
)

// item link to inventory
// link https://www.zabbix.com/documentation/current/en/manual/api/reference/host/object#host-inventory
const (
	Type                = 1
	CPUCore             = 16
	CPUUtilizationID    = 18
	MemoryUtilizationID = 19
	MemoryTotalID       = 20
	UptimeID            = 22
	Model               = 29
	//Add 20240221
	Ping     = 57
	PingLoss = 58
	PingSec  = 59

	//inventory
	OS            = 5
	MemoryUsedID  = 21
	DateHwInstall = 45
	DateHwExpiry  = 46
	MACAddress    = 12
	SerialNo      = 8
	ResourceID    = 9
	Location      = 24
	Department    = 51
	Vendor        = 31
)

// TableName alarm
func (t *System) TableName() string {
	return TableName("system")
}

// get id
func GetSystemByID(id int64) (v *System, err error) {
	o := orm.NewOrm()
	v = &System{ID: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// get all
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

// get all
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
	v.PingTemplateID = m.PingTemplateID
	m.UpdatedAt = time.Now()
	m.CreatedAt = v.CreatedAt
	_, err = o.Update(m, "CPUCore", "CPUUtilizationID", "GroupID", "MemoryTotalID",
		"MemoryUsedID", "MemoryUtilizationID", "UpdatedAt", "CreatedAt", "UptimeID",
		"Model", "PingTemplateID")
	if err != nil {
		return err
	}
	return nil
}

// SystemInit 初始化指标
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
		return err
	}
	return nil
}

// HostTypeSet 根据提供的主机租初始化
func HostTypeSet(s *System, groupId []string) error {
	//根据groupid获取host
	OutputPar := []string{"hostid"}
	rep, err := API.CallWithError("host.get", Params{
		"output":   OutputPar,
		"groupids": groupId})
	if err != nil {
		return err
	}
	type hostData struct {
		HostID string `json:"hostid"`
	}
	var p []hostData
	resByre, resByteErr := json.Marshal(rep.Result)
	if resByteErr != nil {
		return err
	}
	err = json.Unmarshal(resByre, &p)
	if err != nil {
		return err
	}
	//根据id主机类型
	var hType string
	switch s.ID {
	case 1:
		hType = "VM_LIN"
	case 2:
		hType = "VM_WIN"
	case 3:
		hType = "HW_NET"
	case 4:
		hType = "HW_SRV"
	default:
		hType = "VM_LIN"
	}
	//inventory
	InventoryPara := make(map[string]string)
	//开启主机Inventory为自动，并归类
	_, err = API.CallWithError("host.massupdate", Params{
		"hosts":          p,
		"inventory_mode": 1,
		"inventory":      InventoryPara})
	if err != nil {
		return err
	}
	//主机类型直接写入，不关联监控指标
	InventoryPara["type"] = hType
	//其他指标绑定
	ItemToInventory(s.UptimeID, UptimeID)
	ItemToInventory(s.CPUCore, CPUCore)
	ItemToInventory(s.CPUUtilizationID, CPUUtilizationID)
	ItemToInventory(s.MemoryUtilizationID, MemoryUtilizationID)
	ItemToInventory(s.MemoryTotalID, MemoryTotalID)
	ItemToInventory(s.Model, Model)
	//ICMP
	ICMPToInventory(s.PingTemplateID)
	return nil
}

// ItemToInventory 指标绑定到类型
func ItemToInventory(list string, inventoryId int) error {
	itemIDs := strings.Split(list, ",")
	for _, v := range itemIDs {
		_, err := API.CallWithError("item.update", Params{
			"itemid":         v,
			"inventory_link": inventoryId})
		if err != nil {
			return err
		}
	}
	return nil
}

// ICMPToInventory queries ICMP metrics via the Zabbix API and associates them with the inventory
func ICMPToInventory(id string) error {
	type itemData struct {
		ItemID string `json:"itemid"`
		Key    string `json:"key_"`
	}

	rep, err := API.CallWithError("item.get", Params{
		"output":  []string{"itemid", "key_"},
		"hostids": id,
	})
	if err != nil {
		return err
	}

	resByre, resByteErr := json.Marshal(rep.Result)
	if resByteErr != nil {
		return err
	}
	var items []itemData
	if err := json.Unmarshal(resByre, &items); err != nil {
		return err
	}
	for _, item := range items {
		var inventoryLink int
		switch item.Key {
		case "icmpping":
			inventoryLink = Ping
		case "icmppingloss":
			inventoryLink = PingLoss
		case "icmppingsec":
			inventoryLink = PingSec
		default:
			continue
		}
		if _, err := API.CallWithError("item.update", Params{
			"itemid":         item.ItemID,
			"inventory_link": inventoryLink,
		}); err != nil {
			return err
		}
	}

	return nil
}
