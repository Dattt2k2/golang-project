package routes

import (
	service "github.com/Dattt2k2/golang-project/auth-service/service"
	controller "github.com/Dattt2k2/golang-project/auth-service/controller"
	"github.com/Dattt2k2/golang-project/auth-service/repository"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine){
	userRepo := repository.NewUserRepository()
	authService := service.NewAuthService(userRepo)
	authController := controller.NewAuthController(authService)


	incomingRoutes.POST("users/register", authController.SignUp())
	incomingRoutes.POST("users/login", authController.Login())
}

