package models

import (
	"errors"
	"zbxtable/pkg/utils"

	"gorm.io/gorm"
)

type Menu struct {
	Id          int    `gorm:"column:id;primaryKey;autoIncrement"`
	ParentId    int    `gorm:"column:parent_id;default:0"`
	Name        string `gorm:"column:name;size:50"`
	Path        string `gorm:"column:path;size:100"`
	Router      string `gorm:"column:router;size:100"`
	Icon        string `gorm:"column:icon;size:50"`
	Role        string `gorm:"column:role;size:50"`
	Permission  string `gorm:"column:permission;size:100"`
	Invisible   bool   `gorm:"column:in_visible;default:false"`
	IsAvailable bool   `gorm:"column:is_available;default:false"`
	Highlight   string `gorm:"column:highlight;size:100"`
	CacheAble   bool   `gorm:"column:cacheAble;default:false"`
}

func (t *Menu) TableName() string {
	return TableName("menu")
}

// getMenuDefinitions 获取所有菜单定义（用于初始化和检查）
// 返回菜单列表，ParentId 为 0 表示一级菜单，> 0 表示父菜单在列表中的索引（从1开始）
func getMenuDefinitions() []Menu {
	return []Menu{
		// 一级菜单（ParentId: 0）
		{ParentId: 0, Name: "工作台", Path: "dashboard", Router: "dashboard", Icon: "dashboard", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "主机管理", Path: "host", Router: "host", Icon: "hdd", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "网络管理", Path: "net", Router: "net", Icon: "cloud", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "硬件管理", Path: "server", Router: "server", Icon: "database", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "告警管理", Path: "alarm", Router: "alarm", Icon: "alert", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "拓扑管理", Path: "topology", Router: "topology", Icon: "picture", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "报表管理", Path: "report", Router: "report", Icon: "file", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "系统管理", Path: "system", Router: "system", Icon: "setting", CacheAble: false, Role: "admin,user"},

		// 二级菜单（ParentId: 1-8 对应上面一级菜单的索引）
		//dashboard (ParentId: 1 对应"工作台")
		{ParentId: 1, Name: "首页", Path: "workplace", Router: "workplace", Icon: "home", Role: "admin,user"},
		{ParentId: 1, Name: "资产管理", Path: "inventory", Router: "inventory", Icon: "calendar", Role: "admin,user"},
		{ParentId: 1, Name: "状态总览", Path: "overview", Router: "overview", Icon: "appstore", Role: "admin,user"},
		{ParentId: 1, Name: "数据面板", Path: "dash", Router: "dash", Icon: "block", IsAvailable: true, Role: "admin,user"},
		//主机管理 (ParentId: 2 对应"主机管理")
		{ParentId: 2, Name: "Linux主机", Path: "linux", Router: "linux", Icon: "container", Role: "admin,user"},
		{ParentId: 2, Name: "Linux主机详情", Path: "lindetail", Router: "linDetail", Invisible: true, Highlight: "/host", Role: "admin,user"},
		{ParentId: 2, Name: "Windows主机", Path: "windows", Router: "windows", Icon: "windows", Role: "admin,user"},
		{ParentId: 2, Name: "Windows主机详情", Path: "windetail", Router: "winDetail", Invisible: true, Highlight: "/host", Role: "admin,user"},
		//网络管理 (ParentId: 3 对应"网络管理")
		{ParentId: 3, Name: "网络设备", Path: "list", Router: "netList", Icon: "chrome", Role: "admin,user"},
		{ParentId: 3, Name: "设备详情", Path: "detail", Router: "netDetail", Invisible: true, Highlight: "/net", Role: "admin,user"},
		//硬件管理 (ParentId: 4 对应"硬件管理")
		{ParentId: 4, Name: "物理服务器", Path: "list", Router: "srvList", Icon: "mobile", Role: "admin,user"},
		{ParentId: 4, Name: "光纤交换机", Path: "fiber", Router: "sanList", Icon: "mobile", Role: "admin,user"},
		{ParentId: 4, Name: "存储设备", Path: "storage", Router: "stoList", Icon: "mobile", Role: "admin,user"},
		{ParentId: 4, Name: "设备详情", Path: "detail", Router: "srvDetail", Invisible: true, Highlight: "/server", Role: "admin,user"},
		//告警管理 (ParentId: 5 对应"告警管理")
		{ParentId: 5, Name: "告警分析", Path: "analysis", Router: "alarmAnalysis", Icon: "hourglass", Role: "admin,user"},
		{ParentId: 5, Name: "告警查询", Path: "list", Router: "alarmList", Icon: "eye", Role: "admin,user"},
		{ParentId: 5, Name: "告警分发", Path: "rule", Router: "alarmRule", Icon: "message", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 5, Name: "规则添加", Path: "rule-add", Router: "alarmRuleAdd", Invisible: true, Highlight: "/alarm", Role: "admin,user"},
		{ParentId: 5, Name: "规则编辑", Path: "rule-edit", Router: "alarmRuleEdit", Invisible: true, Highlight: "/alarm", Role: "admin,user"},
		{ParentId: 5, Name: "屏蔽规则", Path: "mutes", Router: "alarmMutes", Icon: "stop", Role: "admin,user"},
		//拓扑管理 (ParentId: 6 对应"拓扑管理")
		{ParentId: 6, Name: "拓扑维护", Path: "list", Router: "topologyList", Icon: "environment", Role: "admin,user"},
		{ParentId: 6, Name: "拓扑编辑", Path: "detail", Router: "topologyDetail", Invisible: true, Highlight: "/topology", Role: "admin,user"},
		{ParentId: 6, Name: "拓扑展示", Path: "show", Router: "topologyShow", Invisible: true, Highlight: "/topology", Role: "admin,user"},
		//报表管理 (ParentId: 7 对应"报表管理")
		{ParentId: 7, Name: "主机报表", Path: "hosts", Router: "hostReport", Icon: "file-excel", Role: "admin,user"},
		{ParentId: 7, Name: "流量报表", Path: "traffic", Router: "reportTraffic", Icon: "file-excel", Role: "admin,user"},
		//系统管理 (ParentId: 8 对应"系统管理")
		{ParentId: 8, Name: "用户管理", Path: "users", Router: "systemUsers", Icon: "meh", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 8, Name: "组织管理", Path: "groups", Router: "systemGroups", Icon: "smile", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 8, Name: "Zabbix配置", Path: "zabbix", Router: "zabbixConfig", Icon: "smile", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 8, Name: "租户配置", Path: "tenant", Router: "zabbixTenant", Icon: "smile", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 8, Name: "指标映射", Path: "init", Router: "sysInit", Icon: "interaction", Role: "admin,user"},
		{ParentId: 8, Name: "出口配置", Path: "bandwidth", Router: "systemBandwidth", Icon: "api", Role: "admin,user"},
		{ParentId: 8, Name: "参数配置", Path: "config", Router: "sysConfig", Icon: "api", Role: "admin"},
		{ParentId: 8, Name: "版本信息", Path: "version", Router: "version", Icon: "info", Role: "admin,user"},
	}
}

// InitMenuData 菜单初始化
func InitMenuData() {
	var count int64
	err := DB.Model(&Menu{}).Count(&count).Error
	if err != nil {
		utils.Log.Error(err)
		return
	}
	if count > 0 {
		return
	}
	menus := getMenuDefinitions()
	err = DB.Create(&menus).Error
	if err != nil {
		utils.Log.Error(err)
	}
}

// CheckAndAddMenus 检查并添加缺失的菜单项（用于版本升级）
// 检查 getMenuDefinitions 中定义的所有菜单是否在数据库中存在，如果不存在则添加
func CheckAndAddMenus() {
	menuDefs := getMenuDefinitions()

	// 建立一级菜单名称到数据库ID的映射
	parentMenuMap := make(map[int]int) // key: 菜单定义中的索引(从1开始), value: 数据库中的ID

	// 先处理一级菜单（ParentId: 0）
	for idx, menuDef := range menuDefs {
		if menuDef.ParentId == 0 {
			// 检查一级菜单是否存在
			var existingMenu Menu
			err := DB.Where("name = ? AND parent_id = ?", menuDef.Name, 0).First(&existingMenu).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 一级菜单不存在，添加它
				menuToInsert := menuDef
				err = DB.Create(&menuToInsert).Error
				if err != nil {
					utils.Log.Error("添加一级菜单失败:", menuDef.Name, err)
					continue
				}
				parentMenuMap[idx+1] = menuToInsert.Id // 索引从1开始，对应ParentId
				utils.Log.Info("成功添加一级菜单:", menuDef.Name)
			} else if err != nil {
				utils.Log.Error("查询一级菜单失败:", menuDef.Name, err)
				continue
			} else {
				parentMenuMap[idx+1] = existingMenu.Id
				utils.Log.Debug("一级菜单已存在:", menuDef.Name)
			}
		}
	}

	// 处理二级菜单（ParentId > 0）
	for _, menuDef := range menuDefs {
		if menuDef.ParentId > 0 {
			// 获取父菜单的实际ID
			parentId, exists := parentMenuMap[menuDef.ParentId]
			if !exists {
				utils.Log.Warning("找不到父菜单，跳过菜单:", menuDef.Name, "ParentId:", menuDef.ParentId)
				continue
			}

			// 检查菜单是否存在
			var existingMenu Menu
			err := DB.Where("name = ? AND parent_id = ?", menuDef.Name, parentId).First(&existingMenu).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 菜单不存在，添加它
				menuToInsert := menuDef
				menuToInsert.ParentId = parentId // 使用数据库中的实际父菜单ID
				err = DB.Create(&menuToInsert).Error
				if err != nil {
					utils.Log.Error("添加菜单失败:", menuDef.Name, err)
					continue
				}
				utils.Log.Info("成功添加菜单:", menuDef.Name)
			} else if err != nil {
				utils.Log.Error("查询菜单失败:", menuDef.Name, err)
				continue
			} else {
				utils.Log.Debug("菜单已存在:", menuDef.Name)
			}
		}
	}

	utils.Log.Info("菜单检查完成")
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
	var menus []Menu
	// GORM 查询：is_available = false (0) 且 role 包含 m.Role 或 "admin,user"
	err = DB.Where("is_available = ? AND (role LIKE ? OR role LIKE ?)", false, "%"+m.Role+"%", "%admin,user%").Order("parent_id, id").Find(&menus).Error
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
