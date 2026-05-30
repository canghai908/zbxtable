package model

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"zbxtable/pkg/logger"

	"github.com/robfig/cron/v3"
)

var (
	cronScheduler *cron.Cron
	cronMu        sync.Mutex
)

type scheduledTaskDefinition struct {
	Name           string
	EnabledKey     string
	CronKey        string
	DefaultEnabled string
	DefaultCron    string
	Run            func() error
}

var taskCronParser = cron.NewParser(
	cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

func getScheduledTaskDefinitions() []scheduledTaskDefinition {
	return []scheduledTaskDefinition{
		{
			Name:           "首页 Top 数据同步",
			EnabledKey:     "top_sync_enabled",
			CronKey:        "top_sync_cron",
			DefaultEnabled: "1",
			DefaultCron:    "0/30 * * * * *",
			Run:            TOP,
		},
		{
			Name:           "日报生成",
			EnabledKey:     "day_report_enabled",
			CronKey:        "day_report_cron",
			DefaultEnabled: "1",
			DefaultCron:    "0 55 23 * * *",
			Run:            CreateDayReport,
		},
		{
			Name:           "周报生成",
			EnabledKey:     "week_report_enabled",
			CronKey:        "week_report_cron",
			DefaultEnabled: "1",
			DefaultCron:    "0 55 17 * * 5",
			Run:            CreateWeekReport,
		},
		{
			Name:           "主机分类同步",
			EnabledKey:     "sync_inventory",
			CronKey:        "sync_inventory_cron",
			DefaultEnabled: "1",
			DefaultCron:    "0 */5 * * * *",
			Run:            SyncInventory,
		},
		{
			Name:           "自动指标映射",
			EnabledKey:     "auto_metric_mapping_enabled",
			CronKey:        "auto_metric_mapping_cron",
			DefaultEnabled: "1",
			DefaultCron:    "0 0 * * * *",
			Run:            AutoMetricMapping,
		},
		{
			Name:           "失败指标映射重试",
			EnabledKey:     "retry_failed_mapping_enabled",
			CronKey:        "retry_failed_mapping_cron",
			DefaultEnabled: "1",
			DefaultCron:    "0 */30 * * * *",
			Run:            RetryFailedMappings,
		},
		{
			Name:           "出口数据采集",
			EnabledKey:     "egress_collect_enabled",
			CronKey:        "egress_collect_cron",
			DefaultEnabled: "1",
			DefaultCron:    "0 * * * * *",
			Run:            CollectEgressData,
		},
		{
			Name:           "状态总览同步",
			EnabledKey:     "overview_sync_enabled",
			CronKey:        "overview_sync_cron",
			DefaultEnabled: "1",
			DefaultCron:    "0 */5 * * * *",
			Run:            SyncOverviewData,
		},
		{
			Name:           "主机统计缓存刷新",
			EnabledKey:     "host_count_sync_enabled",
			CronKey:        "host_count_sync_cron",
			DefaultEnabled: "1",
			DefaultCron:    "0 */5 * * * *",
			Run:            SyncHostCountCache,
		},
		{
			Name:           "资产绑定自动初始化",
			EnabledKey:     "binding_auto_init_enabled",
			CronKey:        "binding_auto_init_cron",
			DefaultEnabled: "1",
			DefaultCron:    "0 0 2 * * *",
			Run:            AutoInitSystemBindings,
		},
		{
			Name:           "资产绑定失败重试",
			EnabledKey:     "binding_retry_enabled",
			CronKey:        "binding_retry_cron",
			DefaultEnabled: "1",
			DefaultCron:    "0 */30 * * * *",
			Run:            RetryFailedSystemBindings,
		},
	}
}

func IsTaskConfigKey(key string) bool {
	for _, task := range getScheduledTaskDefinitions() {
		if key == task.EnabledKey || key == task.CronKey {
			return true
		}
	}
	return false
}

func IsTaskCronConfigKey(key string) bool {
	for _, task := range getScheduledTaskDefinitions() {
		if key == task.CronKey {
			return true
		}
	}
	return false
}

func IsTaskEnabledConfigKey(key string) bool {
	for _, task := range getScheduledTaskDefinitions() {
		if key == task.EnabledKey {
			return true
		}
	}
	return false
}

func ValidateTaskConfigValue(key, value string) error {
	if IsTaskEnabledConfigKey(key) {
		if value != "0" && value != "1" {
			return fmt.Errorf("计划任务开关只支持 0 或 1")
		}
		return nil
	}
	if IsTaskCronConfigKey(key) {
		if _, err := taskCronParser.Parse(value); err != nil {
			return err
		}
	}
	return nil
}

// InitTask 初始化定时任务
func InitTask() {
	cronMu.Lock()
	defer cronMu.Unlock()
	startTaskSchedulerLocked()
}

// StopTask 停止定时任务
func StopTask() {
	cronMu.Lock()
	defer cronMu.Unlock()
	stopTaskSchedulerLocked()
}

func ReloadTaskScheduler() {
	cronMu.Lock()
	defer cronMu.Unlock()
	stopTaskSchedulerLocked()
	startTaskSchedulerLocked()
}

func startTaskSchedulerLocked() {
	if cronScheduler != nil {
		cronScheduler.Stop()
	}

	cronScheduler = cron.New(cron.WithSeconds())

	for _, task := range getScheduledTaskDefinitions() {
		if GetConfigValueByKey(task.EnabledKey, task.DefaultEnabled) != "1" {
			logger.Log.Infof("计划任务[%s]已禁用，跳过注册", task.Name)
			continue
		}

		spec := GetConfigValueByKey(task.CronKey, task.DefaultCron)
		if _, err := taskCronParser.Parse(spec); err != nil {
			logger.Log.Errorf("计划任务[%s] Cron表达式无效[%s]: %v", task.Name, spec, err)
			continue
		}

		currentTask := task
		_, err := cronScheduler.AddFunc(spec, func() {
			if err := currentTask.Run(); err != nil {
				logger.Log.Errorf("计划任务[%s]执行失败: %v", currentTask.Name, err)
			}
		})
		if err != nil {
			logger.Log.Errorf("注册计划任务[%s]失败: %v", task.Name, err)
			continue
		}
		logger.Log.Infof("计划任务[%s]已注册，Cron=%s", task.Name, spec)
	}

	cronScheduler.Start()
	logger.Log.Info("Cron scheduler started")

	// 为每条 auto_init=1 的资产绑定注册独立的定时任务（已持锁，直接操作 scheduler）
	systems, sysErr := GetAutoInitSystems()
	if sysErr == nil {
		for _, s := range systems {
			spec := s.InitCron
			if spec == "" {
				spec = "0 0 2 * * *"
			}
			if _, parseErr := taskCronParser.Parse(spec); parseErr != nil {
				logger.Log.Errorf("资产绑定 [ID=%d] cron 表达式无效 [%s]: %v", s.ID, spec, parseErr)
				continue
			}
			id := s.ID
			typeCode := s.TypeCode
			if _, addErr := cronScheduler.AddFunc(spec, func() {
				logger.Log.Infof("执行资产绑定定时初始化 [ID=%d, type=%s]", id, typeCode)
				if err := ExecuteSystemInit(id, "auto"); err != nil {
					logger.Log.Errorf("资产绑定定时初始化失败 [ID=%d]: %v", id, err)
				}
			}); addErr != nil {
				logger.Log.Errorf("注册资产绑定 [ID=%d] cron 失败: %v", s.ID, addErr)
				continue
			}
			logger.Log.Infof("资产绑定 [ID=%d, type=%s] 定时任务已注册，Cron=%s", s.ID, s.TypeCode, spec)
		}
	}
}

func stopTaskSchedulerLocked() {
	if cronScheduler == nil {
		return
	}

	cronScheduler.Stop()
	cronScheduler = nil
	logger.Log.Info("Cron scheduler stopped")
}
func CreateWeekReport() error {
	_, list, err := GetALlReport()
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	//遍历周报
	for _, v := range list {
		//启用周报，只处理循环报表（scheduled模式），跳过实时报表（realtime模式）
		if v.Status == "1" && len(v.Cycle) != 0 && (v.ReportMode == "" || v.ReportMode == "scheduled") {
			// 检查报表类型和配置
			hasConfig := false
			if v.ReportType == "host" {
				hasConfig = len(v.HostIds) > 0 && len(v.ItemIds) > 0
			} else {
				hasConfig = len(v.Items) > 0
			}
			if !hasConfig {
				continue
			}
			cycList := strings.Split(v.Cycle, ",")
			for _, vv := range cycList {
				if vv == "week" {
					start := time.Now()
					//周报生成
					var err error
					if v.ReportType == "host" {
						err = TaskHostReport(v)
					} else {
						err = TaskWeekReport(v)
					}
					if err != nil {
						logger.Log.Error(err)
						///update status failed
						v.ExecStatus = strconv.Itoa(Failed)
						v.StartAt = &start
						now := time.Now()
						v.EndAt = &now
						err = UpdateReportExecStatusByID(&v)
						if err != nil {
							logger.Log.Error(err)
						}
						continue
					}
					//update status success
					v.ExecStatus = strconv.Itoa(Success)
					v.StartAt = &start
					now := time.Now()
					v.EndAt = &now
					err = UpdateReportExecStatusByID(&v)
					if err != nil {
						logger.Log.Error(err)
						continue
					}
				}
			}
		}
	}
	return nil
}

// AutoMetricMapping 自动执行指标映射
func AutoMetricMapping() error {
	mappings, err := GetAutoInitMappings()
	if err != nil {
		logger.Log.Errorf("获取自动初始化配置失败: %v", err)
		return err
	}

	if len(mappings) == 0 {
		logger.Log.Debug("没有启用自动初始化的映射配置")
		return nil
	}

	logger.Log.Infof("开始自动指标映射任务，共 %d 个配置", len(mappings))

	for _, mapping := range mappings {
		// 检查是否需要执行（距离上次成功执行超过24小时）
		if shouldExecuteMapping(&mapping) {
			go func(m MetricMapping) {
				logger.Log.Infof("自动执行指标映射 [ID=%d, Instance=%d, Type=%s]", m.ID, m.ZID, m.SystemType)
				err := ExecuteMetricMapping(&m, "auto")
				if err != nil {
					logger.Log.Errorf("自动执行指标映射失败 [ID=%d]: %v", m.ID, err)
				}
			}(mapping)
		}
	}

	return nil
}

// RetryFailedMappings 重试失败的映射
func RetryFailedMappings() error {
	mappings, err := GetFailedMappingsForRetry()
	if err != nil {
		logger.Log.Errorf("获取失败映射配置失败: %v", err)
		return err
	}

	if len(mappings) == 0 {
		return nil
	}

	logger.Log.Infof("开始重试失败的指标映射，共 %d 个配置", len(mappings))

	for _, mapping := range mappings {
		// 检查是否需要重试（距离上次失败超过1小时）
		if mapping.LastInitAt != nil && time.Since(*mapping.LastInitAt) > time.Hour {
			go func(m MetricMapping) {
				logger.Log.Infof("重试指标映射 [ID=%d, Retry=%d/%d]", m.ID, m.RetryCount+1, m.MaxRetry)
				err := ExecuteMetricMapping(&m, "retry")
				if err != nil {
					logger.Log.Errorf("重试指标映射失败 [ID=%d]: %v", m.ID, err)
				}
			}(mapping)
		}
	}

	return nil
}

// shouldExecuteMapping 判断是否需要执行映射
func shouldExecuteMapping(mapping *MetricMapping) bool {
	// 如果从未执行过，立即执行
	if mapping.LastSuccessAt == nil {
		return true
	}

	// 如果距离上次成功执行超过24小时，执行
	return time.Since(*mapping.LastSuccessAt) > 24*time.Hour
}
func CreateDayReport() error {
	_, list, err := GetALlReport()
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	for _, v := range list {
		// 只处理循环报表（scheduled模式），跳过实时报表（realtime模式）
		if v.Status == "1" && len(v.Cycle) != 0 && (v.ReportMode == "" || v.ReportMode == "scheduled") {
			// 检查报表类型和配置
			hasConfig := false
			if v.ReportType == "host" {
				hasConfig = len(v.HostIds) > 0 && len(v.ItemIds) > 0
			} else {
				hasConfig = len(v.Items) > 0
			}
			if !hasConfig {
				continue
			}
			cycList := strings.Split(v.Cycle, ",")
			for _, vv := range cycList {
				if vv == "day" {
					start := time.Now()
					var err error
					if v.ReportType == "host" {
						err = TaskHostReport(v)
					} else {
						err = TaskDayReport(v)
					}
					if err != nil {
						logger.Log.Error(err)
						//更新report状态
						v.ExecStatus = strconv.Itoa(Failed)
						v.StartAt = &start
						now := time.Now()
						v.EndAt = &now
						err = UpdateReportExecStatusByID(&v)
						if err != nil {
							logger.Log.Error(err)
						}
						continue
					}
					//更新report状态
					v.ExecStatus = strconv.Itoa(Success)
					v.StartAt = &start
					now := time.Now()
					v.EndAt = &now
					err = UpdateReportExecStatusByID(&v)
					if err != nil {
						logger.Log.Error(err)
						continue
					}
				}
			}
		}
	}
	return nil
}

// linux windows top data to redis (多实例聚合版本)
func TOP() error {
	// 获取所有启用的实例
	instances, err := GetAllEnabledAPIInstances()
	if err != nil {
		logger.Log.Errorf("获取启用的实例失败: %v", err)
		return err
	}

	// 从配置读取 Top 数量
	linTopNumStr := GetConfigValueByKey("dash_top_lin_num", "10")
	linTopN, err := strconv.ParseInt(linTopNumStr, 10, 64)
	if err != nil || linTopN <= 0 {
		linTopN = 10
	}

	winTopNumStr := GetConfigValueByKey("dash_top_win_num", "10")
	winTopN, err := strconv.ParseInt(winTopNumStr, 10, 64)
	if err != nil || winTopN <= 0 {
		winTopN = 10
	}

	// 清空旧数据
	_ = CacheDelete("WIN_CPU")
	_ = CacheDelete("WIN_MEM")
	_ = CacheDelete("LIN_CPU")
	_ = CacheDelete("LIN_MEM")

	// 从所有实例收集数据
	for _, inst := range instances {
		err := TOPFromInstance(inst, linTopN, winTopN)
		if err != nil {
			logger.Log.Errorf("从实例 %s 收集 TOP 数据失败: %v", inst.Name, err)
			continue
		}
	}

	return nil
}

// TOPFromInstance 从指定实例收集 TOP 数据
func TOPFromInstance(inst *APIInstance, linTopN, winTopN int64) error {
	OutputPar := []string{"hostid", "host", "available", "status", "name", "error"}
	SelectInterfacesPar := []string{"ip", "port"}
	SearchInventoryKey := []string{"VM_WIN", "VM_LIN"}
	SearchInventoryPar := make(map[string][]string)
	SearchInventoryPar["type"] = SearchInventoryKey
	rep, err := inst.API.CallWithError("host.get", Params{
		"output":           OutputPar,
		"searchByAny":      true,
		"searchInventory":  SearchInventoryPar,
		"selectInventory":  "extend",
		"selectInterfaces": SelectInterfacesPar})
	if err != nil {
		logger.Log.Debug(err)
		return err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logger.Log.Debug(err)
		return err
	}
	var hb ListHosts
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logger.Log.Debug(err)
		return err
	}
	if len(hb) == 0 {
		return nil // 该实例没有主机，不是错误
	}

	// 定义临时存储，用于每个分类的排序和截断
	type hostScore struct {
		key   string
		score float64
	}
	var winCPUs, winMEMs, linCPUs, linMEMs []hostScore

	for _, v := range hb {
		if v.Available == "0" {
			continue
		}
		hostKey := inst.Instance + "_" + v.Host
		switch v.Inventory.Type {
		case "VM_WIN":
			var cpu, mem float64
			if v.Inventory.SoftwareAppA != "" {
				cpu, _ = strconv.ParseFloat(strings.Replace(v.Inventory.SoftwareAppA, " %", "", -1), 64)
			}
			if v.Inventory.SoftwareAppB != "" {
				mem, _ = strconv.ParseFloat(strings.Replace(v.Inventory.SoftwareAppB, " %", "", -1), 64)
			}
			winCPUs = append(winCPUs, hostScore{hostKey, cpu})
			winMEMs = append(winMEMs, hostScore{hostKey, mem})
		case "VM_LIN":
			var cpu, mem float64
			if v.Inventory.SoftwareAppA != "" {
				cpu, _ = strconv.ParseFloat(strings.Replace(v.Inventory.SoftwareAppA, " %", "", -1), 64)
			}
			if v.Inventory.SoftwareAppB != "" {
				mem, _ = strconv.ParseFloat(strings.Replace(v.Inventory.SoftwareAppB, " %", "", -1), 64)
			}
			linCPUs = append(linCPUs, hostScore{hostKey, cpu})
			linMEMs = append(linMEMs, hostScore{hostKey, mem})
		}
	}

	// 将 score 列表转为 map，便于按 key 写入
	toMap := func(list []hostScore) map[string]float64 {
		m := make(map[string]float64, len(list))
		for _, it := range list {
			m[it.key] = it.score
		}
		return m
	}

	// 取 TopN 的 key 集合（按 score 降序）
	topKeys := func(list []hostScore, n int64) []string {
		// 简单排序（数据量通常不大）
		for i := 0; i < len(list); i++ {
			for j := i + 1; j < len(list); j++ {
				if list[i].score < list[j].score {
					list[i], list[j] = list[j], list[i]
				}
			}
		}
		limit := int(n)
		if len(list) < limit {
			limit = len(list)
		}
		keys := make([]string, 0, limit)
		for i := 0; i < limit; i++ {
			keys = append(keys, list[i].key)
		}
		return keys
	}

	// CPU TopN ∪ MEM TopN，最多 2N
	union := func(a, b []string) []string {
		seen := make(map[string]struct{}, len(a)+len(b))
		out := make([]string, 0, len(a)+len(b))
		for _, k := range a {
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			out = append(out, k)
		}
		for _, k := range b {
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			out = append(out, k)
		}
		return out
	}

	// 写入：对并集内 key，把 CPU/MEM 都写入各自 ZSet，确保前端合并后可切换排序
	writeUnion := func(cpuKey, memKey string, cpuList, memList []hostScore, topN int64) error {
		cpuMap := toMap(cpuList)
		memMap := toMap(memList)
		cpuTop := topKeys(append([]hostScore(nil), cpuList...), topN)
		memTop := topKeys(append([]hostScore(nil), memList...), topN)
		keys := union(cpuTop, memTop)

		for _, k := range keys {
			if err := CacheZAdd(cpuKey, k, cpuMap[k]); err != nil {
				return err
			}
			if err := CacheZAdd(memKey, k, memMap[k]); err != nil {
				return err
			}
		}
		return nil
	}

	if err := writeUnion("WIN_CPU", "WIN_MEM", winCPUs, winMEMs, winTopN); err != nil {
		return err
	}
	if err := writeUnion("LIN_CPU", "LIN_MEM", linCPUs, linMEMs, linTopN); err != nil {
		return err
	}
	return nil
}

// update topology data
func UpdateEdgeDataById(id int) error {
	//get topodata
	p, err := GetTopologyById(id)
	if err != nil {
		logger.Log.Debug(err)
		return err
	}
	var allEdges AllEdge
	err = json.Unmarshal([]byte(p.Edges), &allEdges)
	if err != nil {
		logger.Log.Debug(err)
		return err
	}
	var wg sync.WaitGroup
	ch := make(chan struct{}, 10)
	var aedge []AEdge
	for _, v := range allEdges {
		ch <- struct{}{}
		wg.Add(1)
		go func(v AEdge) {
			defer wg.Done()
			//labels attr
			v.Labels[0].Attrs.Label.Text = ""
			v.Labels[0].Position.Angle = 0
			v.Labels[0].Position.Offset = 20
			v.Labels[0].Position.Options.EnsureLegibility = true
			v.Labels[0].Position.Options.KeepGradient = true
			//line attrs
			v.Attrs.Line.StrokeWidth = 4
			v.Attrs.Line.Stroke = "#A4A4A4"
			v.Attrs.Line.StrokeDasharray = 0
			if v.Attrs.Line.FlowID != "" {
				// 使用边上的ZID从对应实例获取流量数据
				flow, err := GetFlowByFlowIDFromInstance(v.Attrs.Line.ZID, v.Attrs.Line.FlowID)
				if err != nil {
					logger.Log.Error(err)
				}
				v.Labels[0].Attrs.Label.Text = flow
			}
			//trigger get
			if v.Attrs.Line.TriggerID != "" {
				// 使用边上的ZID从对应实例获取触发器状态
				status, err := GetTriggerValueByTriggerIDFromInstance(v.Attrs.Line.ZID, v.Attrs.Line.TriggerID)
				if err != nil {
					logger.Log.Error(err)
				}
				switch {
				//trigger正常 未告警
				case status == "0":
					v.Attrs.Line.Stroke = "#00FF00"
					v.Attrs.Line.StrokeDasharray = 5
					v.Attrs.Line.Style.Animation = "ant-line 30s infinite linear"
					//trigger 告警
				case status == "1":
					v.Attrs.Line.Stroke = "#FF0000"
				case status == "2":
					v.Attrs.Line.Stroke = "#A4A4A4"
				default:
					v.Attrs.Line.Stroke = "#A4A4A4"
				}
			}
			aedge = append(aedge, v)
			<-ch
		}(v)
		wg.Wait()
	}
	edgeStr, err := json.Marshal(aedge)
	if err != nil {
		logger.Log.Debug(err)
		return err
	}
	var Topo Topology
	Topo.ID = id
	Topo.Edges = string(edgeStr)
	err = UpdateTopologyEdgesByID(&Topo)
	if err != nil {
		logger.Log.Debug(err)
		return err
	}
	return nil
}

// SyncInventory 同步主机分类及数据绑定（支持多实例）
func SyncInventory() error {
	if GetConfigValueByKey("sync_inventory", "1") != "1" {
		return nil
	}

	// 获取所有启用的实例
	instances, err := GetAllEnabledAPIInstances()
	if err != nil {
		logger.Log.Errorf("获取启用的实例失败: %v", err)
		return err
	}

	// 遍历每个实例，执行同步
	for _, inst := range instances {
		// 查询该实例的系统配置
		var list []System
		err = GetDB().Where("zid = ? AND status = ?", inst.ZID, 1).Find(&list).Error
		if err != nil {
			logger.Log.Errorf("查询实例 %s 的系统配置失败: %v", inst.Name, err)
			continue
		}
		if len(list) == 0 {
			logger.Log.Infof("实例 %s 没有已初始化的系统配置，跳过同步", inst.Name)
			continue
		}

		// 对该实例的每个系统配置执行同步
		for _, v := range list {
			gList := strings.Split(v.GroupID, ",")
			err = HostTypeSetWithInstance(&v, gList, inst)
			if err != nil {
				logger.Log.Errorf("实例 %s 同步系统配置 %d 失败: %v", inst.Name, v.ID, err)
				continue
			}
			logger.Log.Infof("实例 %s 同步系统配置 %d 成功", inst.Name, v.ID)
		}
	}

	return nil
}

// SyncOverviewData 同步状态纵览数据到缓存（支持多实例）
func SyncOverviewData() error {
	logger.Log.Info("开始同步状态纵览数据到缓存")

	// 获取所有启用的实例
	instances, err := GetAllEnabledAPIInstances()
	if err != nil {
		logger.Log.Errorf("获取启用的实例失败: %v", err)
		return err
	}

	// 从数据库动态获取主机类型列表
	assetTypes, err := GetAllAssetTypes()
	if err != nil {
		logger.Log.Errorf("获取资产类型列表失败: %v", err)
		return err
	}
	hostTypes := make([]string, 0, len(assetTypes))
	for _, at := range assetTypes {
		hostTypes = append(hostTypes, at.TypeCode)
	}

	// 遍历每种主机类型
	for _, hostType := range hostTypes {
		var allHosts []Hosts

		// 从所有实例收集该类型的主机数据
		for _, inst := range instances {
			hosts, err := getOverviewHostsFromInstance(inst, hostType)
			if err != nil {
				logger.Log.Errorf("从实例 %s 获取 %s 类型主机失败: %v", inst.Name, hostType, err)
				continue
			}

			// 为每个主机添加实例信息
			for i := range hosts {
				hosts[i].ZID = inst.ZID
				hosts[i].InstanceName = inst.Name
			}

			allHosts = append(allHosts, hosts...)
		}

		// 将数据序列化并写入缓存
		data, err := json.Marshal(allHosts)
		if err != nil {
			logger.Log.Errorf("序列化 %s 类型主机数据失败: %v", hostType, err)
			continue
		}

		err = CacheSet(hostType+"_OVERVIEW", string(data), 0)
		if err != nil {
			logger.Log.Errorf("写入 %s 类型主机数据到缓存失败: %v", hostType, err)
			continue
		}

		// 同步写一份主机列表缓存（供资产树列表查询直接命中，避免重复打 Zabbix）
		_ = CacheSet(hostListCacheKey(hostType), string(data), hostListCacheTTL)

		logger.Log.Infof("成功同步 %s 类型主机数据到缓存，共 %d 台主机", hostType, len(allHosts))
	}

	// 新主机自动初始化检测（按主机组统计，不依赖 inventory.type）
	DetectAndInitNewHosts()

	logger.Log.Info("状态纵览数据同步完成")
	return nil
}

// newHostGroupCountKey 按绑定 ID 保存上次主机组内主机总数（用于检测新主机）
func newHostGroupCountKey(systemID int64) string {
	return fmt.Sprintf("binding_group_host_count_%d", systemID)
}

// DetectAndInitNewHosts 检测开启了 init_on_new_host 的绑定，其主机组内主机总数是否增加。
// 关键：按 group_id 统计组内主机总数（含未初始化主机），而非按 inventory.type 统计，
// 这样新加入组、尚未打类型标签的主机才能被检测到。
func DetectAndInitNewHosts() {
	var systems []System
	if err := DB.Where("init_on_new_host = 1").Find(&systems).Error; err != nil {
		logger.Log.Errorf("查询新主机自动初始化绑定失败: %v", err)
		return
	}
	if len(systems) == 0 {
		return
	}

	for _, s := range systems {
		if s.GroupID == "" {
			continue
		}
		apiInstance, err := GetAPIByZID(s.ZID)
		if err != nil {
			logger.Log.Errorf("新主机检测获取实例失败 [ID=%d, zid=%d]: %v", s.ID, s.ZID, err)
			continue
		}

		// 按主机组查询组内主机总数（countOutput，不过滤 inventory）
		groupIDs := strings.Split(s.GroupID, ",")
		rep, err := apiInstance.API.CallWithError("host.get", Params{
			"countOutput": true,
			"groupids":    groupIDs,
		})
		if err != nil {
			logger.Log.Errorf("新主机检测查询主机组失败 [ID=%d]: %v", s.ID, err)
			continue
		}
		currentCount := 0
		if cntStr, ok := rep.Result.(string); ok {
			fmt.Sscanf(cntStr, "%d", &currentCount)
		}

		countKey := newHostGroupCountKey(s.ID)
		prevCountStr, _ := CacheGet(countKey)
		prevCount := -1 // -1 表示从未记录过
		if prevCountStr != "" {
			fmt.Sscanf(prevCountStr, "%d", &prevCount)
		}

		// 更新缓存
		_ = CacheSet(countKey, fmt.Sprintf("%d", currentCount), 0)

		// 首次记录（prevCount=-1）不触发，仅建立基线；数量增加时触发初始化
		if prevCount >= 0 && currentCount > prevCount {
			logger.Log.Infof("检测到绑定 [ID=%d, type=%s] 主机组新增 %d 台主机，触发自动初始化",
				s.ID, s.TypeCode, currentCount-prevCount)
			go func(id int64) {
				if err := ExecuteSystemInit(id, "new_host"); err != nil {
					logger.Log.Errorf("新主机自动初始化失败 [ID=%d]: %v", id, err)
				}
			}(s.ID)
		}
	}
}

// getOverviewHostsFromInstance 从指定实例获取状态纵览主机数据
func getOverviewHostsFromInstance(inst *APIInstance, hostType string) ([]Hosts, error) {
	SelectInterfacesPar := []string{"ip", "port", "available", "error"}
	SearchInventoryPar := make(map[string]string)
	SearchInventoryPar["type"] = hostType
	filterPar := make(map[string]string)
	filterPar["status"] = "0" // 只获取启用的主机

	rep, err := inst.API.CallWithError("host.get", Params{
		"output":           "extend",
		"filter":           filterPar,
		"searchInventory":  SearchInventoryPar,
		"selectInventory":  "extend",
		"selectInterfaces": SelectInterfacesPar})
	if err != nil {
		return nil, err
	}

	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return nil, err
	}

	var hb ListHosts
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return nil, err
	}

	var hosts []Hosts
	for _, v := range hb {
		hosts = append(hosts, buildHostFromListHost(inst, v, hostType))
	}

	return hosts, nil
}

// AutoInitSystemBindings 自动执行启用了 auto_init 的资产绑定初始化
// AutoInitSystemBindings 立即执行一次所有 auto_init=1 的绑定初始化（用于手动触发或兜底）
func AutoInitSystemBindings() error {
	systems, err := GetAutoInitSystems()
	if err != nil {
		logger.Log.Errorf("获取自动初始化绑定配置失败: %v", err)
		return err
	}
	if len(systems) == 0 {
		return nil
	}
	logger.Log.Infof("开始自动资产绑定初始化，共 %d 个配置", len(systems))
	for _, s := range systems {
		go func(id int64) {
			if err := ExecuteSystemInit(id, "auto"); err != nil {
				logger.Log.Errorf("自动资产绑定初始化失败 [ID=%d]: %v", id, err)
			}
		}(s.ID)
	}
	return nil
}

// RegisterSystemBindingCrons 为每条 auto_init=1 的资产绑定注册独立的 cron 任务。
// 在调度器启动后调用，绑定保存时也应重新调用以更新调度。
func RegisterSystemBindingCrons() {
	cronMu.Lock()
	defer cronMu.Unlock()
	if cronScheduler == nil {
		return
	}

	systems, err := GetAutoInitSystems()
	if err != nil {
		logger.Log.Errorf("注册资产绑定 cron 失败，无法读取绑定列表: %v", err)
		return
	}

	for _, s := range systems {
		spec := s.InitCron
		if spec == "" {
			spec = "0 0 2 * * *" // 默认每天 2 点
		}
		if _, err := taskCronParser.Parse(spec); err != nil {
			logger.Log.Errorf("资产绑定 [ID=%d, type=%s] cron 表达式无效 [%s]: %v",
				s.ID, s.TypeCode, spec, err)
			continue
		}
		id := s.ID
		typCode := s.TypeCode
		_, err := cronScheduler.AddFunc(spec, func() {
			logger.Log.Infof("执行资产绑定定时初始化 [ID=%d, type=%s]", id, typCode)
			if err := ExecuteSystemInit(id, "auto"); err != nil {
				logger.Log.Errorf("资产绑定定时初始化失败 [ID=%d]: %v", id, err)
			}
		})
		if err != nil {
			logger.Log.Errorf("注册资产绑定 [ID=%d] cron 失败: %v", s.ID, err)
			continue
		}
		logger.Log.Infof("资产绑定 [ID=%d, type=%s] 定时任务已注册，Cron=%s",
			s.ID, s.TypeCode, spec)
	}
}

// RetryFailedSystemBindings 重试失败的资产绑定初始化
func RetryFailedSystemBindings() error {
	systems, err := GetFailedSystemsForRetry()
	if err != nil {
		logger.Log.Errorf("获取失败绑定配置失败: %v", err)
		return err
	}
	if len(systems) == 0 {
		return nil
	}
	logger.Log.Infof("开始重试失败的资产绑定，共 %d 个配置", len(systems))
	for _, s := range systems {
		go func(id int64) {
			if err := ExecuteSystemInit(id, "retry"); err != nil {
				logger.Log.Errorf("重试资产绑定失败 [ID=%d]: %v", id, err)
			}
		}(s.ID)
	}
	return nil
}
