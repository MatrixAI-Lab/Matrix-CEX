package routes

import (
	"MatrixAI-CEX/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(engine *gin.Engine) {
	user := engine.Group("/user")
	{
		user.POST("/validate-code", handlers.GetValidateCode)
		user.POST("/login", handlers.Login)
	}
}
