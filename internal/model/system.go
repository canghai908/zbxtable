package models

import (
	"fmt"
	"strings"
	"time"
	"zbxtable/pkg/logger"
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

// GetSystemByID 根据id获取系统列表
func GetSystemByID(id int64) (v *System, err error) {
	v = &System{}
	err = DB.Where("id = ?", id).First(v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetALlSystem 获取所有系统列表
func GetALlSystem() (cnt int64, system []System, err error) {
	var sys []System
	err = DB.Find(&sys).Error
	if err != nil {
		return 0, []System{}, err
	}
	cnt = int64(len(sys))
	return cnt, sys, nil
}

// UpdateSystem 更新系统分类及指标
func UpdateSystem(m *System) (err error) {
	var v System
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err != nil {
		return err
	}
	m.UpdatedAt = time.Now()
	m.CreatedAt = v.CreatedAt
	err = DB.Model(&System{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"cpu_core":              m.CPUCore,
		"cpu_utilization_id":    m.CPUUtilizationID,
		"group_id":              m.GroupID,
		"memory_total_id":       m.MemoryTotalID,
		"memory_used_id":        m.MemoryUsedID,
		"memory_utilization_id": m.MemoryUtilizationID,
		"uptime_id":             m.UptimeID,
		"model":                 m.Model,
		"ping_template_id":      m.PingTemplateID,
		"updated_at":            m.UpdatedAt,
		"created_at":            m.CreatedAt,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// SystemInit 初始化指标
func SystemInit(id int64) error {
	var v System
	err := DB.Where("id = ?", id).First(&v).Error
	if err != nil {
		return err
	}
	list := strings.Split(v.GroupID, ",")
	err = HostTypeSet(&v, list)
	if err != nil {
		return err
	}
	err = DB.Model(&System{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    1,
		"inited_at": time.Now(),
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// HostTypeSet 根据提供的主机组初始化
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

	// 添加检查，如果没有找到主机则返回错误
	if len(p) == 0 {
		return fmt.Errorf("未在指定的主机组中找到任何主机")
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
	//主机类型直接写入，不关联监控指标
	InventoryPara["type"] = hType
	//开启主机Inventory为自动，并归类
	_, err = API.CallWithError("host.massupdate", Params{
		"hosts":          p,
		"inventory_mode": 1,
		"inventory":      InventoryPara})
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	//其他指标绑定
	inventoryItems := []struct {
		ID   string
		Link int
	}{
		{s.UptimeID, UptimeID},
		{s.CPUCore, CPUCore},
		{s.CPUUtilizationID, CPUUtilizationID},
		{s.MemoryUtilizationID, MemoryUtilizationID},
		{s.MemoryTotalID, MemoryTotalID},
		{s.Model, Model},
	}
	// Loop through the inventory items and bind each one
	for _, item := range inventoryItems {
		if item.ID == "" {
			continue
		}
		if err := ItemToInventory(item.ID, item.Link); err != nil {
			logger.Log.Error(err)
			continue
		}
	}
	//ICMP
	if s.PingTemplateID == "" {
		return nil
	}
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
