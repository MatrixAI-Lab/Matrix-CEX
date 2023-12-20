package handlers

import (
	"MatrixAI-CEX/common"
	"MatrixAI-CEX/db/mysql/model"
	logs "MatrixAI-CEX/utils/log_utils"
	"fmt"
	"github.com/gin-gonic/gin"
)

func getAccount(c *gin.Context) (*model.AccountAssets, error) {
	userId := c.MustGet("userId").(string)
	var accountAssets model.AccountAssets
	dbResult := common.Db.
		Where("user_Id = ?", userId).
		Take(&accountAssets)
	if dbResult.Error != nil {
		logs.Error(fmt.Sprintf("Database error: %s \n", dbResult.Error))
		return nil, fmt.Errorf("error sending transaction: %v", dbResult.Error)
	}
	return &accountAssets, nil
}
