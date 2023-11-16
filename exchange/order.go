package exchange

type OrderType int

const (
	Buy OrderType = iota
	Sell
)

type BodyOrder struct {
	ID     uint64    `json:"id"`
	UserID string    `json:"userID" binding:"required"`
	Price  float64   `json:"price" binding:"required"`
	Amount float64   `json:"amount" binding:"required"`
	Type   OrderType `json:"type"`
}

type ResOrder struct {
	Type   OrderType `json:"type"`
	Price  float64   `json:"price"`
	Amount float64   `json:"amount"`
}

type BodyRegister struct {
	Address string `json:"address" binding:"required"`
}

type ResUser struct {
	UserId      string  `json:"userId"`
	Address     string  `json:"address"`
	CexAddress  string  `json:"cexAddress"`
	SolBalance  float64 `json:"solBalance"`
	EcpcBalance float64 `json:"ecpcBalance"`
}
