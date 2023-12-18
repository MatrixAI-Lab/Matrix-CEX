package handlers

import (
	"MatrixAI-CEX/chain/matrix"
	"MatrixAI-CEX/common"
	"MatrixAI-CEX/db/mysql/model"
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
	userId := c.MustGet("userId").(string)
	var accountAssets model.AccountAssets
	dbResult := common.Db.
		Where("user_Id = ?", userId).
		Take(&accountAssets)
	if dbResult.Error != nil {
		logs.Error(fmt.Sprintf("Database error: %s \n", dbResult.Error))
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
		resp.Fail(c, "Transaction fail")
		return
	}

	accountAssets.EcpcBalance -= req.Total
	accountAssets.EcpcTotal -= req.Total

	dbResult = common.Db.Save(&accountAssets)
	if dbResult.Error != nil {
		logs.Error(fmt.Sprintf("Database error: %s \n", dbResult.Error))
		resp.Fail(c, "Database error")
		return
	}

	response := MarketResp{TxHash: txHash}
	resp.Success(c, response)
}
