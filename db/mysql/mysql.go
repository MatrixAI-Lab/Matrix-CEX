package mysql

import (
	"MatrixAI-CEX/db/mysql/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	dsn := "root:root1234@tcp(localhost:3306)/exchange_db?charset=utf8mb4&parseTime=True&loc=Local"
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

	return db, nil
}
