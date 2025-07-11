package routes

import (
	controller "auth-service/controller"
	"auth-service/repository"
	service "auth-service/service"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	userRepo := repository.NewUserRepository()
	authService := service.NewAuthService(userRepo)
	authController := controller.NewAuthController(authService)

	authGroup := router.Group("/auth")

	authGroup.POST("/users/register", authController.SignUp())
	authGroup.POST("/users/login", authController.Login())
}
