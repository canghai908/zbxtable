package model

import (
	"errors"
	"sort"
	"zbxtable/pkg/logger"

	"gorm.io/gorm"
)

type Menu struct {
	Id          int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ParentId    int    `gorm:"column:parent_id;default:0" json:"parent_id"`
	Name        string `gorm:"column:name;size:50" json:"name"`
	Path        string `gorm:"column:path;size:100" json:"path"`
	Router      string `gorm:"column:router;size:100" json:"router"`
	Icon        string `gorm:"column:icon;size:50" json:"icon"`
	Role        string `gorm:"column:role;size:50" json:"role"`
	Permission  string `gorm:"column:permission;size:100" json:"permission"`
	Invisible   bool   `gorm:"column:in_visible;default:false" json:"invisible"`
	IsAvailable bool   `gorm:"column:is_available;default:false" json:"is_available"`
	Highlight   string `gorm:"column:highlight;size:100" json:"highlight"`
	CacheAble   bool   `gorm:"column:cacheAble;default:false" json:"cacheable"`
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
		// 设备管理（一级父菜单，紧随工作台之后，作为设备资产统一入口）
		{ParentId: 0, Name: "设备管理", Path: "assets", Router: "assets", Icon: "appstore", CacheAble: false, Role: "admin,user"},
		// 主机/网络/硬件管理已统一并入"资产管理"，菜单隐藏（保留路由供详情页使用），后续清理页面
		{ParentId: 0, Name: "主机管理", Path: "host", Router: "host", Icon: "hdd", CacheAble: false, Role: "admin,user", Invisible: true},
		{ParentId: 0, Name: "网络管理", Path: "net", Router: "net", Icon: "cloud", CacheAble: false, Role: "admin,user", Invisible: true},
		{ParentId: 0, Name: "硬件管理", Path: "server", Router: "server", Icon: "database", CacheAble: false, Role: "admin,user", Invisible: true},
		{ParentId: 0, Name: "告警管理", Path: "alarm", Router: "alarm", Icon: "alert", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "拓扑管理", Path: "topology", Router: "topology", Icon: "picture", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "报表管理", Path: "report", Router: "report", Icon: "file", CacheAble: false, Role: "admin,user"},
		{ParentId: 0, Name: "系统管理", Path: "system", Router: "system", Icon: "setting", CacheAble: false, Role: "admin,user"},

		// 二级菜单（ParentId 为上面一级菜单的 1-based 序号）
		//工作台 (ParentId: 1)
		{ParentId: 1, Name: "首页", Path: "workplace", Router: "workplace", Icon: "home", Role: "admin,user"},
		{ParentId: 1, Name: "状态总览", Path: "overview", Router: "overview", Icon: "appstore", Role: "admin,user"},
		{ParentId: 1, Name: "数据面板", Path: "dash", Router: "dash", Icon: "block", IsAvailable: true, Role: "admin,user"},
		//设备管理 (ParentId: 2)
		{ParentId: 2, Name: "设备树", Path: "tree", Router: "assetBrowser", Icon: "apartment", Role: "admin,user"},
		// 主机/网络/硬件管理 (ParentId 3/4/5)：列表已并入"资产管理"，仅保留各设备详情路由
		{ParentId: 3, Name: "Linux主机详情", Path: "lindetail", Router: "linDetail", Invisible: true, Highlight: "/host", Role: "admin,user"},
		{ParentId: 3, Name: "Windows主机详情", Path: "windetail", Router: "winDetail", Invisible: true, Highlight: "/host", Role: "admin,user"},
		{ParentId: 3, Name: "通用设备详情", Path: "device-detail", Router: "deviceDetail", Invisible: true, Highlight: "/assets/tree", Role: "admin,user"},
		{ParentId: 4, Name: "网络设备详情", Path: "detail", Router: "netDetail", Invisible: true, Highlight: "/net", Role: "admin,user"},
		{ParentId: 5, Name: "物理设备详情", Path: "detail", Router: "srvDetail", Invisible: true, Highlight: "/server", Role: "admin,user"},
		//告警管理 (ParentId: 6)
		{ParentId: 6, Name: "告警分析", Path: "analysis", Router: "alarmAnalysis", Icon: "hourglass", Role: "admin,user"},
		{ParentId: 6, Name: "告警查询", Path: "list", Router: "alarmList", Icon: "eye", Role: "admin,user"},
		{ParentId: 6, Name: "告警分发", Path: "rule", Router: "alarmRule", Icon: "message", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 6, Name: "屏蔽规则", Path: "mutes", Router: "alarmMutes", Icon: "stop", Role: "admin,user"},
		//拓扑管理 (ParentId: 7)
		{ParentId: 7, Name: "拓扑维护", Path: "list", Router: "topologyList", Icon: "environment", Role: "admin,user"},
		{ParentId: 7, Name: "拓扑编辑", Path: "detail", Router: "topologyDetail", Invisible: true, Highlight: "/topology", Role: "admin,user"},
		{ParentId: 7, Name: "拓扑展示", Path: "show", Router: "topologyShow", Invisible: true, Highlight: "/topology", Role: "admin,user"},
		//报表管理 (ParentId: 8)
		{ParentId: 8, Name: "指标报表", Path: "host", Router: "hostReport", Icon: "file-excel", Role: "admin,user"},
		//系统管理 (ParentId: 9)
		{ParentId: 9, Name: "用户管理", Path: "users", Router: "systemUsers", Icon: "user", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 9, Name: "组织管理", Path: "groups", Router: "systemGroups", Icon: "team", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 9, Name: "菜单管理", Path: "menu", Router: "menuManagement", Icon: "menu", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 9, Name: "Zabbix配置", Path: "zabbix", Router: "zabbix", Icon: "cloud-server", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 9, Name: "设备配置", Path: "asset-management", Router: "assetManagement", Icon: "appstore", Role: "admin", Permission: "['add','edit','delete','update']"},
		{ParentId: 9, Name: "出口配置", Path: "bandwidth", Router: "systemBandwidth", Icon: "swap", Role: "admin,user"},
		{ParentId: 9, Name: "参数配置", Path: "config", Router: "sysConfig", Icon: "control", Role: "admin"},
		{ParentId: 9, Name: "版本信息", Path: "version", Router: "version", Icon: "info-circle", Role: "admin,user"},
		{ParentId: 9, Name: "资产类型", Path: "asset-type", Router: "assetTypeManagement", Icon: "tags", Role: "admin", Permission: "['add','edit','delete','update']", Invisible: true, Highlight: "/system/asset-management"},
		{ParentId: 9, Name: "设备绑定", Path: "asset-binding", Router: "assetBinding", Icon: "link", Role: "admin", Permission: "['add','edit','delete','update']", Invisible: true, Highlight: "/system/asset-management"},
	}
}

// InitMenuData 菜单初始化
func InitMenuData() error {
	var count int64
	err := DB.Model(&Menu{}).Count(&count).Error
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	if count > 0 {
		return nil
	}
	menus := getMenuDefinitions()
	err = DB.Create(&menus).Error
	if err != nil {
		logger.Log.Error(err)
		return err
	}
	return nil
}

// CheckAndAddMenus 检查并添加缺失的菜单项（用于版本升级）
// 检查 getMenuDefinitions 中定义的所有菜单是否在数据库中存在，如果不存在则添加
func CheckAndAddMenus() error {
	cleanupDeprecatedMenus()
	migrateAssetManagementMenus()
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
					logger.Log.Error("添加一级菜单失败:", menuDef.Name, err)
					continue
				}
				parentMenuMap[idx+1] = menuToInsert.Id // 索引从1开始，对应ParentId
				logger.Log.Info("成功添加一级菜单:", menuDef.Name)
			} else if err != nil {
				logger.Log.Error("查询一级菜单失败:", menuDef.Name, err)
				continue
			} else {
				parentMenuMap[idx+1] = existingMenu.Id
				logger.Log.Debug("一级菜单已存在:", menuDef.Name)
			}
		}
	}

	// 处理二级菜单（ParentId > 0）
	for _, menuDef := range menuDefs {
		if menuDef.ParentId > 0 {
			// 获取父菜单的实际ID
			parentId, exists := parentMenuMap[menuDef.ParentId]
			if !exists {
				logger.Log.Warning("找不到父菜单，跳过菜单:", menuDef.Name, "ParentId:", menuDef.ParentId)
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
					logger.Log.Error("添加菜单失败:", menuDef.Name, err)
					continue
				}
				logger.Log.Info("成功添加菜单:", menuDef.Name)
			} else if err != nil {
				logger.Log.Error("查询菜单失败:", menuDef.Name, err)
				continue
			} else {
				logger.Log.Debug("菜单已存在:", menuDef.Name)
			}
		}
	}

	logger.Log.Info("菜单检查完成")
	return nil
}

func cleanupDeprecatedMenus() {
	if err := DB.Where("router = ? OR path = ?", "systemChpwd", "chpwd").Delete(&Menu{}).Error; err != nil {
		logger.Log.Error("清理废弃菜单失败:", err)
	}
	if err := DB.Where("router = ? OR path = ?", "metricMapping", "mapping").Delete(&Menu{}).Error; err != nil {
		logger.Log.Error("清理指标映射菜单失败:", err)
	}
	// 工作台下的"资产管理"(inventory) 已合并进顶级"资产管理"(assetBrowser)，删除存量菜单
	if err := DB.Where("router = ?", "inventory").Delete(&Menu{}).Error; err != nil {
		logger.Log.Error("清理资产管理(inventory)菜单失败:", err)
	}
	// 主机/网络/硬件管理已并入"资产管理"，将存量父菜单隐藏（保留记录与路由，供详情页使用）
	if err := DB.Model(&Menu{}).
		Where("parent_id = 0 AND router IN ?", []string{"host", "net", "server"}).
		Update("in_visible", true).Error; err != nil {
		logger.Log.Error("隐藏主机/网络/硬件管理菜单失败:", err)
	}
	// 删除已废弃的设备列表子菜单（页面已移除，列表统一由"资产管理"提供）
	if err := DB.Where("router IN ?", []string{"linux", "windows", "netList", "srvList", "sanList", "stoList"}).
		Delete(&Menu{}).Error; err != nil {
		logger.Log.Error("清理设备列表子菜单失败:", err)
	}
	// 顶级"资产管理"由直接指向页面(assetBrowser)改造为父容器(assets)，
	// 其下新增"资产树"子菜单承载原页面（由 CheckAndAddMenus 自动补建）
	if err := DB.Model(&Menu{}).
		Where("parent_id = 0 AND name = ? AND router = ?", "资产管理", "assetBrowser").
		Update("router", "assets").Error; err != nil {
		logger.Log.Error("迁移资产管理为父菜单失败:", err)
	}
	// 系统管理下的"资产管理"/"资产设置"统一改名为"设备配置"（幂等）。
	// 历史上该菜单经历 资产管理 → 资产设置 → 设备配置 多次更名，两条 WHERE 均需覆盖。
	if err := DB.Model(&Menu{}).
		Where("router = ? AND name IN ?", "assetManagement", []string{"资产管理", "资产设置"}).
		Update("name", "设备配置").Error; err != nil {
		logger.Log.Error("重命名资产设置菜单失败:", err)
	}
	// 顶级"资产管理"(设备入口)改名为"设备管理"（幂等）。CheckAndAddMenus 按 name 匹配，
	// 必须先在此重命名存量行，否则会重复创建一条"设备管理"菜单。
	if err := DB.Model(&Menu{}).
		Where("parent_id = 0 AND router = ? AND name = ?", "assets", "资产管理").
		Update("name", "设备管理").Error; err != nil {
		logger.Log.Error("重命名顶级资产管理菜单失败:", err)
	}
	// "资产树"子菜单改名为"设备树"（幂等）
	if err := DB.Model(&Menu{}).
		Where("router = ? AND name = ?", "assetBrowser", "资产树").
		Update("name", "设备树").Error; err != nil {
		logger.Log.Error("重命名资产树菜单失败:", err)
	}
	// "资产绑定"菜单改名为"设备绑定"（幂等）
	if err := DB.Model(&Menu{}).
		Where("router = ? AND name = ?", "assetBinding", "资产绑定").
		Update("name", "设备绑定").Error; err != nil {
		logger.Log.Error("重命名资产绑定菜单失败:", err)
	}
}

func migrateAssetManagementMenus() {
	updates := []struct {
		router string
		values map[string]interface{}
	}{
		{
			router: "assetTypeManagement",
			values: map[string]interface{}{
				"in_visible": true,
				"highlight":  "/system/asset-management",
			},
		},
		{
			router: "assetBinding",
			values: map[string]interface{}{
				"in_visible": true,
				"highlight":  "/system/asset-management",
			},
		},
	}

	for _, update := range updates {
		if err := DB.Model(&Menu{}).Where("router = ?", update.router).Updates(update.values).Error; err != nil {
			logger.Log.Error("迁移资产管理菜单失败:", update.router, err)
		}
	}
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
	TypeCode  string `json:"type_code,omitempty"` // 动态设备类型标识，供通用列表组件读取
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
	err = DB.Where("is_available = ? AND (role LIKE ? OR role LIKE ?)", false, "%"+m.Role+"%", "%admin,user%").Order("parent_id, id").Find(&menus).Error
	if err != nil {
		return []RouterRes{}, err
	}
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
	sortMenusByDefinition(rootMenus)
	return rootMenus
}

func getChildMenus(menus []Menu, parentId int, role string) []MenuItem {
	var childDefs []Menu
	for _, menu := range menus {
		if menu.ParentId == parentId {
			childDefs = append(childDefs, menu)
		}
	}
	sortMenusByDefinition(childDefs)

	var childMenus []MenuItem
	for _, menu := range childDefs {
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
	return childMenus
}

// ============ 菜单管理 CRUD 方法 ============

// GetAllMenus 获取所有菜单（用于管理界面）
func GetAllMenus() ([]Menu, error) {
	var menus []Menu
	err := DB.Order("parent_id, id").Find(&menus).Error
	sortMenusByParentAndDefinition(menus)
	return menus, err
}

// GetMenuByID 根据ID获取菜单
func GetMenuByID(id int) (*Menu, error) {
	var menu Menu
	err := DB.Where("id = ?", id).First(&menu).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// CreateMenu 创建菜单
func CreateMenu(menu *Menu) error {
	return DB.Create(menu).Error
}

// UpdateMenu 更新菜单
func UpdateMenu(menu *Menu) error {
	return DB.Model(&Menu{}).Where("id = ?", menu.Id).Updates(menu).Error
}

// DeleteMenu 删除菜单（级联删除子菜单）
func DeleteMenu(id int) error {
	// 先删除所有子菜单
	err := DB.Where("parent_id = ?", id).Delete(&Menu{}).Error
	if err != nil {
		return err
	}
	// 再删除菜单本身
	return DB.Where("id = ?", id).Delete(&Menu{}).Error
}

// GetParentMenus 获取所有父菜单（一级菜单）
func GetParentMenus() ([]Menu, error) {
	var menus []Menu
	err := DB.Where("parent_id = ?", 0).Order("id").Find(&menus).Error
	sortMenusByDefinition(menus)
	return menus, err
}

// GetMenusByParentID 根据父菜单ID获取子菜单
func GetMenusByParentID(parentId int) ([]Menu, error) {
	var menus []Menu
	err := DB.Where("parent_id = ?", parentId).Order("id").Find(&menus).Error
	sortMenusByDefinition(menus)
	return menus, err
}

func sortMenusByParentAndDefinition(menus []Menu) {
	sort.SliceStable(menus, func(i, j int) bool {
		if menus[i].ParentId != menus[j].ParentId {
			return menus[i].ParentId < menus[j].ParentId
		}
		return compareMenuOrder(menus[i], menus[j])
	})
}

func sortMenusByDefinition(menus []Menu) {
	sort.SliceStable(menus, func(i, j int) bool {
		return compareMenuOrder(menus[i], menus[j])
	})
}

func compareMenuOrder(a, b Menu) bool {
	aOrder := menuDefinitionOrder(a)
	bOrder := menuDefinitionOrder(b)
	if aOrder != bOrder {
		return aOrder < bOrder
	}
	return a.Id < b.Id
}

func menuDefinitionOrder(menu Menu) int {
	defs := getMenuDefinitions()
	// 精确匹配（含 parent_id）—— 适用于全新安装，DB 父菜单 ID 与数组索引一致
	for idx, def := range defs {
		if def.ParentId == menu.ParentId && def.Router == menu.Router && def.Path == menu.Path && def.Name == menu.Name {
			return idx
		}
	}
	// 兜底：仅按 router 匹配 —— 适用于已有安装（经多次迁移后 parent_id 与数组索引不再对应）
	// router 在整个菜单体系中全局唯一，可安全用作排序依据
	if menu.Router != "" {
		for idx, def := range defs {
			if def.Router == menu.Router {
				return idx
			}
		}
	}
	return 1 << 30
}
