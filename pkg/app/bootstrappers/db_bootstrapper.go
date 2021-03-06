package bootstrappers

import (
	"fmt"
	"time"

	"github.com/DuC-cnZj/dota2app/pkg/adapter"
	"github.com/DuC-cnZj/dota2app/pkg/contracts"
	"github.com/DuC-cnZj/dota2app/pkg/dlog"
	"github.com/DuC-cnZj/dota2app/pkg/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Models = []interface{}{
	&models.User{},
	&models.Notification{},
	&models.File{},
}

type DBBootstrapper struct{}

func (D *DBBootstrapper) Bootstrap(app contracts.ApplicationInterface) error {
	cfg := app.Config()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", cfg.DBUsername, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBDatabase)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	db.Logger = &adapter.GormLoggerAdapter{}

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Hour)

	if app.IsDebug() {
		db.Logger.LogMode(logger.Info)
	} else {
		db.Logger.LogMode(logger.Error)
	}

	app.RegisterAfterShutdownFunc(func(app contracts.ApplicationInterface) {
		if err := sqlDB.Close(); err != nil {
			dlog.Error(err)
		}

		dlog.Info("db closed.")
	})
	app.DBManager().SetDB(db)

	if err := app.DBManager().AutoMigrate(Models...); err != nil {
		return err
	}

	return nil
}
