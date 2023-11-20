package exchange

import (
	"errors"
	"time"
)

type OrderType int

const (
	Buy OrderType = iota
	Sell
)

func (ot OrderType) Validate() error {
	if ot != Buy && ot != Sell {
		return errors.New("invalid order type")
	}
	return nil
}

type BodyOrder struct {
	UserID string    `json:"userID" binding:"required"`
	Price  float64   `json:"price" binding:"required"`
	Amount float64   `json:"amount" binding:"required"`
	Type   OrderType `json:"type"`
}

type BodyUser struct {
	UserID string `json:"userId" binding:"required"`
}

type ResOrder struct {
	CreatedAt time.Time `json:"createdAt"`
	Type      OrderType `json:"type"`
	Price     float64   `json:"price"`
	Amount    float64   `json:"amount"`
	Total     float64   `json:"total"`
}

type BodyRegister struct {
	Address string `json:"address" binding:"required"`
}

type ResUser struct {
	UserId      string  `json:"userId"`
	Address     string  `json:"address"`
	CexAddress  string  `json:"cexAddress"`
	SolBalance  float64 `json:"solBalance"`
	SolTotal    float64 `json:"solTotal"`
	EcpcBalance float64 `json:"ecpcBalance"`
	EcpcTotal   float64 `json:"ecpcTotal"`
}

type ResRecords struct {
	CreatedAt         time.Time `json:"createdAt"`
	TransactionPrice  float64   `json:"transactionPrice"`
	TransactionAmount float64   `json:"transactionAmount"`
}
