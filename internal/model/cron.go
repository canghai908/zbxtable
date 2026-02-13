package model

import (
	"strconv"
	"strings"
	"sync"
	"time"
	"zbxtable/pkg/logger"

	"github.com/robfig/cron/v3"
)

var (
	cronScheduler *cron.Cron
)

// InitTask 初始化定时任务
func InitTask() {
	// 创建 cron 调度器
	cronScheduler = cron.New(cron.WithSeconds())

	// 添加任务
	// 注意：cron 表达式格式为 "秒 分 时 日 月 周"
	cronScheduler.AddFunc("0/10 * * * * *", func() { _ = TOP() })
	cronScheduler.AddFunc("0 55 23 * * *", func() { _ = CreateDayReport() })  // 每天23:55执行
	cronScheduler.AddFunc("0 55 17 * * 5", func() { _ = CreateWeekReport() }) // 每周五17:55执行
	cronScheduler.AddFunc("0 */5 * * * *", func() { _ = SyncInventory() })    // 每5分钟执行

	// 新增：自动指标映射任务（每小时检查一次）
	cronScheduler.AddFunc("0 0 * * * *", func() { _ = AutoMetricMapping() })

	// 新增：失败重试任务（每30分钟检查一次）
	cronScheduler.AddFunc("0 */30 * * * *", func() { _ = RetryFailedMappings() })

	// 新增：出口数据采集任务（每分钟执行一次）
	cronScheduler.AddFunc("0 * * * * *", func() { _ = CollectEgressData() })

	// 新增：状态纵览数据采集任务（每5分钟执行一次）
	cronScheduler.AddFunc("0 */5 * * * *", func() { _ = SyncOverviewData() })

	// 启动调度器
	cronScheduler.Start()
	logger.Log.Info("Cron scheduler started")
}

// StopTask 停止定时任务
func StopTask() {
	if cronScheduler != nil {
		cronScheduler.Stop()
		logger.Log.Info("Cron scheduler stopped")
	}
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

	// 从配置读取 Top 数量，用于控制写入缓存的数据量（兜底 5）
	topNumStr := GetConfigValueByKey("dash_top_num", "5")
	topN, err := strconv.ParseInt(topNumStr, 10, 64)
	if err != nil || topN <= 0 {
		topN = 5
	}

	// 清空旧数据
	_ = CacheDelete("WIN_CPU")
	_ = CacheDelete("WIN_MEM")
	_ = CacheDelete("LIN_CPU")
	_ = CacheDelete("LIN_MEM")

	// 从所有实例收集数据
	for _, inst := range instances {
		err := TOPFromInstance(inst, topN)
		if err != nil {
			logger.Log.Errorf("从实例 %s 收集 TOP 数据失败: %v", inst.Name, err)
			continue
		}
	}

	return nil
}

// TOPFromInstance 从指定实例收集 TOP 数据
func TOPFromInstance(inst *APIInstance, topN int64) error {
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
	writeUnion := func(cpuKey, memKey string, cpuList, memList []hostScore) error {
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

	if err := writeUnion("WIN_CPU", "WIN_MEM", winCPUs, winMEMs); err != nil {
		return err
	}
	if err := writeUnion("LIN_CPU", "LIN_MEM", linCPUs, linMEMs); err != nil {
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
	var data []Config
	//查询配置表，id 3为同步配置
	err := DB.Where("id = ?", 3).Find(&data).Error
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	//1为开启，其他为关闭
	if data[0].ConfigValue != "1" {
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
		err = GetDB().Where("instance = ? AND status = ?", inst.ZID, 1).Find(&list).Error
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

	// 定义主机类型列表
	hostTypes := []string{"VM_LIN", "VM_WIN", "HW_NET", "HW_SRV"}

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

		logger.Log.Infof("成功同步 %s 类型主机数据到缓存，共 %d 台主机", hostType, len(allHosts))
	}

	logger.Log.Info("状态纵览数据同步完成")
	return nil
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
		var d Hosts
		d.HostID = v.Hostid
		d.Host = v.Host
		d.Name = v.Name
		if len(v.Interfaces) != 0 {
			d.Interfaces = v.Interfaces[0].IP
			d.Available = v.Interfaces[0].Available
			d.Error = v.Interfaces[0].Error
		}
		d.Status = v.Status
		d.Model = v.Inventory.Model
		d.OS = v.Inventory.Os
		d.NumberOfCores = v.Inventory.Software
		d.CPUUtilization = v.Inventory.SoftwareAppA
		d.MemoryUtilization = v.Inventory.SoftwareAppB
		d.MemoryTotal = v.Inventory.SoftwareAppC
		d.MemoryUsed = v.Inventory.SoftwareAppD
		d.Uptime = v.Inventory.SoftwareAppE
		d.DateHwInstall = v.Inventory.DateHwInstall
		d.DateHwExpiry = v.Inventory.DateHwExpiry
		d.MAC = v.Inventory.MacaddressA
		d.ResourceID = v.Inventory.SerialnoB
		d.Vendor = v.Inventory.Vendor
		d.Ping = v.Inventory.Poc1Name
		d.PingLoss = v.Inventory.Poc1Email
		d.PingSec = v.Inventory.Poc1PhoneA

		// 网络设备和物理服务器的特殊字段
		if hostType == "HW_NET" || hostType == "HW_SRV" {
			d.SerialNo = v.Inventory.SerialnoA
			d.Location = v.Inventory.Location
			d.Department = v.Inventory.SiteCity
		}

		if !inst.IsV54OrLater {
			//网络和设备
			if hostType == "HW_NET" || hostType == "HW_SRV" {
				d.Available = v.SnmpAvailable
				d.Error = v.SnmpError

			} else {
				//主机设备
				d.Available = v.Available
				d.Error = v.Error
			}

		}
		hosts = append(hosts, d)
	}

	return hosts, nil
}
