package main

import (
	"MatrixAI-CEX/chain/conn"
	"MatrixAI-CEX/db"
	"MatrixAI-CEX/routes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"MatrixAI-CEX/config"
	"MatrixAI-CEX/db/mysql"
	"MatrixAI-CEX/exchange"
	logs "MatrixAI-CEX/utils/log_utils"
)

var ex = &exchange.Exchange{}

func main() {

	mysqlDB, err := mysql.InitDB()
	if err != nil {
		panic(err)
	}
	ex.MysqlDB = mysqlDB

	db.InitRedis()

	newConn, err := conn.NewConn()
	if err != nil {
		panic(err)
	}
	ex.Conn = newConn

	r := gin.Default()
	routes.RegisterRoutes(r)

	r.POST("/order", addOrder)
	r.DELETE("/order/:orderId", deleteOrder)
	r.GET("/trades", getTrades)
	r.GET("/userOrders/:userId", getUserOrders)
	r.POST("/register", registerUser)
	r.GET("/user/:userId", getUserAccountAssets)
	r.GET("/userAll", getUserAll)
	r.GET("/transactionRecords", getTransactionRecords)
	r.POST("/updateUser", updateUser)
	// r.POST("/withdraw", withdraw)
	r.POST("/debug", debug)
	r.Run(config.URL)
}

func addOrder(c *gin.Context) {
	var order exchange.BodyOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logs.Result(fmt.Sprintf("/addOrder: %+v", order))

	id, err := ex.AddOrder(order)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		logs.Error(fmt.Sprintf("/addOrder: %+v", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order added successfully", "id": id})
}

func deleteOrder(c *gin.Context) {
	orderId := c.Param("orderId")
	if err := ex.DeleteOrder(orderId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		logs.Error(fmt.Sprintf("/deleteOrder: %+v", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}

func getTrades(c *gin.Context) {
	buyOrderList, sellOrderList, err := ex.GetAndProcessOrders()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		logs.Error(fmt.Sprintf("/getTrades: %+v", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"buyOrders": buyOrderList, "sellOrders": sellOrderList})
}

func getUserOrders(c *gin.Context) {
	id := c.Param("userId")

	userOrderList, err := ex.GetUserOrders(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		logs.Error(fmt.Sprintf("/getUserOrders: %+v", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"userOrders": userOrderList})
}

func registerUser(c *gin.Context) {
	var register exchange.BodyRegister
	if err := c.ShouldBindJSON(&register); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId, err := ex.RegisterUser(register)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		logs.Error(fmt.Sprintf("/register: %+v", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "registered successfully", "userId": userId})
}

func getUserAccountAssets(c *gin.Context) {
	userId := c.Param("userId")
	accountAssets, err := ex.GetUserAccountAssets(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		logs.Error(fmt.Sprintf("/getUserAccountAssets: %+v", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "get user account assets successfully", "accountAssets": accountAssets})
}

func getUserAll(c *gin.Context) {
	userAll, err := ex.GetUserAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		logs.Error(fmt.Sprintf("/getUserAll: %+v", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "get user all successfully", "userAll": userAll})
}

func getTransactionRecords(c *gin.Context) {
	records, err := ex.GetTransactionRecords()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		logs.Error(fmt.Sprintf("/getTransactionRecords: %+v", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "get transaction records successfully", "records": records})
}

func updateUser(c *gin.Context) {
	var user exchange.BodyUpdateUser
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ex.UpdateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		logs.Error(fmt.Sprintf("/updateUser: %+v", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "update user successfully"})
}

// func withdraw(c *gin.Context) {
// 	var withdraw exchange.BodyWithdraw
// 	if err := c.ShouldBindJSON(&withdraw); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}
// 	sig, err := ex.Withdraw(withdraw)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		logs.Error(fmt.Sprintf("/withdraw: %+v", err))
// 		return
// 	}
// 	c.JSON(http.StatusOK, gin.H{"message": "withdraw successfully", "sig": sig})
// }

func debug(c *gin.Context) {
	var debug exchange.BodyDebug
	if err := c.ShouldBindJSON(&debug); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if debug.Password != config.DEBUG_PASSWORD {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wrong password"})
		return
	}

	err := ex.TransactionEcpc()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "debug successfully"})
}
