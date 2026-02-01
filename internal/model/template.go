package model

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"zbxtable/pkg/logger"

	zabbix "github.com/canghai908/zabbix-go"
)

// TemplateGet func
func TemplateGet(page, limit, templates string) ([]Template, int64, error) {
	par := []string{"host", "name", "templateid"}
	hostspar := []string{"host", "name", "hostid"}
	rep, err := API.Call("template.get", Params{"output": par,
		"selectApplications": "count", "selectItems": "count",
		"selectTriggers": "count", "selectGraphs": "count",
		"selectDiscoveries": "count", "selectScreens": "count",
		"selectHosts": hostspar})
	if err != nil {
		logger.Log.Error(err)
		return []Template{}, 0, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		logger.Log.Error(err)
		return []Template{}, 0, err
	}

	var hb []Template
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logger.Log.Error(err)
		return []Template{}, 0, err
	}
	var dt []Template
	var d Template

	if templates != "" {
		for _, v := range hb {
			if strings.Contains(v.Name, templates) {
				d.Host = v.Host
				d.Templateid = v.Templateid
				d.Name = v.Name
				d.Hosts = v.Hosts
				d.Applications = v.Applications
				d.Triggers = v.Triggers
				d.Items = v.Items
				d.Graphs = v.Graphs
				d.Screens = v.Screens
				d.Discoveries = v.Discoveries
				dt = append(dt, d)
			}
		}
	} else {
		for _, v := range hb {
			d.Host = v.Host
			d.Templateid = v.Templateid
			d.Name = v.Name
			d.Hosts = v.Hosts
			d.Applications = v.Applications
			d.Triggers = v.Triggers
			d.Items = v.Items
			d.Graphs = v.Graphs
			d.Screens = v.Screens
			d.Discoveries = v.Discoveries
			dt = append(dt, d)
		}
	}

	IntPage, err := strconv.Atoi(page)
	if err != nil {
		IntPage = 1
	}
	IntLimit, err := strconv.Atoi(limit)
	if err != nil {
		IntLimit = 10
	}
	//如果dt为空直接返回
	if len(dt) == 0 {
		return dt, int64(len(dt)), err
	}
	//分页
	nums := len(dt)
	//page总数
	totalpages := int(math.Ceil(float64(nums) / float64(IntLimit)))
	if IntPage >= totalpages {
		IntPage = totalpages
	}
	if IntPage <= 0 {
		IntPage = 1
	}
	//结束页数据
	var end int
	//begin 开始页数据

	begin := (IntPage - 1) * IntLimit
	if IntPage == totalpages {
		end = nums
	}
	if IntPage < totalpages {
		end = IntPage * IntLimit
	} else {
		end = nums
	}
	//根据开始和结束返回数据列表
	var newtemplates []Template
	for i := begin; i < end; i++ {
		newtemplates = append(newtemplates, dt[i])
	}
	return newtemplates, int64(len(dt)), err
}

// TemplateAllGet func
func TemplateAllGet() ([]Template, int64, error) {
	par := []string{"host", "name", "templateid"}
	hostspar := []string{"host", "name", "hostid"}
	rep, err := API.Call("template.get", Params{"output": par,
		"selectApplications": "count", "selectItems": "count",
		"selectTriggers": "count", "selectGraphs": "count",
		"selectDiscoveries": "count", "selectScreens": "count",
		"selectHosts": hostspar})
	if err != nil {
		return []Template{}, 0, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []Template{}, 0, err
	}

	var hb []Template
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []Template{}, 0, err
	}
	return hb, int64(len(hb)), nil
}

// TemplateAllGet func
func TemplateListGet() ([]TemplateByItemList, error) {
	par := []string{"host", "name", "templateid"}
	rep, err := API.Call("template.get", Params{"output": par})
	if err != nil {
		return []TemplateByItemList{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []TemplateByItemList{}, err
	}

	var hb []TemplateByItemList
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []TemplateByItemList{}, err
	}
	return hb, nil
}

// TemplateListGetFromInstance 从指定实例获取模板列表
func TemplateListGetFromInstance(id string) ([]TemplateByItemList, error) {
	// 如果没有指定实例ID，使用全局API
	if id == "" {
		return TemplateListGet()
	}
	var instance *ZabbixInstance
	var err error

	idInt, _ := strconv.Atoi(id)
	if err == nil {
		// 如果转换成功，按 ID 查询
		instance, err = GetZabbixInstanceByZID(idInt)
	}

	if err != nil {
		return []TemplateByItemList{}, fmt.Errorf("未找到启用的实例 (instance_id=%s): %w", id, err)
	}

	// 检查实例是否启用
	if !instance.Enabled {
		return []TemplateByItemList{}, fmt.Errorf("实例未启用 (zid=%d, instance_id=%s)", instance.ID, instance.InstanceID)
	}

	// 创建 Zabbix API 实例
	apiURL := instance.URL + "/api_jsonrpc.php"
	api := zabbix.NewAPI(apiURL)

	// 设置认证
	if instance.Token != "" {
		api.Auth = instance.Token
	} else {
		_, err := api.Login(instance.User, instance.Pass)
		if err != nil {
			return []TemplateByItemList{}, fmt.Errorf("登录 Zabbix 失败: %w", err)
		}
	}

	par := []string{"host", "name", "templateid"}
	rep, err := api.Call("template.get", Params{"output": par})
	if err != nil {
		return []TemplateByItemList{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []TemplateByItemList{}, err
	}

	var hb []TemplateByItemList
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []TemplateByItemList{}, err
	}
	return hb, nil
}

// TemplateAllGet func
func TemplateByItem(templateid string) ([]TemplateByItemList, error) {
	par := []string{"host", "name", "templateid"}
	itemParams := []string{"itemid", "name"}
	rep, err := API.Call("template.get", Params{"output": par,
		"templateids": templateid,
		"selectItems": itemParams,
	})
	if err != nil {
		return []TemplateByItemList{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []TemplateByItemList{}, err
	}
	var hb []TemplateByItemList
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []TemplateByItemList{}, err
	}
	return hb, nil
}

// TemplateByItemFromInstance 从指定实例根据模板ID获取监控项
func TemplateByItemFromInstance(templateid string, zid string) ([]TemplateByItemList, error) {
	// 如果没有指定实例ID，使用全局API
	if zid == "" {
		return TemplateByItem(templateid)
	}
	var instance *ZabbixInstance
	var err error

	// 尝试将 instanceID 转换为 int64（数据库ID）
	id, _ := strconv.Atoi(zid)
	instance, err = GetZabbixInstanceByZID(id)
	if err != nil {
		return []TemplateByItemList{}, fmt.Errorf("未找到启用的实例 (instance_id=%s): %w", instance.ID, err)
	}
	// 检查实例是否启用
	if !instance.Enabled {
		return []TemplateByItemList{}, fmt.Errorf("实例未启用 (id=%d, tenant_id=%s)", instance.ID, instance.InstanceID)
	}
	// 创建 Zabbix API 实例
	apiURL := instance.URL + "/api_jsonrpc.php"
	api := zabbix.NewAPI(apiURL)

	// 设置认证
	if instance.Token != "" {
		api.Auth = instance.Token
	} else {
		_, err := api.Login(instance.User, instance.Pass)
		if err != nil {
			return []TemplateByItemList{}, fmt.Errorf("登录 Zabbix 失败: %w", err)
		}
	}

	par := []string{"host", "name", "templateid"}
	itemParams := []string{"itemid", "name"}
	rep, err := api.Call("template.get", Params{"output": par,
		"templateids": templateid,
		"selectItems": itemParams,
	})
	if err != nil {
		return []TemplateByItemList{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []TemplateByItemList{}, err
	}
	var hb []TemplateByItemList
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []TemplateByItemList{}, err
	}
	return hb, nil
}
