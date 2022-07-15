package models

import (
	"bufio"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"os"
	"strconv"
	"strings"
)

//TableName Topology
func (t *Topology) TableName() string {
	return TableName("topology")
}

// GetZmsTopologyById retrieves ZmsTopology by Id. Returns error if
// Id doesn't exist
func GetTopologyById(id int) (v *Topology, err error) {
	o := orm.NewOrm()
	v = &Topology{ID: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllTopology t
func GetAllTopology(page, limit, name string) (cnt int64, topo []Topology, err error) {
	o := orm.NewOrm()
	var topologys []Topology
	var CountTopologys []Topology
	al := new(Topology)
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	//count topology
	_, err = o.QueryTable(al).Filter("topology__contains", name).All(&CountTopologys)
	_, err = o.QueryTable(al).Limit(limits, (pages-1)*limits).OrderBy("created_at").Filter("topology__contains", name).All(&topologys)
	if err != nil {
		logs.Debug(err)
		return 0, []Topology{}, err
	}
	cnt = int64(len(CountTopologys))
	return cnt, topologys, nil
}

// GetAllTopology t
func GetDeployTopoly() (cnt int64, topo []Topology, err error) {
	o := orm.NewOrm()
	var topologys []Topology
	var CountTopologys []Topology
	al := new(Topology)
	//count topology
	_, err = o.QueryTable(al).Filter("status", "1").All(&CountTopologys)
	if err != nil {
		logs.Debug(err)
		return 0, []Topology{}, err
	}
	_, err = o.QueryTable(al).OrderBy("-created_at").Filter("status", "1").All(&topologys)
	if err != nil {
		logs.Debug(err)
		return 0, []Topology{}, err
	}
	cnt = int64(len(CountTopologys))
	return cnt, topologys, nil
}

// AddTopology insert a new ZmsTopology into database and returns
// last inserted Id on success.
func AddTopology(m *Topology) (id int64, err error) {
	o := orm.NewOrm()
	m.Status = "0"
	id, err = o.Insert(m)
	if err != nil {
		logs.Debug(err)
		return 0, err
	}
	return id, err
}

// UpdateTopologyByID updates Alarm by Id and returns error if
func UpdateTopologyByID(m *Topology) (err error) {
	o := orm.NewOrm()
	v := Topology{ID: m.ID}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		_, err = o.Update(m)
		if err != nil {
			return err
		}
	}
	return
}

// UpdateTopologyEdgesByID updates Alarm by Id and returns error if
func UpdateTopologyEdgesByID(m *Topology) (err error) {
	o := orm.NewOrm()
	v := Topology{ID: m.ID}
	if err = o.Read(&v); err == nil {
		v.Edges = m.Edges
		_, err = o.Update(m, "Edges")
		if err != nil {
			return err
		}
	}
	return
}

// UpdateTopologyEdgesByID updates Alarm by Id and returns error if
func UpdateTopologyNodesByID(m *Topology) (err error) {
	o := orm.NewOrm()
	v := Topology{ID: m.ID}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		v.Nodes = m.Nodes
		_, err = o.Update(m, "Nodes")
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateTopologyEdgesByID updates Alarm by Id and returns error if
func UpdateTopologyStatusByID(m *Topology) (err error) {
	o := orm.NewOrm()
	v := Topology{ID: m.ID}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		if v.Status == "0" {
			m.Status = "1"
		} else {
			m.Status = "0"
		}
		_, err = o.Update(m, "Status")
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteAlarm deletes Alarm by Id and returns error if
// the record to be deleted doesn't exist
func DeleteTopology(id int) (err error) {
	o := orm.NewOrm()
	v := Topology{ID: id}
	if err = o.Read(&v); err == nil {
		_, err = o.Delete(&Topology{ID: id})
		if err != nil {
			return err
		}
	}
	return nil
}

//GetTopologyFromWeather node
func GetTopologyFromWeather() (Data, error) {
	//get hw-switch
	OutputPar := []string{"hostid", "host", "name"}
	SearchInventoryInventoryPar := make(map[string]string)
	SearchInventoryInventoryPar["type"] = "HW_NET"
	rep, err := API.CallWithError("host.get", Params{
		"output":          OutputPar,
		"searchInventory": SearchInventoryInventoryPar})
	if err != nil {
		logs.Debug(err)
		return Data{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logs.Debug(err)
		return Data{}, err
	}
	var hb ListHosts
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Debug(err)
		return Data{}, err
	}

	//split part.conf
	var lines [][]string
	f, err := os.Open("./part.conf")
	if err != nil {
		logs.Debug(err)
		return Data{}, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	newLine := true
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if len(line) == 0 {
			newLine = true
			continue
		}
		if newLine {
			newLine = false
			lines = append(lines, make([]string, 0))
		}
		last := len(lines) - 1
		lines[last] = append(lines[last], line)
	}
	if err := s.Err(); err != nil {
		return Data{}, err
	}
	var pp Nodes
	var plist []Nodes
	for _, line := range lines {
		//fmt.Println(k, len(line), ":", strings.Join(line, " || "))
		//根据设备名称确定设备位置及id
		if strings.Contains(line[0], "NODE ") {
			id := strings.Split(line[0], " ")
			name := strings.Split(line[1], " ")
			POS := strings.Split(line[3], " ")
			for _, v := range hb {
				if name[1] == v.Name {
					pp.ID = v.Hostid
					pp.Label = name[1]
					pp.Text = name[1]
					pp.InternalName = id[1]
					pp.X = POS[1]
					pp.Y = POS[2]
					pp.Type = "rect"
					plist = append(plist, pp)
				}
			}
		}
	}
	//获取边数据
	var eps Edges
	var eds []Edges
	for _, line := range lines {
		if strings.Contains(line[0], "LINK ") {
			li := strings.Split(line[0], " ")
			plik := strings.Split(li[1], "-")
			sor := plik[0]
			tar := plik[1]
			//遍历边的nodeid
			for _, v := range plist {
				eps.Style.StartArrow = true
				eps.Style.EndArrow = true
				eps.Style.Stroke = "#33cc33"
				eps.Style.LineWidth = 5
				//	eps.Type = "quadratic"
				if sor == v.InternalName {
					eps.Source = v.ID
					for _, v := range plist {
						if tar == v.InternalName {
							eps.Target = v.ID
							eds = append(eds, eps)
						}
					}
				}
			}
		}
	}
	var ppt Data
	ppt.Nodes = plist
	ppt.Edges = eds
	return ppt, nil
}
