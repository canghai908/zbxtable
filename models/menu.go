package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

type Menu struct {
	Id          int    `orm:"column(id);auto"`
	ParentId    int    `orm:"column(parent_id);default(0)"`
	Name        string `orm:"column(name);size(50)"`
	Path        string `orm:"column(path);size(100)"`
	Router      string `orm:"column(router);size(100)"`
	Icon        string `orm:"column(icon);size(50)"`
	Role        string `orm:"column(role);size(50)"`
	Permission  string `orm:"column(permission);size(100)"`
	Invisible   bool   `orm:"column(in_visible);default(false)"`
	IsAvailable bool   `orm:"column(is_available);default(false)"`
	Highlight   string `orm:"column(highlight);size(100)"`
	CacheAble   bool   `orm:"column(cacheAble);default(false)"`
}

func (t *Menu) TableName() string {
	return TableName("menu")
}

// InitMenuData 菜单初始化
func InitMenuData() {
	o := orm.NewOrm()
	count, err := o.QueryTable(Menu{}).Count()
	if err != nil {
		logs.Error(err)
	}
	if count > 0 {
		return
	}
	menus := []Menu{
		// 一级菜单
		{ParentId: 0, Name: "工作台", Path: "dashboard", Router: "dashboard", Icon: "dashboard", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "主机管理", Path: "host", Router: "host", Icon: "hdd", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "网络管理", Path: "net", Router: "net", Icon: "cloud", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "硬件管理", Path: "server", Router: "server", Icon: "database", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "告警管理", Path: "alarm", Router: "alarm", Icon: "alert", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "拓扑管理", Path: "topology", Router: "topology", Icon: "picture", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "报表管理", Path: "report", Router: "report", Icon: "file", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "系统管理", Path: "system", Router: "system", Icon: "setting", CacheAble: false, Role: "admin,user"},

		// 二级菜单
		//dashboard
		{ParentId: 1, Name: "首页", Path: "workplace", Router: "workplace", Icon: "home", Role: "admin,user"},
		{ParentId: 1, Name: "资产管理", Path: "inventory", Router: "inventory", Icon: "calendar", Role: "admin,user"},
		{ParentId: 1, Name: "状态总览", Path: "overview", Router: "overview", Icon: "appstore", Role: "admin,user"},
		{ParentId: 1, Name: "数据面板", Path: "dash", Router: "dash", Icon: "block", IsAvailable: true, Role: "admin,user"},
		//主机管理
		{ParentId: 2, Name: "Linux主机", Path: "linux", Router: "linux", Icon: "container", Role: "admin,user"},
		{ParentId: 2, Name: "Linux主机详情", Path: "lindetail", Router: "linDetail", Invisible: true, Highlight: "/host", Role: "admin,user"},
		{ParentId: 2, Name: "Windows主机", Path: "windows", Router: "windows", Icon: "windows", Role: "admin,user"},
		{ParentId: 2, Name: "Windows主机详情", Path: "windetail", Router: "winDetail", Invisible: true, Highlight: "/host", Role: "admin,user"},
		//网络管理
		{ParentId: 3, Name: "网络设备", Path: "list", Router: "netList", Icon: "chrome", Role: "admin,user"},
		{ParentId: 3, Name: "设备详情", Path: "detail", Router: "netDetail", Invisible: true, Highlight: "/net", Role: "admin,user"},
		//硬件管理
		{ParentId: 4, Name: "物理服务器", Path: "list", Router: "srvList", Icon: "mobile", Role: "admin,user"},
		{ParentId: 4, Name: "光纤交换机", Path: "fiber", Router: "sanList", Icon: "mobile", Role: "admin,user"},
		{ParentId: 4, Name: "存储设备", Path: "storage", Router: "stoList", Icon: "mobile", Role: "admin,user"},
		{ParentId: 4, Name: "设备详情", Path: "detail", Router: "srvDetail", Invisible: true, Highlight: "/server", Role: "admin,user"},
		//告警管理
		{ParentId: 5, Name: "告警分析", Path: "analysis", Router: "alarmAnalysis", Icon: "hourglass", Role: "admin,user"},
		{ParentId: 5, Name: "告警查询", Path: "list", Router: "alarmList", Icon: "eye", Role: "admin,user"},
		{ParentId: 5, Name: "告警分发", Path: "rule", Router: "alarmRule", Icon: "message", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 5, Name: "规则添加", Path: "rule-add", Router: "alarmRuleAdd", Invisible: true, Highlight: "/alarm", Role: "admin,user"},
		{ParentId: 5, Name: "规则编辑", Path: "rule-edit", Router: "alarmRuleEdit", Invisible: true, Highlight: "/alarm", Role: "admin,user"},
		{ParentId: 5, Name: "屏蔽规则", Path: "mutes", Router: "alarmMutes", Icon: "stop", Role: "admin,user"},
		//拓扑管理
		{ParentId: 6, Name: "拓扑维护", Path: "list", Router: "topologyList", Icon: "environment", Role: "admin,user"},
		{ParentId: 6, Name: "拓扑编辑", Path: "detail", Router: "topologyDetail", Invisible: true, Highlight: "/topology", Role: "admin,user"},
		{ParentId: 6, Name: "拓扑展示", Path: "show", Router: "topologyShow", Invisible: true, Highlight: "/topology", Role: "admin,user"},
		//报表管理
		{ParentId: 7, Name: "流量报表", Path: "traffic", Router: "reportTraffic", Icon: "file-excel", Role: "admin,user"},
		{ParentId: 7, Name: "报表编辑", Path: "edit", Router: "reportTrafficEdit", Invisible: true, Highlight: "/report", Role: "admin,user"},
		{ParentId: 7, Name: "报表添加", Path: "add", Router: "reportTrafficAdd", Invisible: true, Highlight: "/report", Role: "admin,user"},
		//系统管理
		{ParentId: 8, Name: "用户管理", Path: "users", Router: "systemUsers", Icon: "meh", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 8, Name: "组织管理", Path: "groups", Router: "systemGroups", Icon: "smile", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 8, Name: "指标映射", Path: "init", Router: "sysInit", Icon: "interaction", Role: "admin,user"},
		{ParentId: 8, Name: "出口配置", Path: "bandwidth", Router: "systemBandwidth", Icon: "api", Role: "admin,user"},
		{ParentId: 8, Name: "参数配置", Path: "config", Router: "sysConfig", Icon: "api", Role: "admin"},
		{ParentId: 8, Name: "版本信息", Path: "version", Router: "version", Icon: "info", Role: "admin,user"},
	}
	_, err = o.InsertMulti(len(menus), menus)
	if err != nil {
		logs.Error(err)
	}
	return
}

type MenuItem struct {
	Router    string     `json:"router"`
	Path      string     `json:"path"`
	Name      string     `json:"name"`
	Icon      string     `json:"icon"`
	Meta      Meta       `json:"meta"`
	Authority Authority  `json:"authority,omitempty"`
	Children  []MenuItem `json:"children"`
}
type Meta struct {
	Highlight string `json:"highlight"`
	Invisible bool   `json:"invisible"`
	Page      Page   `json:"page"`
}
type Page struct {
	CacheAble bool `json:"cacheAble"`
}

type Authority struct {
	Role       string `json:"role,omitempty"`
	Permission string `json:"permission,omitempty"`
}

func GetRouter(username string) ([]RouterRes, error) {
	m, err := GetManagerByName(username)
	if err != nil {
		return []RouterRes{}, err
	}
	o := orm.NewOrm()
	var menus []Menu
	_, err = o.QueryTable(Menu{}).Filter("is_available", "0").Filter("role__in", m.Role, "admin,user").OrderBy("parent_id", "id").All(&menus)
	if err != nil {
		return []RouterRes{}, err
	}
	// 构建菜单树
	pRouter := buildMenuTree(menus, m.Role)
	tree := make([]RouterRes, 1)
	tree[0].Router = "root"
	tree[0].Children = pRouter
	return tree, nil
}

func buildMenuTree(menus []Menu, role string) []MenuItem {
	var tree []MenuItem
	rootMenus := getRootMenus(menus)
	for _, root := range rootMenus {
		item := MenuItem{
			Router: root.Router,
			Path:   root.Path,
			Name:   root.Name,
			Icon:   root.Icon,
			Meta: Meta{
				Highlight: root.Highlight,
				Invisible: root.Invisible,
				Page: Page{
					CacheAble: root.CacheAble,
				},
			},
			Authority: Authority{
				Role:       role,
				Permission: root.Permission,
			},
			Children: getChildMenus(menus, root.Id, role),
		}
		tree = append(tree, item)
	}
	return tree
}

func getRootMenus(menus []Menu) []Menu {
	var rootMenus []Menu
	for _, menu := range menus {
		if menu.ParentId == 0 {
			rootMenus = append(rootMenus, menu)
		}
	}
	return rootMenus
}

func getChildMenus(menus []Menu, parentId int, role string) []MenuItem {
	var childMenus []MenuItem
	for _, menu := range menus {
		if menu.ParentId == parentId {
			child := MenuItem{
				Router: menu.Router,
				Path:   menu.Path,
				Name:   menu.Name,
				Icon:   menu.Icon,
				Meta: Meta{
					Highlight: menu.Highlight,
					Invisible: menu.Invisible,
					Page: Page{
						CacheAble: menu.CacheAble,
					},
				},
				Authority: Authority{
					Role:       role,
					Permission: menu.Permission,
				},
				Children: getChildMenus(menus, menu.Id, role),
			}
			childMenus = append(childMenus, child)
		}
	}
	return childMenus
}
