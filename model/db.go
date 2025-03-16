package model

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	// 引入logrus
	"github.com/sirupsen/logrus"
)

var DB *gorm.DB

func InitDB(dsn string) error {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt: true,                                // 启用预编译提升性能[6](@ref)
		Logger:      logger.Default.LogMode(logger.Warn), // 生产环境关闭日志
	})

	if err != nil {
		// 使用logrus记录错误日志
		logrus.Error("数据库连接失败:", err)
		return err
	}

	// 配置连接池
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	DB = db
	// 使用logrus记录信息日志
	logrus.Info("数据库初始化成功")
	return nil
}
