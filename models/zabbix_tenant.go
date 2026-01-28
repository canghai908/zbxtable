package models

import (
	"errors"
	"strings"
)

func ListZabbixTenantBindings() ([]ZabbixTenantBinding, error) {
	var list []ZabbixTenantBinding
	err := DB.Order("id desc").Find(&list).Error
	return list, err
}

func GetZabbixTenantBindingByTenant(tenantID string) (*ZabbixTenantBinding, error) {
	tid := strings.TrimSpace(tenantID)
	if tid == "" {
		return nil, errors.New("tenant_id is empty")
	}
	var v ZabbixTenantBinding
	err := DB.Where("tenant_id = ?", tid).First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func UpsertZabbixTenantBinding(m *ZabbixTenantBinding) error {
	if m == nil {
		return errors.New("binding is nil")
	}
	m.TenantID = strings.TrimSpace(m.TenantID)
	m.Token = strings.TrimSpace(m.Token)
	if m.TenantID == "" {
		return errors.New("tenant_id is empty")
	}
	if m.ZabbixInstanceID == 0 {
		return errors.New("zabbix_instance_id is empty")
	}
	// upsert by tenant_id (unique)
	var exist ZabbixTenantBinding
	err := DB.Where("tenant_id = ?", m.TenantID).First(&exist).Error
	if err == nil {
		return DB.Model(&ZabbixTenantBinding{}).
			Where("id = ?", exist.ID).
			Updates(map[string]interface{}{
				"zabbix_instance_id": m.ZabbixInstanceID,
				"token":              m.Token,
				"enabled":            m.Enabled,
			}).Error
	}
	return DB.Create(m).Error
}

func DeleteZabbixTenantBinding(id int) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return DB.Delete(&ZabbixTenantBinding{}, id).Error
}


