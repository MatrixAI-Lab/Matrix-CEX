package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"MatrixAI-CEX/chain/conn"
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

	r.POST("/order", func(c *gin.Context) {
		var order exchange.BodyOrder
		if err := c.ShouldBindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		id := ex.AddOrder(order)
		c.JSON(http.StatusOK, gin.H{"message": "Order added successfully", "id": id})
	})

	r.DELETE("/order/:id", func(c *gin.Context) {
		id := c.Param("id")
		if err = ex.DeleteOrder(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
	})

	r.GET("/trades", func(c *gin.Context) {
		buyOrderList, sellOrderList, err := ex.GetAndProcessOrders()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"buyOrders": buyOrderList, "sellOrders": sellOrderList})
	})

	r.POST("/register", func(c *gin.Context) {
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
	})

	r.GET("/user/:userId", func(c *gin.Context) {
		userId := c.Param("userId")
		accountAssets, err := ex.GetUserAccountAssets(userId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "get user account assets successfully", "accountAssets": accountAssets})
	})

	r.Run("")
}
