package mysql

import (
	"MatrixAI-CEX/common"
	"MatrixAI-CEX/config"
	"MatrixAI-CEX/db/mysql/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	dsn := config.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&model.Order{},
		&model.TransactionRecord{},
		&model.AccountAssets{},
		&model.AssetsCNY2User{})
	if err != nil {
		return nil, err
	}

	common.Db = db
	return db, nil
}
