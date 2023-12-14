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

	err = db.AutoMigrate(&model.Order{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&model.TransactionRecord{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&model.AccountAssets{})
	if err != nil {
		return nil, err
	}

	common.Db = db
	return db, nil
}
