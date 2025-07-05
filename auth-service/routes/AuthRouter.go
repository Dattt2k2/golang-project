package routes

import (
	service "auth-service/service"
	controller "auth-service/controller"
	"auth-service/repository"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine){
	userRepo := repository.NewUserRepository()
	authService := service.NewAuthService(userRepo)
	authController := controller.NewAuthController(authService)


	incomingRoutes.POST("users/register", authController.SignUp())
	incomingRoutes.POST("users/login", authController.Login())
}

