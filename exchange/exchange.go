package exchange

import (
	"MatrixAI-CEX/chain/conn"
	"MatrixAI-CEX/db/mysql/model"
	"MatrixAI-CEX/utils"

	"context"
	"fmt"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
	"github.com/reactivex/rxgo/v2"
	"gorm.io/gorm"
)

type Exchange struct {
	MysqlDB *gorm.DB
	Conn    *conn.Conn
}

func (e *Exchange) AddOrder(bodyOrder BodyOrder) (string, error) {
	var accountAssets model.AccountAssets
	result := e.MysqlDB.Where("user_id = ?", bodyOrder.UserID).First(&accountAssets)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("userId does not exist")
		}
		return "", result.Error
	}

	// Validate the Type value
	if err := bodyOrder.Type.Validate(); err != nil {
		return "", err
	}

	side := model.Sell
	if bodyOrder.Type == Buy {
		side = model.Buy
		if accountAssets.SolBalance < bodyOrder.Amount {
			return "", fmt.Errorf("insufficient balance: Sol")
		}
	} else {
		if accountAssets.EcpcBalance < bodyOrder.Amount {
			return "", fmt.Errorf("insufficient balance: Ecpc")
		}
	}

	order := model.Order{
		CreatedAt: time.Now(),
		OrderId:   fmt.Sprintf("%d%s", time.Now().Unix(), bodyOrder.UserID),
		UserId:    bodyOrder.UserID,
		OrderSide: side,
		Price:     bodyOrder.Price,
		Total:     bodyOrder.Amount,
		Quantity:  bodyOrder.Amount,
	}

	e.MysqlDB.Create(&order)
	e.matchOrders()

	return order.OrderId, nil
}

func (e *Exchange) DeleteOrder(orderID string) error {
	var order model.Order
	result := e.MysqlDB.Where("order_id = ?", orderID).Delete(&order)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (e *Exchange) GetAndProcessOrders() ([]ResOrder, []ResOrder, error) {
	var buyOrders, sellOrders []model.Order

	if err := e.MysqlDB.Where("order_side = ?", model.Buy).Order("price desc").Find(&buyOrders).Error; err != nil {
		return nil, nil, err
	}
	if err := e.MysqlDB.Where("order_side = ?", model.Sell).Order("price asc").Find(&sellOrders).Error; err != nil {
		return nil, nil, err
	}

	buyOrderObs := rxgo.Just(buyOrders)().Map(func(_ context.Context, i interface{}) (interface{}, error) {
		oldOrder := i.(model.Order)
		return ResOrder{Type: Buy, Price: oldOrder.Price, Amount: oldOrder.Quantity}, nil
	}).Scan(func(_ context.Context, a, b interface{}) (interface{}, error) {
		var orderA ResOrder
		if a != nil {
			orderA = a.(ResOrder)
		}
		orderB := b.(ResOrder)
		if orderA.Price == orderB.Price {
			orderA.Amount += orderB.Amount
			return orderA, nil
		}
		return orderB, nil
	})

	sellOrderObs := rxgo.Just(sellOrders)().Map(func(_ context.Context, i interface{}) (interface{}, error) {
		oldOrder := i.(model.Order)
		return ResOrder{Type: Sell, Price: oldOrder.Price, Amount: oldOrder.Quantity}, nil
	}).Scan(func(_ context.Context, a, b interface{}) (interface{}, error) {
		var orderA ResOrder
		if a != nil {
			orderA = a.(ResOrder)
		}
		orderB := b.(ResOrder)
		if orderA.Price == orderB.Price {
			orderA.Amount += orderB.Amount
			return orderA, nil
		}
		return orderB, nil
	})

	buyOrderList, err := clearDuplicateData(buyOrderObs)
	if err != nil {
		return nil, nil, err
	}

	sellOrderList, err := clearDuplicateData(sellOrderObs)
	if err != nil {
		return nil, nil, err
	}

	return buyOrderList, sellOrderList, nil
}

func (e *Exchange) GetUserOrders(user BodyUser) ([]ResOrder, error) {
	var userOrders []model.Order

	if err := e.MysqlDB.Where("user_id = ?", user.UserID).Find(&userOrders).Error; err != nil {
		return nil, err
	}

	userOrderObs := rxgo.Just(userOrders)().Map(func(_ context.Context, i interface{}) (interface{}, error) {
		oldOrder := i.(model.Order)
		orderType := Buy
		if oldOrder.OrderSide == model.Sell {
			orderType = Sell
		}
		return ResOrder{CreatedAt: oldOrder.CreatedAt, Type: orderType, Price: oldOrder.Price, Amount: oldOrder.Quantity, Total: oldOrder.Total}, nil
	})

	orderList := make([]ResOrder, 0)
	for item := range userOrderObs.Observe() {
		if item.Error() {
			return nil, item.E
		}
		orderList = append(orderList, item.V.(ResOrder))
	}

	return orderList, nil
}

func (e *Exchange) RegisterUser(register BodyRegister) (string, error) {

	_, err := solana.PublicKeyFromBase58(register.Address)
	if err != nil {
		return "", err
	}

	var accountAssets model.AccountAssets

	var count int64
	result := e.MysqlDB.Model(&model.AccountAssets{}).Where("address = ?", register.Address).Count(&count)
	if result.Error != nil {
		return "", result.Error
	}
	if count > 0 {
		e.MysqlDB.Where("address = ?", register.Address).First(&accountAssets)
		return accountAssets.UserId, nil
	}

	solAccount := solana.NewWallet()
	accountAssets = model.AccountAssets{
		UserId:        uuid.New().String(),
		Address:       register.Address,
		CexAddress:    solAccount.PublicKey().String(),
		CexPrivateKey: solAccount.PrivateKey.String(),
	}

	result = e.MysqlDB.Create(&accountAssets)
	if result.Error != nil {
		return "", result.Error
	}

	return accountAssets.UserId, nil
}

func (e *Exchange) GetUserAccountAssets(userId string) (ResUser, error) {
	var resUser ResUser

	var accountAssets model.AccountAssets
	result := e.MysqlDB.Where("user_id = ?", userId).First(&accountAssets)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return resUser, fmt.Errorf("userId does not exist")
		}
		return resUser, result.Error
	}

	_, amount, err := e.Conn.RechargeSol(accountAssets)
	if err != nil {
		return resUser, err
	}

	if amount > 0 {
		accountAssets.SolBalance += float64(amount) / float64(solana.LAMPORTS_PER_SOL)
		accountAssets.SolTotal += float64(amount) / float64(solana.LAMPORTS_PER_SOL)
	}

	e.MysqlDB.Save(&accountAssets)

	resUserObservable := rxgo.Just(accountAssets)().Map(func(_ context.Context, i interface{}) (interface{}, error) {
		accountAssets := i.(model.AccountAssets)
		return ResUser{
			UserId:      accountAssets.UserId,
			Address:     accountAssets.Address,
			CexAddress:  accountAssets.CexAddress,
			SolBalance:  accountAssets.SolBalance,
			SolTotal:    accountAssets.SolTotal,
			EcpcBalance: accountAssets.EcpcBalance,
			EcpcTotal:   accountAssets.EcpcTotal,
		}, nil
	})
	resUserResult := <-resUserObservable.Observe()
	if resUserResult.Error() {
		return resUser, resUserResult.E
	}
	resUser = resUserResult.V.(ResUser)

	return resUser, nil
}

func (e *Exchange) GetTransactionRecords() ([]ResRecords, error) {
	var records []model.TransactionRecord
	e.MysqlDB.Order("created_at desc").Find(&records)

	recordsObs := rxgo.Just(records)().Map(func(_ context.Context, i interface{}) (interface{}, error) {
		oldRecords := i.(model.TransactionRecord)
		return ResRecords{
			CreatedAt:         oldRecords.CreatedAt,
			TransactionPrice:  oldRecords.TransactionPrice,
			TransactionAmount: oldRecords.TransactionAmount,
		}, nil
	})

	recordsList := make([]ResRecords, 0)
	for item := range recordsObs.Observe() {
		if item.Error() {
			return nil, item.E
		}
		recordsList = append(recordsList, item.V.(ResRecords))
	}

	return recordsList, nil
}

func clearDuplicateData(o rxgo.Observable) ([]ResOrder, error) {
	orderList := make([]ResOrder, 0)
	for item := range o.Observe() {
		if item.Error() {
			return nil, item.E
		}
		orderList = append(orderList, item.V.(ResOrder))
	}

	orderMap := make(map[float64]ResOrder, len(orderList))
	for _, order := range orderList {
		if existingOrder, found := orderMap[order.Price]; found && order.Amount <= existingOrder.Amount {
			continue
		}
		orderMap[order.Price] = order
	}

	result := make([]ResOrder, 0, len(orderMap))
	for _, order := range orderMap {
		result = append(result, order)
	}
	return result, nil
}

func (e *Exchange) matchOrders() {
	var buyOrder model.Order
	var sellOrder model.Order
	e.MysqlDB.Where("order_side = ?", model.Buy).Order("price desc").First(&buyOrder)
	e.MysqlDB.Where("order_side = ?", model.Sell).Order("price asc").First(&sellOrder)

	for buyOrder.ID != 0 && sellOrder.ID != 0 {
		if buyOrder.Price >= sellOrder.Price {
			tradeQuantity := utils.Min(buyOrder.Quantity, sellOrder.Quantity)

			buyOrder.Quantity -= tradeQuantity
			sellOrder.Quantity -= tradeQuantity

			e.MysqlDB.Save(&buyOrder)
			e.MysqlDB.Save(&sellOrder)

			var buyAccount, sellAccount model.AccountAssets
			e.MysqlDB.Where("user_id = ?", buyOrder.UserId).First(&buyAccount)
			e.MysqlDB.Where("user_id = ?", sellOrder.UserId).First(&sellAccount)

			buyAccount.EcpcBalance += tradeQuantity
			buyAccount.SolBalance -= tradeQuantity
			sellAccount.EcpcBalance -= tradeQuantity
			sellAccount.SolBalance += tradeQuantity

			e.MysqlDB.Save(&buyAccount)
			e.MysqlDB.Save(&sellAccount)

			transactionRecord := model.TransactionRecord{
				BuyOrderId:        buyOrder.OrderId,
				BuyerId:           buyOrder.UserId,
				SellOrderId:       sellOrder.OrderId,
				SellerId:          sellOrder.UserId,
				TransactionAmount: tradeQuantity,
				TransactionPrice:  buyOrder.Price,
			}
			e.MysqlDB.Create(&transactionRecord)

			if buyOrder.Quantity == 0 {
				e.MysqlDB.Delete(&buyOrder)
				buyOrder = model.Order{}
				e.MysqlDB.Where("order_side = ?", model.Buy).Order("price desc").First(&buyOrder)
			}
			if sellOrder.Quantity == 0 {
				e.MysqlDB.Delete(&sellOrder)
				sellOrder = model.Order{}
				e.MysqlDB.Where("order_side = ?", model.Sell).Order("price asc").First(&sellOrder)
			}
		} else {
			break
		}
	}
}
