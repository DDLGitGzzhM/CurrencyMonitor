package database

import (
	"CurrencyMonitor/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	var err error

	// 使用SQLite数据库
	DB, err = gorm.Open(sqlite.Open("currency_monitor.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	// 自动迁移数据库表结构
	err = DB.AutoMigrate(&models.LongShortRatio{}, &models.APILog{})
	if err != nil {
		return err
	}

	log.Println("数据库初始化完成")
	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
