package models

import (
	"fmt"
	"net/url"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// DB GORM 数据库实例
	DB *gorm.DB
)

// InitGormDB 初始化 GORM 数据库连接
func InitGormDB(dbtype, dbhost, dbuser, dbpass, dbname, dbport string, runmode string) error {
	var dialector gorm.Dialector
	var err error

	switch dbtype {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FShanghai&timeout=5s&charset=utf8&collation=utf8_general_ci",
			dbuser, dbpass, dbhost, dbport, dbname)
		dialector = mysql.Open(dsn)
	case "postgresql":
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			dbuser, url.QueryEscape(dbpass), dbhost, dbport, dbname)
		dialector = postgres.Open(dsn)
	case "sqlite":
		// SQLite 使用文件路径作为数据库名，dbhost 作为文件路径
		// 如果 dbname 为空，使用默认路径
		dbPath := dbname
		if dbPath == "" {
			dbPath = "./data/zbxtable.db"
		}
		// 确保目录存在
		if err := os.MkdirAll("./data", 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}
		dialector = sqlite.Open(dbPath)
	default:
		// 默认使用 MySQL
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FShanghai&timeout=5s&charset=utf8&collation=utf8_general_ci",
			dbuser, dbpass, dbhost, dbport, dbname)
		dialector = mysql.Open(dsn)
	}

	// 配置日志级别
	logLevel := logger.Silent
	if runmode == "dev" {
		logLevel = logger.Info
	}

	// 打开数据库连接
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().Local()
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 获取底层 sql.DB 设置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err = sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// AutoMigrate 自动迁移数据库表
func AutoMigrate() error {
	// GORM 会自动使用模型的 TableName() 方法（如果存在）
	// 否则使用默认的命名策略（结构体名的复数形式）
	return DB.AutoMigrate(
		&Alarm{},
		&Manager{},
		&Topology{},
		&System{},
		&Report{},
		&Egress{},
		&TaskLog{},
		&Rule{},
		&UserGroup{},
		&EventLog{},
		&Config{},
		&Menu{},
		&ZabbixTenant{},
	//&ZabbixInstance{},
	//&ZabbixTenantBinding{},
	)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
