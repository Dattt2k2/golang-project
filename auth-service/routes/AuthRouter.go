package routes

import (
	controller "auth-service/controller"
	"auth-service/repository"
	service "auth-service/service"
	"auth-service/websocket"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	userRepo := repository.NewUserRepository()
	authService := service.NewAuthService(userRepo)
	authController := controller.NewAuthController(authService)

	authGroup := router.Group("/auth")

	authGroup.POST("/users/register", authController.SignUp())
	authGroup.POST("/users/login", authController.Login())
	authGroup.POST("/refresh-token", authController.RefreshToken())

	authGroup.POST("/verify-otp", authController.VerifyOTP())
	authGroup.POST("/resend-otp", authController.ResendOTP())
	authGroup.GET("/ws", websocket.WebSocketHander)
}
