package routes

import (
	"MatrixAI-CEX/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine) {
	user := engine.Group("/user")
	{
		user.POST("/email-code", handlers.EmailCode)
		user.POST("/login", handlers.Login)
	}
}
