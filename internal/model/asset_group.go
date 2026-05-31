package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// AssetGroup 设备分组（一级菜单分类），支持内置和用户自定义
type AssetGroup struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;size:50;not null" json:"name"`
	GroupKey  string    `gorm:"column:group_key;size:50;not null;uniqueIndex" json:"group_key"`
	Icon      string    `gorm:"column:icon;size:50;default:''" json:"icon"`
	SortOrder int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	IsBuiltin bool      `gorm:"column:is_builtin;default:false" json:"is_builtin"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (g *AssetGroup) TableName() string {
	return TableName("asset_group")
}

func GetAllAssetGroups() ([]AssetGroup, error) {
	var list []AssetGroup
	err := DB.Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

func GetAssetGroupByID(id int64) (*AssetGroup, error) {
	var g AssetGroup
	err := DB.Where("id = ?", id).First(&g).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func GetAssetGroupByKey(key string) (*AssetGroup, error) {
	var g AssetGroup
	err := DB.Where("group_key = ?", key).First(&g).Error
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func CreateAssetGroup(g *AssetGroup) error {
	return DB.Create(g).Error
}

func UpdateAssetGroup(g *AssetGroup) error {
	return DB.Model(&AssetGroup{}).Where("id = ?", g.ID).Updates(map[string]interface{}{
		"name":       g.Name,
		"icon":       g.Icon,
		"sort_order": g.SortOrder,
		"updated_at": time.Now(),
	}).Error
}

// DeleteAssetGroup 删除分组，内置分组不可删除，有关联类型时也不可删除
func DeleteAssetGroup(id int64) error {
	g, err := GetAssetGroupByID(id)
	if err != nil {
		return err
	}
	if g.IsBuiltin {
		return errors.New("内置分组不可删除")
	}
	var count int64
	DB.Model(&AssetType{}).Where("menu_group = ?", g.GroupKey).Count(&count)
	if count > 0 {
		return errors.New("该分组下还有设备类型，请先修改或删除对应类型")
	}
	return DB.Delete(&AssetGroup{}, id).Error
}

// InitDefaultAssetGroups 幂等预置内置分组
func InitDefaultAssetGroups() error {
	defaults := []AssetGroup{
		{Name: "主机管理", GroupKey: "host", Icon: "hdd", SortOrder: 1, IsBuiltin: true},
		{Name: "网络管理", GroupKey: "net", Icon: "cloud", SortOrder: 2, IsBuiltin: true},
		{Name: "硬件管理", GroupKey: "server", Icon: "database", SortOrder: 3, IsBuiltin: true},
	}
	for _, g := range defaults {
		var existing AssetGroup
		err := DB.Where("group_key = ?", g.GroupKey).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := DB.Create(&g).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
