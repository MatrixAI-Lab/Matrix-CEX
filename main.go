package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"MatrixAI-CEX/chain/conn"
	"MatrixAI-CEX/config"
	"MatrixAI-CEX/db/mysql"
	"MatrixAI-CEX/exchange"
)

var ex = &exchange.Exchange{}

func main() {

	mysqlDB, err := mysql.InitDB()
	if err != nil {
		panic(err)
	}
	ex.MysqlDB = mysqlDB

	newConn, err := conn.NewConn()
	if err != nil {
		panic(err)
	}
	ex.Conn = newConn

	r := gin.Default()

	r.POST("/order", addOrder)
	r.DELETE("/order/:id", deleteOrder)
	r.GET("/trades", getTrades)
	r.GET("/userOrders", getUserOrders)
	r.POST("/register", registerUser)
	r.GET("/user/:userId", getUserAccountAssets)
	r.GET("/transactionRecords", getTransactionRecords)
	r.Run(config.URL)
}

func addOrder(c *gin.Context) {
	var order exchange.BodyOrder
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := ex.AddOrder(order)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Order added successfully", "id": id})
}

func deleteOrder(c *gin.Context) {
	id := c.Param("id")
	if err := ex.DeleteOrder(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}

func getTrades(c *gin.Context) {
	buyOrderList, sellOrderList, err := ex.GetAndProcessOrders()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"buyOrders": buyOrderList, "sellOrders": sellOrderList})
}

func getUserOrders(c *gin.Context) {
	var user exchange.BodyUser

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userOrderList, err := ex.GetUserOrders(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "registered successfully", "userId": userId})
}

func getUserAccountAssets(c *gin.Context) {
	userId := c.Param("userId")
	accountAssets, err := ex.GetUserAccountAssets(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "get user account assets successfully", "accountAssets": accountAssets})
}

func getTransactionRecords(c *gin.Context) {
	records, err := ex.GetTransactionRecords()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "get transaction records successfully", "records": records})
}