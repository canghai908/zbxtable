package model

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"zbxtable/pkg/logger"
)

// GetZmsTopologyById retrieves ZmsTopology by Id. Returns error if
// Id doesn't exist
func GetTopologyById(id int) (v *Topology, err error) {
	v = &Topology{}
	err = DB.Where("id = ?", id).First(v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetAllTopology t
func GetAllTopology(page, limit, name string) (cnt int64, topo []Topology, err error) {
	var topologys []Topology
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	if limits == 0 {
		limits = 10
	}
	if pages == 0 {
		pages = 1
	}

	query := DB.Model(&Topology{})
	if name != "" {
		query = query.Where("topology LIKE ?", "%"+name+"%")
	}

	// 获取总数
	err = query.Count(&cnt).Error
	if err != nil {
		return 0, []Topology{}, err
	}

	// 获取分页数据
	offset := (pages - 1) * limits
	err = query.Order("created_at").Limit(limits).Offset(offset).Find(&topologys).Error
	if err != nil {
		return 0, []Topology{}, err
	}
	return cnt, topologys, nil
}

// GetAllTopology t
func GetDeployTopoly() (topo []*Topology, err error) {
	var list []*Topology
	err = DB.Where("status = ?", "1").Find(&list).Error
	if err != nil {
		return []*Topology{}, err
	}
	return list, nil
}

// AddTopology insert a new ZmsTopology into database and returns
// last inserted Id on success.
func AddTopology(m *Topology) (id int64, err error) {
	m.Status = "0"
	err = DB.Create(m).Error
	if err != nil {
		logger.Log.Debug(err)
		return 0, err
	}
	return int64(m.ID), err
}

// UpdateTopologyByID updates Alarm by Id and returns error if
func UpdateTopologyByID(m *Topology) (err error) {
	var v Topology
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err == nil {
		err = DB.Model(&Topology{}).Where("id = ?", m.ID).Updates(m).Error
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

// UpdateTopologyEdgesByID updates Alarm by Id and returns error if
func UpdateTopologyEdgesByID(m *Topology) (err error) {
	var v Topology
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err == nil {
		err = DB.Model(&Topology{}).Where("id = ?", m.ID).Update("edges", m.Edges).Error
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

// UpdateTopologyEdgesByID updates Alarm by Id and returns error if
func UpdateTopologyNodesByID(m *Topology) (err error) {
	var v Topology
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err == nil {
		err = DB.Model(&Topology{}).Where("id = ?", m.ID).Update("nodes", m.Nodes).Error
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

// UpdateTopologyEdgesByID updates Alarm by Id and returns error if
func UpdateTopologyStatusByID(m *Topology) (err error) {
	var v Topology
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err == nil {
		newStatus := "1"
		if v.Status == "0" {
			newStatus = "1"
		} else {
			newStatus = "0"
		}
		err = DB.Model(&Topology{}).Where("id = ?", m.ID).Update("status", newStatus).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteAlarm deletes Alarm by Id and returns error if
// the record to be deleted doesn't exist
func DeleteTopology(id int) (err error) {
	err = DB.Delete(&Topology{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

// GetTopologyFromWeather node
func GetTopologyFromWeather() (Data, error) {
	//get hw-switch
	OutputPar := []string{"hostid", "host", "name"}
	SearchInventoryInventoryPar := make(map[string]string)
	SearchInventoryInventoryPar["type"] = "HW_NET"
	rep, err := API.CallWithError("host.get", Params{
		"output":          OutputPar,
		"searchInventory": SearchInventoryInventoryPar})
	if err != nil {
		logger.Log.Debug(err)
		return Data{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logger.Log.Debug(err)
		return Data{}, err
	}
	var hb ListHosts
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logger.Log.Debug(err)
		return Data{}, err
	}

	//split part.conf
	var lines [][]string
	f, err := os.Open("./part.conf")
	if err != nil {
		logger.Log.Debug(err)
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
