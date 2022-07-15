package models

import (
	"context"
	"fmt"
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

//top
func InitTask() {
	top := toolbox.NewTask("top", "0/30 * * * * *", TOP)
	//history := toolbox.NewTask("history", "0/59 * * * * *", Histroy)
	//每天执行一次
	DayReport := toolbox.NewTask("DayReport", "0 55 23 * * *", CreateDayReport)
	WeekReport := toolbox.NewTask("DayReport", "0 55 17 * * 5", CreateWeekReport)
	updatetopologydata := toolbox.NewTask("updatetopologydata", "0/30 * * * * *", UpdateTopoByStatus)
	hostTypeHostList := toolbox.NewTask("hostTypeHostList", "0 */5 * * * *", GetTypeHostList)
	Egress := toolbox.NewTask("EgressCache", "0/30 * * * * *", EgressCache)

	toolbox.AddTask("top", top)
	toolbox.AddTask("updatetopologydata", updatetopologydata)
	toolbox.AddTask("hostTypeHostList", hostTypeHostList)
	toolbox.AddTask("Egress", Egress)
	toolbox.AddTask("DayReport", DayReport)
	toolbox.AddTask("WeekReport", WeekReport)

}
func CreateWeekReport() error {
	_, p, err := GetALlReport()
	if err != nil {
		logs.Error(err)
	}
	for _, v := range p {
		if v.Status == "1" {
			if len(v.Cycle) != 0 {
				var cycle []string
				if err := json.Unmarshal([]byte(v.Cycle), &cycle); err != nil {
					logs.Error(err)
				}
				for _, vv := range cycle {
					if vv == "week" {
						start := time.Now()
						err := TaskWeekReport(v)
						if err != nil {
							logs.Error(err)
						}
						//更新report状态
						v.ExecStatus = strconv.Itoa(Success)
						v.StartAt = start
						v.EndAt = time.Now()
						err = UpdateReportExecStatusByID(&v)
						if err != nil {
							logs.Error(err)
						}
					}
				}
			}
		}
	}
	return nil
}
func CreateDayReport() error {
	_, p, err := GetALlReport()
	if err != nil {
		logs.Error(err)
	}
	for _, v := range p {
		if v.Status == "1" {
			if len(v.Cycle) != 0 {
				var cycle []string
				if err := json.Unmarshal([]byte(v.Cycle), &cycle); err != nil {
					logs.Error(err)
				}
				for _, vv := range cycle {
					if vv == "day" {
						start := time.Now()
						err := TaskDayReport(v)
						if err != nil {
							logs.Error(err)
						}
						//更新report状态
						v.ExecStatus = strconv.Itoa(Success)
						v.StartAt = start
						v.EndAt = time.Now()
						err = UpdateReportExecStatusByID(&v)
						if err != nil {
							logs.Error(err)
						}
					}
				}
			}
		}
	}
	return nil
}
func Week() error {
	fmt.Println("week")
	return nil
}
func Day() error {
	fmt.Println("day")
	return nil
}

//linux windows top data to redis
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
	for _, v := range hb {
		d.HostID = v.Hostid
		d.Host = v.Host
		d.Name = v.Name
		d.Interfaces = v.Interfaces[0].IP
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

//by topology status update data
func UpdateTopoByStatus() error {
	_, topo, err := GetDeployTopoly()
	if err != nil {
		logs.Debug(err)
		return err
	}
	for _, v := range topo {
		UpdateTopologyData(v)
	}
	return nil
}

//update topology data
func UpdateTopologyData(v Topology) error {
	var alledges AllEdge
	err := json.Unmarshal([]byte(v.Edges), &alledges)
	if err != nil {
		logs.Debug(err)
		return err
	}
	var wg sync.WaitGroup
	trigger := make(chan string)
	flow := make(chan string)
	var aedge []AEdge
	for _, v := range alledges {
		wg.Add(2)
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
			go GetFlowByFlowID(v.Attrs.Line.FlowID, &wg, flow)
			v.Labels[0].Attrs.Label.Text = <-flow
		}
		//trigger get
		if v.Attrs.Line.TriggerID != "" {
			go GetTriggerValueByTriggerID(v.Attrs.Line.TriggerID, &wg, trigger)
			status := <-trigger
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
	}
	go func() {
		wg.Wait()
		close(trigger)
		close(flow)
	}()
	aedgestr, err := json.Marshal(aedge)
	if err != nil {
		logs.Debug(err)
		return err
	}
	var Topo Topology
	Topo.ID = v.ID
	Topo.Edges = string(aedgestr)
	err = UpdateTopologyEdgesByID(&Topo)
	if err != nil {
		logs.Debug(err)
		return err
	}

	//update nodes data
	var allnodes AllNodes
	err = json.Unmarshal([]byte(v.Nodes), &allnodes)
	if err != nil {
		logs.Debug(err)
		return err
	}
	var wg2 sync.WaitGroup
	var nodes []PNodes
	for _, v := range allnodes {
		info := make(chan string)
		wg2.Add(1)
		if v.Attrs.Label.HostID != "" {
			go GetHostInfoByID(v.Attrs.Label.HostID, &wg2, info)
			var NewHost Hosts
			err := json.Unmarshal([]byte(<-info), &NewHost)
			if err != nil {
				logs.Debug(err)
				return err
			}
			v.Attrs.Label.Text = NewHost.Name
			v.Attrs.Label.HostAlarm = NewHost.Alarm
			v.Attrs.Label.HostCPU = NewHost.CPUUtilization
			v.Attrs.Label.HostMem = NewHost.MemoryUtilization
			v.Attrs.Label.HostClock = time.Now().Format("2006-01-02 15:04:05")
			v.Attrs.Label.HostStatus = NewHost.Status
			v.Attrs.Label.HostError = NewHost.Error
			nodes = append(nodes, v)
		}
		close(info)
	}
	wg2.Wait()
	alnodes, err := json.Marshal(nodes)
	if err != nil {
		logs.Debug(err)
		return err
	}
	var Topoly Topology
	Topoly.ID = v.ID
	Topoly.Nodes = string(alnodes)
	err = UpdateTopologyNodesByID(&Topoly)
	if err != nil {
		logs.Debug(err)
		return err
	}
	return nil
}

//机器列表缓存
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

//出口带宽流量获取
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
		var dlist EgressList
		dlist.NameOne = v.NameOne
		dlist.InOne = "0Kb/s"
		dlist.OutOne = "0Kb/s"
		dlist.NameTwo = v.NameTwo
		dlist.InTwo = "0Kb/s"
		dlist.OutTwo = "0Kb/s"
		dlist.Date = time.Now().Format(utils.TimeFormat)
		p1, _ := json.Marshal(&dlist)
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
		return nil
	}
	if err != nil {
		return err
	}
	var dlist EgressList
	dlist.NameOne = v.NameOne
	dlist.InOne = utils.FormatTraffic(list[0].Lastvalue)
	dlist.OutOne = utils.FormatTraffic(list[1].Lastvalue)
	dlist.NameTwo = v.NameTwo
	dlist.InTwo = utils.FormatTraffic(list[2].Lastvalue)
	dlist.OutTwo = utils.FormatTraffic(list[3].Lastvalue)
	dlist.Date = utils.UnixTimeFormater(list[0].Lastclock)
	p1, _ := json.Marshal(&dlist)
	var ctx = context.Background()
	err = RDB.Set(ctx, "Egress", string(p1), -1*time.Second).Err()
	if err != nil {
		return err
	}
	return nil

}
