package routes

import (
	controller "github.com/Dattt2k2/golang-project/auth-service/controller"
	"github.com/Dattt2k2/golang-project/auth-service/repository"
	service "github.com/Dattt2k2/golang-project/auth-service/service"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	userRepo := repository.NewUserRepository()
	authService := service.NewAuthService(userRepo)
	authController := controller.NewAuthController(authService)

	// Existing routes
	incomingRoutes.POST("users/register", authController.SignUp())
	incomingRoutes.POST("users/login", authController.Login())
	incomingRoutes.POST("users/validate", authController.ValidateToken())

	// API routes for Kong gateway
	incomingRoutes.POST("/api/auth/users/register", authController.SignUp())
	incomingRoutes.POST("/api/auth/users/login", authController.Login())
	incomingRoutes.POST("/api/auth/users/validate", authController.ValidateToken())
}
