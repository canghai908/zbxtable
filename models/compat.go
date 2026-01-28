package models

import (
	"os"

	"github.com/astaxie/beego/orm"
)

// UseGorm 检查是否使用 GORM
func UseGorm() bool {
	return os.Getenv("USE_GORM") == "true"
}

// GetOrm 获取 ORM 实例（兼容层）
// 如果使用 GORM，返回 nil（需要使用 GetDB()）
// 如果使用 beego/orm，返回 orm.Ormer
func GetOrm() orm.Ormer {
	if UseGorm() {
		return nil
	}
	return orm.NewOrm()
}

// 兼容函数：QueryTable
// 如果使用 GORM，需要调用者使用 GetDB() 进行查询
// 如果使用 beego/orm，返回 QuerySeter
func QueryTable(model interface{}) interface{} {
	if UseGorm() {
		// 返回一个标记，调用者需要使用 GetDB() 进行查询
		return "USE_GORM"
	}
	o := orm.NewOrm()
	return o.QueryTable(model)
}
