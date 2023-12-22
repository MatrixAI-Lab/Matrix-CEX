package exchange

import (
	"MatrixAI-CEX/chain/conn"
	"MatrixAI-CEX/db/mysql/model"
	"MatrixAI-CEX/utils"
	logs "MatrixAI-CEX/utils/log_utils"

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
		if accountAssets.SolBalance < bodyOrder.SolAmount {
			return "", fmt.Errorf("insufficient balance: Sol")
		} else {
			accountAssets.SolBalance -= bodyOrder.SolAmount
		}
	} else {
		if accountAssets.EcpcBalance < bodyOrder.EcpcAmount {
			return "", fmt.Errorf("insufficient balance: Ecpc")
		} else {
			accountAssets.EcpcBalance -= bodyOrder.EcpcAmount
		}
	}

	price, err := utils.FormatFloat(bodyOrder.Price)
	if err != nil {
		return "", err
	}
	// amout, err := utils.FormatFloat(bodyOrder.SolAmount)
	// if err != nil {
	// 	return "", err
	// }
	amout, err := utils.FormatFloat(bodyOrder.EcpcAmount)
	if err != nil {
		return "", err
	}

	order := model.Order{
		CreatedAt: time.Now(),
		OrderId:   utils.GenerateOrderID(),
		UserId:    bodyOrder.UserID,
		OrderSide: side,
		Price:     price,
		Total:     amout,
		Quantity:  amout,
	}

	logs.Result(fmt.Sprintf("accountAssets: %+v", accountAssets))
	logs.Result(fmt.Sprintf("order: %+v", order))

	e.MysqlDB.Save(&accountAssets)
	e.MysqlDB.Create(&order)
	e.matchOrders(side, accountAssets)

	return order.OrderId, nil
}

func (e *Exchange) DeleteOrder(orderID string) error {
	var order model.Order
	// result := e.MysqlDB.Where("order_id = ?", orderID).Delete(&order)

	result := e.MysqlDB.Where("order_id = ?", orderID).First(&order)
	if result.Error != nil {
		return result.Error
	}

	var accountAssets model.AccountAssets
	result = e.MysqlDB.Where("user_id = ?", order.UserId).First(&accountAssets)
	if result.Error != nil {
		return result.Error
	}

	if order.OrderSide == model.Buy {
		solQuantity := order.Quantity * order.Price
		accountAssets.SolBalance += solQuantity
	} else {
		accountAssets.EcpcBalance += order.Quantity
	}

	result = e.MysqlDB.Delete(&order)
	if result.Error != nil {
		return result.Error
	}

	e.MysqlDB.Save(&accountAssets)
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

func (e *Exchange) GetUserOrders(id string) ([]ResOrder, error) {
	var userOrders []model.Order

	if err := e.MysqlDB.Where("user_id = ?", id).Find(&userOrders).Error; err != nil {
		return nil, err
	}

	userOrderObs := rxgo.Just(userOrders)().Map(func(_ context.Context, i interface{}) (interface{}, error) {
		oldOrder := i.(model.Order)
		orderType := Buy
		if oldOrder.OrderSide == model.Sell {
			orderType = Sell
		}
		return ResOrder{CreatedAt: oldOrder.CreatedAt, OrderID: oldOrder.OrderId, Type: orderType, Price: oldOrder.Price, Amount: oldOrder.Quantity, Total: oldOrder.Total}, nil
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

	// sig, amount, err := e.Conn.RechargeSol(accountAssets)
	// if err != nil {
	// 	return resUser, err
	// }

	// logs.Normal(fmt.Sprintf("amount: %+v", amount))
	// logs.Normal(fmt.Sprintf("sig: %+v", sig))

	// if amount > 0 {
	// 	accountAssets.SolBalance += float64(amount) / float64(solana.LAMPORTS_PER_SOL)
	// 	accountAssets.SolTotal += float64(amount) / float64(solana.LAMPORTS_PER_SOL)
	// }

	// logs.Normal(fmt.Sprintf("accountAssets: %+v", accountAssets))

	// e.MysqlDB.Save(&accountAssets)

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

func (e *Exchange) GetUserAll() ([]ResUser, error) {
	var assets []model.AccountAssets
	e.MysqlDB.Order("ecpc_total desc").Find(&assets)

	assetsObs := rxgo.Just(assets)().Map(func(_ context.Context, i interface{}) (interface{}, error) {
		oldAssets := i.(model.AccountAssets)
		return ResUser{
			UserId:      oldAssets.UserId,
			Address:     oldAssets.Address,
			CexAddress:  oldAssets.CexAddress,
			SolBalance:  oldAssets.SolBalance,
			SolTotal:    oldAssets.SolTotal,
			EcpcBalance: oldAssets.EcpcBalance,
			EcpcTotal:   oldAssets.EcpcTotal,
		}, nil
	})
	assetsList := make([]ResUser, 0)
	for item := range assetsObs.Observe() {
		if item.Error() {
			return nil, item.E
		}
		assetsList = append(assetsList, item.V.(ResUser))
	}
	return assetsList, nil
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

func (e *Exchange) UpdateUser(user BodyUpdateUser) error {
	var accountAssets model.AccountAssets

	result := e.MysqlDB.Where("user_id = ?", user.UserID).First(&accountAssets)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("userId does not exist")
		}
		return result.Error
	}

	// Validate the Type value
	if err := user.Type.Validate(); err != nil {
		return err
	}

	amount := user.Amount
	if user.AssetType == "Sol" {
		if user.Type == Deposite {
			accountAssets.SolBalance += amount
			accountAssets.SolTotal += amount
		} else {
			accountAssets.SolBalance -= amount
			accountAssets.SolTotal -= amount
		}
	} else {
		if user.Type == Deposite {
			accountAssets.EcpcBalance += amount
			accountAssets.EcpcTotal += amount
		} else {
			accountAssets.EcpcBalance -= amount
			accountAssets.EcpcTotal -= amount
		}
	}
	e.MysqlDB.Save(&accountAssets)
	return nil
}

// func (e *Exchange) Withdraw(bodyWithdraw BodyWithdraw) (string, error) {

// 	var accountAssets model.AccountAssets
// 	result := e.MysqlDB.Where("user_id = ?", bodyWithdraw.UserID).First(&accountAssets)
// 	if result.Error != nil {
// 		if result.Error == gorm.ErrRecordNotFound {
// 			return "", fmt.Errorf("userId does not exist")
// 		}
// 		return "", result.Error
// 	}

// 	if accountAssets.SolBalance < bodyWithdraw.Amount {
// 		return "", fmt.Errorf("insufficient balance: Sol")
// 	}

// 	toAmount := bodyWithdraw.Amount * 1000000000

// 	sig, err := e.Conn.Withdraw(accountAssets.Address, uint64(toAmount))
// 	if err != nil {
// 		return "", err
// 	}

// 	accountAssets.SolBalance -= bodyWithdraw.Amount
// 	accountAssets.SolTotal -= bodyWithdraw.Amount
// 	e.MysqlDB.Save(&accountAssets)
// 	return sig, nil
// }

func clearDuplicateData(o rxgo.Observable) ([]ResOrder, error) {
	orderList := make([]ResOrder, 0)
	for item := range o.Observe() {
		if item.Error() {
			return nil, item.E
		}
		orderList = append(orderList, item.V.(ResOrder))
	}

	result := make([]ResOrder, 0)
	var oldOrder, newOrder ResOrder
	for _, order := range orderList {
		newOrder = order
		if oldOrder.Price != 0 && newOrder.Price != 0 && oldOrder.Price != newOrder.Price {
			result = append(result, oldOrder)
		}
		oldOrder = newOrder
	}
	if len(orderList) > 0 {
		result = append(result, oldOrder)
	}

	// orderMap := make(map[float64]ResOrder, len(orderList))
	// for _, order := range orderList {
	// 	if existingOrder, found := orderMap[order.Price]; found && order.Amount <= existingOrder.Amount {
	// 		continue
	// 	}
	// 	orderMap[order.Price] = order
	// }

	// result := make([]ResOrder, 0, len(orderMap))
	// for _, order := range orderMap {
	// 	result = append(result, order)
	// }

	return result, nil
}

func (e *Exchange) matchOrders(side model.OrderSide, currentAccout model.AccountAssets) {
	var buyOrder model.Order
	var sellOrder model.Order
	e.MysqlDB.Where("order_side = ?", model.Buy).Order("price desc").First(&buyOrder)
	e.MysqlDB.Where("order_side = ?", model.Sell).Order("price asc").First(&sellOrder)

	for buyOrder.ID != 0 && sellOrder.ID != 0 {
		if buyOrder.Price >= sellOrder.Price {

			tradeQuantity := utils.Min(buyOrder.Quantity, sellOrder.Quantity)
			maxSol := tradeQuantity * buyOrder.Price

			logs.Result(fmt.Sprintf("buyOrder: %+v", buyOrder))
			logs.Result(fmt.Sprintf("sellOrder: %+v", sellOrder))
			logs.Result(fmt.Sprintf("tradeQuantity: %+v", tradeQuantity))
			logs.Result(fmt.Sprintf("maxSol: %+v", maxSol))

			buyOrder.Quantity -= tradeQuantity
			sellOrder.Quantity -= tradeQuantity

			e.MysqlDB.Save(&buyOrder)
			e.MysqlDB.Save(&sellOrder)

			var otherAccount model.AccountAssets
			if side == model.Buy {

				currentAccout.SolTotal -= maxSol
				currentAccout.EcpcBalance += tradeQuantity
				currentAccout.EcpcTotal += tradeQuantity

				e.MysqlDB.Save(&currentAccout)

				e.MysqlDB.Where("user_id = ?", sellOrder.UserId).First(&otherAccount)
				otherAccount.SolTotal += maxSol
				otherAccount.SolBalance += maxSol
				otherAccount.EcpcTotal -= tradeQuantity
			} else {

				currentAccout.SolBalance += maxSol
				currentAccout.SolTotal += maxSol
				currentAccout.EcpcTotal -= tradeQuantity

				e.MysqlDB.Save(&currentAccout)

				e.MysqlDB.Where("user_id = ?", buyOrder.UserId).First(&otherAccount)
				otherAccount.SolTotal -= maxSol
				otherAccount.EcpcBalance += tradeQuantity
				otherAccount.EcpcTotal += tradeQuantity
			}

			e.MysqlDB.Save(&otherAccount)

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

func (e *Exchange) TransactionEcpc() error {

	result := e.MysqlDB.
		Model(&model.AccountAssets{}).
		Where("user_id = ?", "26b6fb7f-6146-4e97-b6d6-89f1eba76d34").
		Updates(model.AccountAssets{SolBalance: 15.259, SolTotal: 17.998})

	if result.Error != nil {
		return result.Error
	}
	return nil
}
