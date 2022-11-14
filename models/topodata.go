package models

import (
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"strings"
	"sync"
	"zbxtable/utils"
)

// last inserted Id on success.
func AddTopoData(m *TopologyData) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	if err != nil {
		return 0, err
	}
	return id, err
}

// GetAllTopology t
func GetAllTopoData() (cnt int64, topodata []TopologyData, err error) {
	o := orm.NewOrm()
	var topologys []TopologyData
	al := new(TopologyData)
	//count topology
	_, err = o.QueryTable(al).OrderBy("-created_at").All(&topologys)
	if err != nil {
		logs.Debug(err)
		return 0, []TopologyData{}, err
	}
	cnt = int64(len(topologys))
	return cnt, topologys, nil
}

//func UpdateLineData(v AEdge) {
//	//labels attr
//	v.Labels[0].Attrs.Label.Text = ""
//	v.Labels[0].Position.Angle = 0
//	v.Labels[0].Position.Offset = 20
//	v.Labels[0].Position.Options.EnsureLegibility = true
//	v.Labels[0].Position.Options.KeepGradient = true
//	//line attrs
//	v.Attrs.Line.StrokeWidth = 4
//	v.Attrs.Line.Stroke = "#A4A4A4"
//	v.Attrs.Line.StrokeDasharray = 0
//	if v.Attrs.Line.FlowID != "" {
//		GetFlowByFlowID(v.Attrs.Line.FlowID, &wg, flow)
//		v.Labels[0].Attrs.Label.Text = <-flow
//	}
//	//trigger get
//	if v.Attrs.Line.TriggerID != "" {
//		go GetTriggerValueByTriggerID(v.Attrs.Line.TriggerID, &wg, trigger)
//		status := <-trigger
//		switch {
//		//trigger正常 未告警
//		case status == "0":
//			v.Attrs.Line.Stroke = "#00FF00"
//			v.Attrs.Line.StrokeDasharray = 5
//			v.Attrs.Line.Style.Animation = "ant-line 30s infinite linear"
//			//trigger 告警
//		case status == "1":
//			v.Attrs.Line.Stroke = "#FF0000"
//		case status == "2":
//			v.Attrs.Line.Stroke = "#A4A4A4"
//		default:
//			v.Attrs.Line.Stroke = "#A4A4A4"
//		}
//	}
//	fmt.Println(v.Attrs.Line.Stroke)
//	aedge = append(aedge, v)
//}
////go func() {
////	wg.Wait()
////	close(trigger)
////	close(flow)
////}()
////wg.Wait()
//fmt.Println("AAAA")
//aedgestr, err := json.Marshal(aedge)
//if err != nil {
//logs.Debug(err)
//return err
//}
//var Topo Topology
//Topo.ID = v.ID
//Topo.Edges = string(aedgestr)
//err = UpdateTopologyEdgesByID(&Topo)
//if err != nil {
//logs.Debug(err)
//return err
//}

func GetTriggerValueByTriggerID(TriggerID string) (value string, err error) {
	tri, err := GetTriggerValue(TriggerID)
	if err != nil || len(tri) == 0 {
		logs.Debug(err)
		return "2", err
	}
	return tri[0].Value, nil
}

func GetFlowByFlowID(FLowID string) (flow string, err error) {
	p, err := GetItemByID(FLowID)
	if err != nil {
		logs.Debug(err)
		return "", err
	}
	if len(p) == 0 {
		return "", errors.New("flow id is null")
	} else {
		//获取数据，判断是否为流量接口
		var NewItemKey string
		if strings.Contains(p[0].Name, "Bits sent") {
			NewkeyIN := strings.Replace(p[0].Key, "net.if.out", "net.if.in", -1)
			NewItemKey = strings.Replace(NewkeyIN, "Out", "In", -1)
		}
		if strings.Contains(p[0].Name, "Bits received") {
			NewkeyOut := strings.Replace(p[0].Key, "net.if.in", "net.if.out", -1)
			NewItemKey = strings.Replace(NewkeyOut, "In", "Out", -1)
		}
		NetItem, err := GetItemByKey(p[0].Hostid, NewItemKey)
		if err != nil {
			return "", err
		}
		flow = utils.FormatTraffic(NetItem[0].Lastvalue) + "/" + utils.FormatTraffic(p[0].Lastvalue)
		return flow, nil
	}
}

//host info
func GetHostInfoByID(hostid string, wg *sync.WaitGroup, info chan string) {
	defer wg.Done()
	p, err := GetHostInfoTopology(hostid)
	if err != nil {
		logs.Debug(err)
	}
	StrP, err := json.Marshal(&p)
	if err != nil {
		logs.Debug(err)
	}
	info <- string(StrP)
	return
}
