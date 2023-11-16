package model

import (
	"time"

	"gorm.io/gorm"
)

type OrderSide string

const (
	Buy  OrderSide = "Buy"
	Sell OrderSide = "Sell"
)

// 订单表
type Order struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	OrderId   string
	UserId    string
	OrderSide OrderSide
	AssetType string `gorm:"default:SOL-ECPC"`
	Price     float64
	Quantity  float64
}

// 交易记录表
type TransactionRecord struct {
	gorm.Model
	BuyOrderId        string
	BuyerId           string
	SellOrderId       string
	SellerId          string
	AssetType         string `gorm:"default:SOL-ECPC"`
	TransactionAmount float64
	TransactionPrice  float64
}

// 用户资产表
type AccountAssets struct {
	gorm.Model
	UserId        string
	Address       string
	CexAddress    string
	CexPrivateKey string
	SolBalance    float64 `gorm:"default:0"`
	EcpcBalance   float64 `gorm:"default:0"`
}
