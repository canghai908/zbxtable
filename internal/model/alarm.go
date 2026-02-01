package model

import (
	"strconv"
	"time"
)

// TableName alarm
func (t *Alarm) TableName() string {
	return TableName("alarm")
}

// AddAlarm insert a new Alarm into database and returns
// last inserted Id on success.
func AddAlarm(m *Alarm) (id int64, err error) {
	err = DB.Create(m).Error
	if err != nil {
		return 0, err
	}
	return int64(m.ID), nil
}

// //update alarm notifystatus
func UpdateAlarmStatus(m *Alarm) (id int64, err error) {
	err = DB.Model(&Alarm{}).Where("id = ?", m.ID).Update("notify_status", m.NotifyStatus).Error
	if err != nil {
		return 0, err
	}
	return int64(m.ID), nil
}

// GetAlarmByID retrieves Alarm by Id. Returns error if
// Id doesn't exist
func GetAlarmByID(id int) (v *Alarm, err error) {
	v = &Alarm{}
	err = DB.Where("id = ?", id).First(v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetAllAlarm retrieves all Alarm matches certain condition. Returns empty list if
// no records exist
func GetAllAlarm(begin, end time.Time, page, limit, hosts, ip, tenant_id, status, level string) (cnt int64, al []Alarm, err error) {
	var alarms []Alarm
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	if limits == 0 {
		limits = 10
	}
	if pages == 0 {
		pages = 1
	}
	query := DB.Model(&Alarm{}).Where("occurtime >= ? AND occurtime <= ?", begin, end)
	// 多实例场景：显示所有实例的告警，不再过滤实例
	// 注释掉原来的实例过滤逻辑，让用户可以看到所有实例的告警
	// if inst, _ := GetActiveZabbixInstance(); inst != nil && inst.ZID != 0 {
	// 	query = query.Where("zabbix_instance_id = ?", inst.ZID)
	// }

	if hosts != "" {
		query = query.Where("host LIKE ?", "%"+hosts+"%")
	}
	if tenant_id != "" {
		query = query.Where("instance_id = ?", tenant_id)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if ip != "" {
		query = query.Where("host_ip LIKE ?", "%"+ip+"%")
	}

	// 获取总数
	err = query.Count(&cnt).Error
	if err != nil {
		return 0, []Alarm{}, err
	}

	// 获取分页数据
	offset := (pages - 1) * limits
	err = query.Order("occurtime DESC").Limit(limits).Offset(offset).Find(&alarms).Error
	if err != nil {
		return 0, []Alarm{}, err
	}

	// 填充实例名称
	fillAlarmInstanceNames(&alarms)

	return cnt, alarms, nil
}

// get alarm tenant list
func GetAlarmTenant() (cnt int64, data interface{}, err error) {
	type TenantResult struct {
		InstanceID string `gorm:"column:instance_id"`
	}
	var results []TenantResult
	query := DB.Model(&Alarm{})
	// 多实例场景：显示所有实例的租户
	// 注释掉原来的实例过滤逻辑
	// if inst, _ := GetActiveZabbixInstance(); inst != nil && inst.ZID != 0 {
	// 	query = query.Where("zabbix_instance_id = ?", inst.ZID)
	// }
	err = query.Select("DISTINCT tenant_id").Find(&results).Error
	if err != nil {
		return 0, []Alarm{}, err
	}
	type list struct {
		ID         int    `json:"id"`
		InstanceID string `json:"instance_id"`
	}
	var ss []list
	for i, result := range results {
		ss = append(ss, list{ID: i, InstanceID: result.InstanceID})
	}
	return int64(len(ss)), ss, nil
}

// ExportAlarm export
func ExportAlarm(begin, end time.Time,
	hosts, tenant_id, status, level, hostIP string) ([]byte, error) {
	var alarms []Alarm
	intbegin := begin.Unix()
	intend := end.Unix()

	query := DB.Model(&Alarm{}).Where("occurtime >= ? AND occurtime <= ?", begin, end)

	if hosts != "" {
		query = query.Where("host LIKE ?", "%"+hosts+"%")
	}
	if tenant_id != "" {
		query = query.Where("instance_id = ?", tenant_id)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if hostIP != "" {
		query = query.Where("host_ip LIKE ?", "%"+hostIP+"%")
	}

	err := query.Order("occurtime DESC").Find(&alarms).Error
	if err != nil {
		return []byte{}, err
	}

	// 填充实例名称
	fillAlarmInstanceNames(&alarms)

	cnt := int64(len(alarms))
	pbye, err := CreateAlarmXlsx(alarms, cnt, intbegin, intend)
	if err != nil {
		return []byte{}, err
	}
	return pbye, nil
}

// AnalysisAlarm all alarm
func AnalysisAlarm(begin, end time.Time, tenant_id string) (arrytile []string, pie []Pie, na []string, va []int, err error) {
	strbeing := begin.Format("2006-01-02 15:04:05")
	strend := end.Format("2006-01-02 15:04:05")
	var ss []string
	dpie := []Pie{}

	// 饼图数据查询
	type LevelCount struct {
		Level      string `gorm:"column:level"`
		LevelCount int    `gorm:"column:level_count"`
	}
	var levelCounts []LevelCount

	query := DB.Model(&Alarm{}).
		Select("level, COUNT(DISTINCT id) AS level_count").
		Where("occurtime >= ? AND occurtime <= ?", strbeing, strend).
		Where("(status = ? OR status = ?)", "故障", "1")

	// 多实例场景：显示所有实例的告警分析
	// 注释掉原来的实例过滤逻辑
	// if inst, _ := GetActiveZabbixInstance(); inst != nil && inst.ZID != 0 {
	// 	query = query.Where("zabbix_instance_id = ?", inst.ZID)
	// }

	if tenant_id != "" {
		query = query.Where("instance_id = ?", tenant_id)
	}

	err = query.Group("level").Order("level_count DESC").Find(&levelCounts).Error
	if err == nil && len(levelCounts) > 0 {
		for _, lc := range levelCounts {
			ss = append(ss, lc.Level)
			dpie = append(dpie, Pie{Value: lc.LevelCount, Name: lc.Level})
		}
	}

	// Top10 主机查询（包含实例信息）
	type HostCount struct {
		Hostname         string `gorm:"column:hostname"`
		ZabbixInstanceID int    `gorm:"column:zabbix_instance_id"`
		HostCount        int    `gorm:"column:host_count"`
	}
	var hostCounts []HostCount

	hostQuery := DB.Model(&Alarm{}).
		Select("hostname, zabbix_instance_id, COUNT(DISTINCT id) AS host_count").
		Where("occurtime >= ? AND occurtime <= ?", strbeing, strend).
		Where("(status = ? OR status = ?)", "故障", "1")

	// 多实例场景：显示所有实例的主机统计
	// 注释掉原来的实例过滤逻辑
	// if inst, _ := GetActiveZabbixInstance(); inst != nil && inst.ZID != 0 {
	// 	hostQuery = hostQuery.Where("zabbix_instance_id = ?", inst.ZID)
	// }

	if tenant_id != "" {
		hostQuery = hostQuery.Where("instance_id = ?", tenant_id)
	}

	err = hostQuery.Group("hostname, zabbix_instance_id").Order("host_count DESC").Limit(10).Find(&hostCounts).Error
	var name []string
	var values []int
	if err == nil && len(hostCounts) > 0 {
		// 批量查询实例信息
		instanceMap := make(map[int]string)
		instanceIDs := make(map[int]bool)
		for _, hc := range hostCounts {
			if hc.ZabbixInstanceID > 0 {
				instanceIDs[hc.ZabbixInstanceID] = true
			}
		}
		for id := range instanceIDs {
			tenant, err := GetZabbixInstanceByZID(id)
			if err == nil && tenant != nil {
				instanceMap[id] = tenant.Name
			}
		}

		// 组合主机名和实例名
		for _, hc := range hostCounts {
			hostName := hc.Hostname
			if instanceName, ok := instanceMap[hc.ZabbixInstanceID]; ok && instanceName != "" {
				hostName = hc.Hostname + " [" + instanceName + "]"
			}
			name = append(name, hostName)
			values = append(values, hc.HostCount)
		}
	}

	return ss, dpie, name, values, nil
}

// fillAlarmInstanceNames 填充告警数据的实例信息（InstanceID 和 InstanceName）
func fillAlarmInstanceNames(alarms *[]Alarm) {
	if alarms == nil || len(*alarms) == 0 {
		return
	}

	// 收集所有唯一的 ZID
	instanceIDs := make(map[int]bool)
	for _, alarm := range *alarms {
		if alarm.ZID > 0 {
			instanceIDs[alarm.ZID] = true
		}
	}

	// 批量查询实例信息
	instanceMap := make(map[int]*ZabbixInstance)
	for id := range instanceIDs {
		instance, err := GetZabbixInstanceByZID(id)
		if err == nil && instance != nil {
			instanceMap[id] = instance
		}
	}

	// 填充实例信息（InstanceID 和 InstanceName）
	for i := range *alarms {
		if (*alarms)[i].ZID > 0 {
			if instance, ok := instanceMap[(*alarms)[i].ZID]; ok {
				(*alarms)[i].InstanceName = instance.Name
				(*alarms)[i].InstanceID = instance.InstanceID
			}
		}
	}
}
