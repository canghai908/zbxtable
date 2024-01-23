package models

import (
	"context"
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/toolbox"
	"github.com/go-redis/redis/v8"
	"strconv"
	"strings"
	"sync"
	"time"
	"zbxtable/utils"
)

// top
func InitTask() {
	top := toolbox.NewTask("top", "0/30 * * * * *", TOP)
	//history := toolbox.NewTask("history", "0/59 * * * * *", Histroy)
	//每天执行一次
	DayReport := toolbox.NewTask("DayReport", "0 55 23 * * *", CreateDayReport)
	WeekReport := toolbox.NewTask("DayReport", "0 55 17 * * 5", CreateWeekReport)
	//拓朴图数据更新
	//UpdateTopoData := toolbox.NewTask("UpdateTopoData", "0/30 * * * * *", UpdateTopoData)
	hostTypeHostList := toolbox.NewTask("hostTypeHostList", "0 */5 * * * *", GetTypeHostList)
	//出口带宽流量获取
	Egress := toolbox.NewTask("EgressCache", "0/30 * * * * *", EgressCache)

	toolbox.AddTask("top", top)
	//toolbox.AddTask("UpdateTopoData", UpdateTopoData)
	toolbox.AddTask("hostTypeHostList", hostTypeHostList)
	toolbox.AddTask("Egress", Egress)
	toolbox.AddTask("DayReport", DayReport)
	toolbox.AddTask("WeekReport", WeekReport)
}
func CreateWeekReport() error {
	_, list, err := GetALlReport()
	if err != nil {
		logs.Error(err)
		return err
	}
	//遍历周报
	for _, v := range list {
		//启用周报
		if v.Status == "1" && len(v.Cycle) != 0 && len(v.Items) != 0 {
			cycList := strings.Split(v.Cycle, ",")
			for _, vv := range cycList {
				if vv == "week" {
					start := time.Now()
					//周报生成
					err := TaskWeekReport(v)
					if err != nil {
						logs.Error(err)
						///update status failed
						v.ExecStatus = strconv.Itoa(Failed)
						v.StartAt = start
						v.EndAt = time.Now()
						err = UpdateReportExecStatusByID(&v)
						if err != nil {
							logs.Error(err)
						}
						continue
					}
					//update status success
					v.ExecStatus = strconv.Itoa(Success)
					v.StartAt = start
					v.EndAt = time.Now()
					err = UpdateReportExecStatusByID(&v)
					if err != nil {
						logs.Error(err)
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
		logs.Error(err)
		return err
	}
	for _, v := range list {
		if v.Status == "1" && len(v.Cycle) != 0 && len(v.Items) != 0 {
			cycList := strings.Split(v.Cycle, ",")
			for _, vv := range cycList {
				if vv == "day" {
					start := time.Now()
					err := TaskDayReport(v)
					if err != nil {
						logs.Error(err)
						//更新report状态
						v.ExecStatus = strconv.Itoa(Failed)
						v.StartAt = start
						v.EndAt = time.Now()
						err = UpdateReportExecStatusByID(&v)
						if err != nil {
							logs.Error(err)
						}
						continue
					}
					//更新report状态
					v.ExecStatus = strconv.Itoa(Success)
					v.StartAt = start
					v.EndAt = time.Now()
					err = UpdateReportExecStatusByID(&v)
					if err != nil {
						logs.Error(err)
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
		logs.Debug(err)
		return err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Debug(err)
		return err
	}
	var hb ListHosts
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Debug(err)
		return err
	}
	var ctx = context.Background()
	//var dt []Hosts
	var d Hosts
	if len(hb) == 0 {
		logs.Error(errors.New("host list is null"))
		return errors.New("host list is null")
	}
	for _, v := range hb {
		d.HostID = v.Hostid
		d.Host = v.Host
		d.Name = v.Name
		if len(v.Interfaces) != 0 {
			d.Interfaces = v.Interfaces[0].IP
		}
		d.Status = v.Status
		d.Available = v.Available
		d.Error = v.Error
		d.NumberOfCores = v.Inventory.Software
		d.CPUUtilization = v.Inventory.SoftwareAppA
		d.MemoryUtilization = v.Inventory.SoftwareAppB
		d.MemoryUsed = v.Inventory.SoftwareAppD
		d.MemoryTotal = v.Inventory.SoftwareAppC
		d.Uptime = v.Inventory.SoftwareAppE
		//排除异常主机
		if d.Available == "0" {
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
			err := RDB.ZAdd(ctx, "WIN_CPU", &redis.Z{
				Member: v.Host,
				Score:  float64CPU,
			}).Err()
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
			err = RDB.ZAdd(ctx, "WIN_MEM", &redis.Z{
				Member: v.Host,
				Score:  float64MEM,
			}).Err()
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
			err := RDB.ZAdd(ctx, "LIN_CPU", &redis.Z{
				Member: v.Host,
				Score:  float64CPU,
			}).Err()
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
			err = RDB.ZAdd(ctx, "LIN_MEM", &redis.Z{
				Member: v.Host,
				Score:  float64MEM,
			}).Err()
			if err != nil {
				logs.Debug(err)
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
		logs.Debug(err)
		return err
	}
	var allEdges AllEdge
	err = json.Unmarshal([]byte(p.Edges), &allEdges)
	if err != nil {
		logs.Debug(err)
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
					logs.Error(err)
				}
				v.Labels[0].Attrs.Label.Text = flow
			}
			//trigger get
			if v.Attrs.Line.TriggerID != "" {
				status, err := GetTriggerValueByTriggerID(v.Attrs.Line.TriggerID)
				if err != nil {
					logs.Error(err)
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
		logs.Debug(err)
		return err
	}
	var Topo Topology
	Topo.ID = id
	Topo.Edges = string(edgeStr)
	err = UpdateTopologyEdgesByID(&Topo)
	if err != nil {
		logs.Debug(err)
		return err
	}
	return nil
}

// GetTypeHostList 机器列表缓存
func GetTypeHostList() error {
	//func GetHostByType(htype string) ([]TreeChildren, int, error) {
	var list = []string{"VM_LIN", "VM_WIN", "HW_NET", "HW_SRV"}
	var ctx = context.Background()
	for _, v := range list {
		p, _, err := GetHostsList(v)
		if err != nil {
			logs.Error(err)
			return err
			continue
		}
		//hosts info to redis
		hostsdata, err := json.Marshal(p)
		if err != nil {
			logs.Error(err)
		}
		err = RDB.Set(ctx, v+"_OVERVIEW", string(hostsdata), 3600*time.Second).Err()
		if err != nil {
			logs.Error(err)
			return err
			continue
		}
		//inventor info to redis
		var t TreeChildren
		var tt []TreeChildren
		for _, vv := range p {
			var tid int64
			var err error
			tid, err = strconv.ParseInt(vv.HostID, 10, 64)
			if err != nil {
				logs.Error(err)
				tid = 0
			}
			t.ID = tid
			t.Name = vv.Name
			tt = append(tt, t)
		}
		data, err := json.Marshal(tt)
		if err != nil {
			logs.Error(err)
		}
		err = RDB.Set(ctx, v+"_INVENTORY", string(data), 3600*time.Second).Err()
		if err != nil {
			logs.Error(err)
			return err
			continue
		}
	}
	return nil
}

// EgressCache 出口带宽流量获取并写入redis
func EgressCache() error {
	o := orm.NewOrm()
	v := &Egress{ID: 1}
	err := o.Read(v)
	if err != nil {
		return err
	}
	var itemlist []string
	//空返回
	if v.InOne == "" || v.OutOne == "" || v.InTwo == "" || v.OutTwo == "" {
		var dList EgressList
		dList.NameOne = v.NameOne
		dList.InOne = "0Kb/s"
		dList.OutOne = "0Kb/s"
		dList.NameTwo = v.NameTwo
		dList.InTwo = "0Kb/s"
		dList.OutTwo = "0Kb/s"
		dList.Date = time.Now().Format(utils.TimeFormat)
		p1, _ := json.Marshal(&dList)
		var ctx = context.Background()
		err = RDB.Set(ctx, "Egress", string(p1), -1*time.Second).Err()
		if err != nil {
			return err
		}
		return nil
	}
	itemlist = append(itemlist, v.InOne, v.OutOne, v.InTwo, v.OutTwo)
	list, err := GetItemByIDS(itemlist)
	//数据异常返回
	if len(list) != 4 {
		logs.Error("出口Item数据获取异常")
		return errors.New("出口Item数据获取异常")
	}
	if err != nil {
		return err
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
	var ctx = context.Background()
	err = RDB.Set(ctx, "Egress", string(p1), -1*time.Second).Err()
	if err != nil {
		return err
	}
	return nil

}
