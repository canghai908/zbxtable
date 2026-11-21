package model

import "testing"

func TestOnCallMenuReadRoleMigrationPostgres(t *testing.T) {
	db := onCallScenarioDB(t)
	if err := db.AutoMigrate(&Menu{}); err != nil {
		t.Fatal(err)
	}
	parent := Menu{Name: "告警管理", Router: "alarm", Role: "admin,user"}
	if err := db.Create(&parent).Error; err != nil {
		t.Fatal(err)
	}
	old := Menu{ParentId: parent.Id, Name: "值班排班", Router: "alarmOnCall", Role: "admin"}
	if err := db.Create(&old).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := CheckAndAddMenus(); err != nil {
			t.Fatal(err)
		}
		var saved Menu
		if err := db.First(&saved, old.Id).Error; err != nil {
			t.Fatal(err)
		}
		if saved.Role != "admin,user" {
			t.Fatalf("old menu role %q", saved.Role)
		}
	}
	var count int64
	if err := db.Model(&Menu{}).Where("router = ?", "alarmOnCall").Count(&count).Error; err != nil || count != 1 {
		t.Fatalf("migration duplicated schedule menu: %d %v", count, err)
	}
}
