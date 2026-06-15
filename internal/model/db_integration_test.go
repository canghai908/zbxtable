package model

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	pkglogger "zbxtable/pkg/logger"
)

type configValueProbe struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	ConfigValue string `gorm:"column:config_value;type:text"`
}

func (configValueProbe) TableName() string {
	return "codex_config_value_probe"
}

func initIntegrationLogger(t *testing.T) {
	t.Helper()
	logPath := filepath.Join(t.TempDir(), "integration.log")
	if err := pkglogger.InitLogger(logPath, 4, 1, 0, 1, false); err != nil {
		t.Fatalf("init logger: %v", err)
	}
}

func getIntegrationDBConfig(t *testing.T) (dbtype, dbhost, dbuser, dbpass, dbname, dbport string) {
	t.Helper()
	dbtype = os.Getenv("ZBXTABLE_TEST_DBTYPE")
	dbhost = os.Getenv("ZBXTABLE_TEST_DBHOST")
	dbuser = os.Getenv("ZBXTABLE_TEST_DBUSER")
	dbpass = os.Getenv("ZBXTABLE_TEST_DBPASS")
	dbname = os.Getenv("ZBXTABLE_TEST_DBNAME")
	dbport = os.Getenv("ZBXTABLE_TEST_DBPORT")
	if dbtype == "" || dbhost == "" || dbuser == "" || dbname == "" || dbport == "" {
		t.Skip("integration database env not set")
	}
	return dbtype, dbhost, dbuser, dbpass, dbname, dbport
}

func TestConfigValueTextProbe(t *testing.T) {
	initIntegrationLogger(t)
	dbtype, dbhost, dbuser, dbpass, dbname, dbport := getIntegrationDBConfig(t)

	if err := InitGormDB(dbtype, dbhost, dbuser, dbpass, dbname, dbport, "test"); err != nil {
		t.Fatalf("InitGormDB failed: %v", err)
	}

	_ = DB.Migrator().DropTable(&configValueProbe{})
	t.Cleanup(func() {
		_ = DB.Migrator().DropTable(&configValueProbe{})
		sqlDB, err := DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	if err := DB.AutoMigrate(&configValueProbe{}); err != nil {
		t.Fatalf("probe AutoMigrate failed: %v", err)
	}

	if !DB.Migrator().HasTable(&configValueProbe{}) {
		t.Fatal("probe table was not created")
	}

	var dataType string
	switch dbtype {
	case "mysql":
		if err := DB.Raw(`
			SELECT DATA_TYPE
			FROM information_schema.columns
			WHERE table_schema = DATABASE()
			  AND table_name = ?
			  AND column_name = ?`,
			"codex_config_value_probe", "config_value").Scan(&dataType).Error; err != nil {
			t.Fatalf("query mysql column type failed: %v", err)
		}
	case "postgresql":
		if err := DB.Raw(`
			SELECT data_type
			FROM information_schema.columns
			WHERE table_catalog = current_database()
			  AND table_name = ?
			  AND column_name = ?`,
			"codex_config_value_probe", "config_value").Scan(&dataType).Error; err != nil {
			t.Fatalf("query postgres column type failed: %v", err)
		}
	default:
		t.Fatalf("unsupported db type for probe: %s", dbtype)
	}

	if dataType != "text" {
		t.Fatalf("config_value column type = %q, want text", dataType)
	}
}

func TestDatabaseInitializationIntegration(t *testing.T) {
	initIntegrationLogger(t)
	dbtype, dbhost, dbuser, dbpass, dbname, dbport := getIntegrationDBConfig(t)

	if err := InitGormDB(dbtype, dbhost, dbuser, dbpass, dbname, dbport, "test"); err != nil {
		t.Fatalf("InitGormDB failed: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	if err := AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	if err := DatabaseInit(); err != nil {
		t.Fatalf("DatabaseInit failed: %v", err)
	}

	if !DB.Migrator().HasTable(&Config{}) {
		t.Fatal("config table was not created")
	}

	var adminCount int64
	if err := DB.Model(&User{}).Where("username = ?", "admin").Count(&adminCount).Error; err != nil {
		t.Fatalf("count admin user failed: %v", err)
	}
	if adminCount == 0 {
		t.Fatal("expected admin user to be initialized")
	}

	var encryptionKeyCount int64
	if err := DB.Model(&Config{}).Where("config_key = ?", "encryption_key").Count(&encryptionKeyCount).Error; err != nil {
		t.Fatalf("count encryption_key config failed: %v", err)
	}
	if encryptionKeyCount == 0 {
		t.Fatal("expected encryption_key config to be initialized")
	}

	t.Log(fmt.Sprintf("database initialization succeeded for %s", dbtype))
}
