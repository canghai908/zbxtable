package model

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"

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
	// toolbox 的 "0/30 * * * * *" 表示每30秒执行一次
	cronScheduler.AddFunc("0/30 * * * * *", func() { _ = TOP() })
	cronScheduler.AddFunc("0 55 23 * * *", func() { _ = CreateDayReport() })  // 每天23:55执行
	cronScheduler.AddFunc("0 55 17 * * 5", func() { _ = CreateWeekReport() }) // 每周五17:55执行
	cronScheduler.AddFunc("0 */5 * * * *", func() { _ = GetTypeHostList() })  // 每5分钟执行
	cronScheduler.AddFunc("0/30 * * * * *", func() { _ = EgressCache() })     // 每30秒执行
	cronScheduler.AddFunc("0 */5 * * * *", func() { _ = SyncInventory() })    // 每5分钟执行

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
						v.StartAt = start
						v.EndAt = time.Now()
						err = UpdateReportExecStatusByID(&v)
						if err != nil {
							logger.Log.Error(err)
						}
						continue
					}
					//update status success
					v.ExecStatus = strconv.Itoa(Success)
					v.StartAt = start
					v.EndAt = time.Now()
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
						v.StartAt = start
						v.EndAt = time.Now()
						err = UpdateReportExecStatusByID(&v)
						if err != nil {
							logger.Log.Error(err)
						}
						continue
					}
					//更新report状态
					v.ExecStatus = strconv.Itoa(Success)
					v.StartAt = start
					v.EndAt = time.Now()
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

// linux windows top data to redis
func TOP() error {
	OutputPar := []string{"hostid", "host", "available", "status", "name", "error"}
	SelectInterfacesPar := []string{"ip", "port"}
	SearchInventoryKey := []string{"VM_WIN", "VM_LIN"}
	SearchInventoryPar := make(map[string][]string)
	SearchInventoryPar["type"] = SearchInventoryKey
	rep, err := API.CallWithError("host.get", Params{
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
		return errors.New("host list is null")
	}
	for _, v := range hb {
		if v.Available == "0" {
			continue
		}
		switch v.Inventory.Type {
		case "VM_WIN":
			var float64CPU float64
			if v.Inventory.SoftwareAppA == "" {
				float64CPU = 0
			} else {
				float64CPU, err = strconv.ParseFloat(strings.Replace(v.Inventory.SoftwareAppA, " %", "", -1), 64)
				if err != nil {
					float64CPU = 0
				}
			}
			err = CacheZAdd("WIN_CPU", v.Host, float64CPU)
			if err != nil {
				return err
			}
			//memory
			var float64MEM float64

			if v.Inventory.SoftwareAppB == "" {
				float64MEM = 0
			} else {
				float64MEM, err = strconv.ParseFloat(strings.Replace(v.Inventory.SoftwareAppB, " %", "", -1), 64)
				if err != nil {
					float64MEM = 0
				}
			}
			err = CacheZAdd("WIN_MEM", v.Host, float64MEM)
			if err != nil {
				return err
			}
		case "VM_LIN":
			var float64CPU float64
			if v.Inventory.SoftwareAppA == "" {
				float64CPU = 0
			} else {
				float64CPU, err = strconv.ParseFloat(strings.Replace(v.Inventory.SoftwareAppA, " %", "", -1), 64)
				if err != nil {
					float64CPU = 0
				}
			}
			err = CacheZAdd("LIN_CPU", v.Host, float64CPU)
			if err != nil {
				return err
			}
			//memory
			var float64MEM float64
			if v.Inventory.SoftwareAppB == "" {
				float64MEM = 0
			} else {
				float64MEM, err = strconv.ParseFloat(strings.Replace(v.Inventory.SoftwareAppB, " %", "", -1), 64)
				if err != nil {
					float64MEM = 0
				}
			}
			err = CacheZAdd("LIN_MEM", v.Host, float64MEM)
			if err != nil {
				logger.Log.Debug(err)
				return err
			}
		}
	}
	return err
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
				flow, err := GetFlowByFlowID(v.Attrs.Line.FlowID)
				if err != nil {
					logger.Log.Error(err)
				}
				v.Labels[0].Attrs.Label.Text = flow
			}
			//trigger get
			if v.Attrs.Line.TriggerID != "" {
				status, err := GetTriggerValueByTriggerID(v.Attrs.Line.TriggerID)
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

// GetTypeHostList 机器列表缓存
func GetTypeHostList() error {
	//func GetHostByType(htype string) ([]TreeChildren, int, error) {
	var list = []string{"VM_LIN", "VM_WIN", "HW_NET", "HW_SRV"}
	for _, v := range list {
		p, _, err := GetHostsList(v)
		if err != nil {
			logger.Log.Error(err)
			continue
		}
		//hosts info to cache
		hostsdata, err := json.Marshal(p)
		if err != nil {
			logger.Log.Error(err)
		}
		err = CacheSet(v+"_OVERVIEW", string(hostsdata), 3600*time.Second)
		if err != nil {
			logger.Log.Error(err)
			continue
		}
		//inventor info to cache
		var t TreeChildren
		var tt []TreeChildren
		for _, vv := range p {
			var tid int64
			var err error
			tid, err = strconv.ParseInt(vv.HostID, 10, 64)
			if err != nil {
				tid = 0
				logger.Log.Error(err)
			}
			t.ID = tid
			t.Name = vv.Name
			tt = append(tt, t)
		}
		data, err := json.Marshal(tt)
		if err != nil {
			logger.Log.Error(err)
		}
		err = CacheSet(v+"_INVENTORY", string(data), 3600*time.Second)
		if err != nil {
			logger.Log.Error(err)
			continue
		}
	}
	return nil
}

// EgressCache 出口带宽流量获取并写入缓存
func EgressCache() error {
	var v Egress
	err := DB.First(&v, 1).Error
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	var itemList []string
	//空返回
	if v.InOne == "" && v.OutOne == "" && v.InTwo == "" && v.OutTwo == "" {
		var dList EgressList
		dList.NameOne = v.NameOne
		dList.InOne = "0Kb/s"
		dList.OutOne = "0Kb/s"
		dList.NameTwo = v.NameTwo
		dList.InTwo = "0Kb/s"
		dList.OutTwo = "0Kb/s"
		dList.Date = time.Now().Format(utils.TimeFormat)
		p1, _ := json.Marshal(&dList)
		// 使用0表示永不过期
		err = CacheSet("Egress", string(p1), 0)
		if err != nil {
			return err
		}
		return nil
	}
	itemList = append(itemList, v.InOne, v.OutOne, v.InTwo, v.OutTwo)
	list, err := GetItemByIDS(itemList)
	if err != nil {
		logger.Log.Error("出口Item数据获取异常", err)
		return errors.New("出口Item数据获取异常")
	}
	//数据异常返回
	if len(list) != 4 {
		logger.Log.Error("出口Item数据结果异常", len(list))
		return errors.New("出口Item数据结果异常")
	}
	var dList EgressList
	dList.NameOne = v.NameOne
	dList.InOne = utils.FormatTraffic(list[0].Lastvalue)
	dList.OutOne = utils.FormatTraffic(list[1].Lastvalue)
	dList.NameTwo = v.NameTwo
	dList.InTwo = utils.FormatTraffic(list[2].Lastvalue)
	dList.OutTwo = utils.FormatTraffic(list[3].Lastvalue)
	dList.Date = utils.UnixTimeFormater(list[0].Lastclock)
	p1, _ := json.Marshal(&dList)
	// 使用0表示永不过期
	err = CacheSet("Egress", string(p1), 0)
	if err != nil {
		return err
	}
	return nil
}

// SyncInventory 同步主机分类及数据绑定
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
	if data[0].Value != "1" {
		return nil
	}
	var list []System
	err = GetDB().Where("id = ?", 1).Find(&list).Error
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	if len(list) == 0 {
		return nil
	}
	for _, v := range list {
		gList := strings.Split(v.GroupID, ",")
		err = HostTypeSet(&v, gList)
		if err != nil {
			return err
		}
	}
	return nil
}
