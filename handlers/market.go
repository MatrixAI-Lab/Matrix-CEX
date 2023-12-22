package handlers

import (
	"MatrixAI-CEX/chain/matrix"
	"MatrixAI-CEX/common"
	logs "MatrixAI-CEX/utils/log_utils"
	"MatrixAI-CEX/utils/resp"
	"fmt"
	"github.com/gin-gonic/gin"
)

type MarketResp struct {
	TxHash string
}

type PlaceOrderReq struct {
	matrix.PlaceOrderParams
}

func PlaceOrder(c *gin.Context) {
	accountAssets, err := getAccount(c)
	if err != nil {
		resp.Fail(c, "User not found")
		return
	}

	var req PlaceOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, "Parameter missing")
		return
	}

	if req.Total > accountAssets.EcpcBalance {
		resp.Fail(c, "Insufficient balance")
		return
	}

	txHash, err := matrix.PlaceOrder(req.PlaceOrderParams, accountAssets.CexPrivateKey)
	if err != nil {
		logs.Error(fmt.Sprintf("Transaction error: %s \n", err))
		resp.Fail(c, "Transaction fail")
		return
	}

	accountAssets.EcpcBalance -= req.Total
	accountAssets.EcpcTotal -= req.Total

	dbResult := common.Db.Save(accountAssets)
	if dbResult.Error != nil {
		logs.Error(fmt.Sprintf("Database error: %s \n", dbResult.Error))
		resp.Fail(c, "Database error")
		return
	}

	response := MarketResp{TxHash: txHash}
	resp.Success(c, response)
}

type RenewOrderReq struct {
	matrix.RenewOrderParams
}

func RenewOrder(c *gin.Context) {
	accountAssets, err := getAccount(c)
	if err != nil {
		resp.Fail(c, "User not found")
		return
	}

	var req RenewOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, "Parameter missing")
		return
	}

	if req.Total > accountAssets.EcpcBalance {
		resp.Fail(c, "Insufficient balance")
		return
	}

	txHash, err := matrix.RenewOrder(req.RenewOrderParams, accountAssets.CexPrivateKey)
	if err != nil {
		resp.Fail(c, "Transaction fail")
		return
	}

	accountAssets.EcpcBalance -= req.Total
	accountAssets.EcpcTotal -= req.Total

	dbResult := common.Db.Save(accountAssets)
	if dbResult.Error != nil {
		logs.Error(fmt.Sprintf("Database error: %s \n", dbResult.Error))
		resp.Fail(c, "Database error")
		return
	}

	response := MarketResp{TxHash: txHash}
	resp.Success(c, response)
}
