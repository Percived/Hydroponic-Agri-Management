package db

import (
	"fmt"
	"log"
	"os"

	"hydroponic-backend/internal/platform/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func NewMySQL(cfg config.MySQLConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Params)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newGormLogger()})
}

func newGormLogger() gormlogger.Interface {
	return gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold: 0,
			LogLevel:      gormlogger.Warn,
			Colorful:      true,
		},
	)
}

func CloseMySQL(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
