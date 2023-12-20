package routes

import (
	"MatrixAI-CEX/handlers"
	"MatrixAI-CEX/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine) {
	user := engine.Group("/user")
	{
		user.POST("/email-code", handlers.EmailCode)
		user.POST("/login", handlers.Login)
	}
	market := engine.Group("/market", middleware.Jwt())
	{
		market.POST("/place-order", handlers.PlaceOrder)
		market.POST("/renew-order", handlers.RenewOrder)
	}
}
