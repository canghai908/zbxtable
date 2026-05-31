package model

import (
	"fmt"
	"strings"
	"time"
	"zbxtable/pkg/logger"

	"gorm.io/gorm"
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

// GetSystemByInstance 根据实例ID和系统ID获取配置
func GetSystemByInstance(systemID int64, instanceID int) (v *System, err error) {
	v = &System{}
	err = DB.Where("id = ? AND zid = ?", systemID, instanceID).First(v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetALlSystem 获取所有系统列表
func GetALlSystem() (cnt int64, system []System, err error) {
	if err = EnsureSystemBindingsForAssetTypes(); err != nil {
		return 0, []System{}, err
	}
	var sys []System
	err = DB.Find(&sys).Error
	if err != nil {
		return 0, []System{}, err
	}
	cnt = int64(len(sys))
	return cnt, sys, nil
}

// GetSystemsByInstance 获取指定实例的所有系统配置
func GetSystemsByInstance(instanceID int) ([]System, error) {
	var sys []System
	err := DB.Where("zid = ?", instanceID).Find(&sys).Error
	if err != nil {
		return []System{}, err
	}
	return sys, nil
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
	initCron := m.InitCron
	if initCron == "" {
		initCron = "0 0 2 * * *"
	}
	err = DB.Model(&System{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"zid":                   m.ZID,
		"type_code":             m.TypeCode,
		"cpu_core":              m.CPUCore,
		"cpu_utilization_id":    m.CPUUtilizationID,
		"group_id":              m.GroupID,
		"memory_total_id":       m.MemoryTotalID,
		"memory_used_id":        m.MemoryUsedID,
		"memory_utilization_id": m.MemoryUtilizationID,
		"uptime_id":             m.UptimeID,
		"model":                 m.Model,
		"ping_template_id":      m.PingTemplateID,
		// 自动初始化相关字段（之前缺失导致 UI 开关保存后 DB 不变）
		"auto_init":        m.AutoInit,
		"init_cron":        initCron,
		"init_on_new_host": m.InitOnNewHost,
		"max_retry":        m.MaxRetry,
		"updated_at":       m.UpdatedAt,
		"created_at":       m.CreatedAt,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// EnsureSystemBindingsForAssetTypes 确保每个资产类型都有一条资产绑定配置
func EnsureSystemBindingsForAssetTypes() error {
	assetTypes, err := GetAllAssetTypes()
	if err != nil {
		return err
	}
	if len(assetTypes) == 0 {
		return nil
	}

	now := time.Now()
	return DB.Transaction(func(tx *gorm.DB) error {
		var systems []System
		if err := tx.Find(&systems).Error; err != nil {
			return err
		}

		existingByType := make(map[string]bool, len(systems))
		for _, sys := range systems {
			if sys.TypeCode == "" {
				continue
			}
			existingByType[sys.TypeCode] = true
		}

		var missing []System
		for _, assetType := range assetTypes {
			if existingByType[assetType.TypeCode] {
				continue
			}
			missing = append(missing, System{
				Name:          assetType.Name,
				TypeCode:      assetType.TypeCode,
				Status:        0,
				InitedAt:      &now,
				AutoInit:      1,             // 默认开启定时自动初始化
				InitCron:      "0 0 2 * * *", // 默认每天凌晨 2 点
				InitOnNewHost: 1,             // 默认开启新主机自动初始化
				MaxRetry:      3,
			})
		}
		if len(missing) == 0 {
			return nil
		}
		return tx.Create(&missing).Error
	})
}

// CreateOrUpdateSystem 创建或更新系统配置（支持多实例）
func CreateOrUpdateSystem(m *System) error {
	var existing System
	err := DB.Where("id = ?", m.ID).First(&existing).Error

	if err != nil {
		// 不存在，创建新记录
		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()
		if createErr := DB.Create(m).Error; createErr != nil {
			return createErr
		}
		// 新建后重载调度器，使 auto_init / init_cron 生效
		go ReloadTaskScheduler()
		return nil
	}

	// 存在，更新记录
	m.UpdatedAt = time.Now()
	m.CreatedAt = existing.CreatedAt
	initCron := m.InitCron
	if initCron == "" {
		initCron = "0 0 2 * * *"
	}
	err = DB.Model(&System{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"zid":                   m.ZID,
		"type_code":             m.TypeCode,
		"cpu_core":              m.CPUCore,
		"cpu_utilization_id":    m.CPUUtilizationID,
		"group_id":              m.GroupID,
		"memory_total_id":       m.MemoryTotalID,
		"memory_used_id":        m.MemoryUsedID,
		"memory_utilization_id": m.MemoryUtilizationID,
		"uptime_id":             m.UptimeID,
		"model":                 m.Model,
		"ping_template_id":      m.PingTemplateID,
		// 自动化配置字段（之前缺失导致编辑保存后不生效）
		"auto_init":        m.AutoInit,
		"init_cron":        initCron,
		"init_on_new_host": m.InitOnNewHost,
		"max_retry":        m.MaxRetry,
		"updated_at":       m.UpdatedAt,
	}).Error
	if err != nil {
		return err
	}
	// 配置变更后重启调度器，使新的 auto_init / init_cron 立即生效
	go ReloadTaskScheduler()
	return nil
}

// CreateSystem 创建新的 System 绑定记录
func CreateSystem(m *System) error {
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	return DB.Create(m).Error
}

// DeleteSystem 删除 System 绑定记录（同时级联删除其历史记录）
func DeleteSystem(id int64) error {
	if err := DB.Where("system_id = ?", id).Delete(&SystemHistory{}).Error; err != nil {
		logger.Log.Errorf("删除绑定历史记录失败 [ID=%d]: %v", id, err)
	}
	return DB.Delete(&System{}, id).Error
}

// SystemInit 初始化指标
// ExecuteSystemInit 执行资产绑定初始化，记录历史，支持 execType: manual/auto/retry
func ExecuteSystemInit(systemID int64, execType string) error {
	var v System
	if err := DB.Where("id = ?", systemID).First(&v).Error; err != nil {
		return err
	}

	hist := &SystemHistory{
		SystemID:  systemID,
		ZID:       v.ZID,
		TypeCode:  v.TypeCode,
		ExecType:  execType,
		StartTime: time.Now(),
		Status:    "running",
	}
	DB.Create(hist)

	apiInstance, err := GetAPIByZID(v.ZID)
	if err != nil {
		return systemInitError(&v, hist, fmt.Errorf("获取实例API失败: %w", err))
	}

	list := strings.Split(v.GroupID, ",")
	affectedHosts, err := hostTypeSetWithCount(&v, list, apiInstance)
	if err != nil {
		retErr := systemInitError(&v, hist, err)
		go CleanupSystemHistory(systemID, systemHistoryKeepPerBinding)
		return retErr
	}

	retErr := systemInitSuccess(&v, hist, affectedHosts)
	// 异步清理该绑定的超额历史记录，避免历史表无限增长
	go CleanupSystemHistory(systemID, systemHistoryKeepPerBinding)
	return retErr
}

// systemHistoryKeepPerBinding 每条绑定保留的历史记录条数
const systemHistoryKeepPerBinding = 50

// CleanupSystemHistory 仅保留指定绑定最近 keep 条历史记录，删除更早的
func CleanupSystemHistory(systemID int64, keep int) {
	if keep <= 0 {
		return
	}
	// 找出排在第 keep 条之后的旧记录 ID（按开始时间倒序）
	var oldIDs []int64
	if err := DB.Model(&SystemHistory{}).
		Where("system_id = ?", systemID).
		Order("start_time DESC").
		Offset(keep).
		Limit(100000).
		Pluck("id", &oldIDs).Error; err != nil {
		logger.Log.Errorf("查询待清理历史记录失败 [ID=%d]: %v", systemID, err)
		return
	}
	if len(oldIDs) == 0 {
		return
	}
	if err := DB.Where("id IN ?", oldIDs).Delete(&SystemHistory{}).Error; err != nil {
		logger.Log.Errorf("清理历史记录失败 [ID=%d]: %v", systemID, err)
		return
	}
	logger.Log.Infof("已清理绑定 [ID=%d] 的 %d 条旧历史记录（保留最近 %d 条）", systemID, len(oldIDs), keep)
}

func systemInitError(v *System, hist *SystemHistory, err error) error {
	now := time.Now()
	DB.Model(hist).Updates(map[string]interface{}{
		"end_time":      &now,
		"duration":      int(now.Sub(hist.StartTime).Seconds()),
		"status":        "failed",
		"error_message": err.Error(),
	})
	DB.Model(&System{}).Where("id = ?", v.ID).Updates(map[string]interface{}{
		"status":      2,
		"init_error":  err.Error(),
		"retry_count": v.RetryCount + 1,
		"inited_at":   &now,
	})
	logger.Log.Errorf("资产绑定初始化失败 [ID=%d, type=%s]: %v", v.ID, v.TypeCode, err)
	return err
}

func systemInitSuccess(v *System, hist *SystemHistory, affectedHosts int) error {
	now := time.Now()
	DB.Model(hist).Updates(map[string]interface{}{
		"end_time":       &now,
		"duration":       int(now.Sub(hist.StartTime).Seconds()),
		"status":         "success",
		"affected_hosts": affectedHosts,
	})
	DB.Model(&System{}).Where("id = ?", v.ID).Updates(map[string]interface{}{
		"status":          1,
		"inited_at":       &now,
		"last_success_at": &now,
		"init_error":      "",
		"retry_count":     0,
	})
	logger.Log.Infof("资产绑定初始化成功 [ID=%d, type=%s, 影响主机=%d]", v.ID, v.TypeCode, affectedHosts)
	return nil
}

// GetAutoInitSystems 获取启用自动初始化的绑定配置
func GetAutoInitSystems() ([]System, error) {
	var list []System
	err := DB.Where("auto_init = ?", 1).Find(&list).Error
	return list, err
}

// GetFailedSystemsForRetry 获取需要重试的失败配置
func GetFailedSystemsForRetry() ([]System, error) {
	var list []System
	err := DB.Where("status = ? AND retry_count < max_retry", 2).Find(&list).Error
	return list, err
}

// GetSystemHistory 获取指定绑定配置的执行历史（分页）
func GetSystemHistory(systemID int64, page, pageSize int) ([]SystemHistory, int64, error) {
	var list []SystemHistory
	var total int64
	query := DB.Model(&SystemHistory{}).Where("system_id = ?", systemID)
	query.Count(&total)
	err := query.Order("start_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

// SystemInit 保留兼容旧调用
func SystemInit(id int64) error {
	return ExecuteSystemInit(id, "manual")
}

// SystemInitWithInstance 保留兼容旧调用
func SystemInitWithInstance(systemID int64, _ int) error {
	return ExecuteSystemInit(systemID, "manual")
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

	// 从 type_code 字段读取资产类型（动态，不再硬编码）
	hType := s.TypeCode
	if hType == "" {
		return fmt.Errorf("系统配置 [id=%d] 未设置资产类型，请先在设备绑定配置中设置 type_code", s.ID)
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

// HostTypeSetWithInstance 根据提供的主机组初始化（支持多实例）
// hostTypeSetWithCount 执行初始化并返回影响主机数
func hostTypeSetWithCount(s *System, groupId []string, apiInstance *APIInstance) (int, error) {
	err := HostTypeSetWithInstance(s, groupId, apiInstance)
	if err != nil {
		return 0, err
	}
	// 查询主机组实际主机数
	OutputPar := []string{"hostid"}
	rep, err2 := apiInstance.API.CallWithError("host.get", Params{
		"output":   OutputPar,
		"groupids": groupId,
	})
	if err2 != nil {
		return 0, nil
	}
	type hostData struct {
		HostID string `json:"hostid"`
	}
	var hosts []hostData
	if b, e := json.Marshal(rep.Result); e == nil {
		_ = json.Unmarshal(b, &hosts)
	}
	return len(hosts), nil
}

func HostTypeSetWithInstance(s *System, groupId []string, apiInstance *APIInstance) error {
	//根据groupid获取host
	OutputPar := []string{"hostid"}
	rep, err := apiInstance.API.CallWithError("host.get", Params{
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

	// 从 type_code 字段读取资产类型（动态，不再硬编码）
	hType := s.TypeCode
	if hType == "" {
		return fmt.Errorf("系统配置 [id=%d] 未设置资产类型，请先在设备绑定配置中设置 type_code", s.ID)
	}
	//inventory
	InventoryPara := make(map[string]string)
	//主机类型直接写入，不关联监控指标
	InventoryPara["type"] = hType

	// 分批调用 host.massupdate（防止主机数过多时单次请求超时/内存溢出）
	const massUpdateBatchSize = 500
	for i := 0; i < len(p); i += massUpdateBatchSize {
		end := i + massUpdateBatchSize
		if end > len(p) {
			end = len(p)
		}
		batch := p[i:end]
		if _, batchErr := apiInstance.API.CallWithError("host.massupdate", Params{
			"hosts":          batch,
			"inventory_mode": 1,
			"inventory":      InventoryPara,
		}); batchErr != nil {
			logger.Log.Errorf("host.massupdate 批次 [%d-%d] 失败: %v", i, end, batchErr)
			return batchErr
		}
		logger.Log.Infof("host.massupdate [type=%s] 批次 [%d/%d] 完成", hType, end, len(p))
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
		if err := ItemToInventoryWithInstance(item.ID, item.Link, apiInstance); err != nil {
			logger.Log.Error(err)
			continue
		}
	}
	//ICMP
	if s.PingTemplateID == "" {
		return nil
	}
	ICMPToInventoryWithInstance(s.PingTemplateID, apiInstance)
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

// ItemToInventoryWithInstance 指标绑定到类型（支持多实例）
func ItemToInventoryWithInstance(list string, inventoryId int, apiInstance *APIInstance) error {
	itemIDs := strings.Split(list, ",")
	for _, v := range itemIDs {
		_, err := apiInstance.API.CallWithError("item.update", Params{
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

// ICMPToInventoryWithInstance queries ICMP metrics via the Zabbix API and associates them with the inventory（支持多实例）
func ICMPToInventoryWithInstance(id string, apiInstance *APIInstance) error {
	type itemData struct {
		ItemID string `json:"itemid"`
		Key    string `json:"key_"`
	}

	rep, err := apiInstance.API.CallWithError("item.get", Params{
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
		if _, err := apiInstance.API.CallWithError("item.update", Params{
			"itemid":         item.ItemID,
			"inventory_link": inventoryLink,
		}); err != nil {
			return err
		}
	}

	return nil
}
