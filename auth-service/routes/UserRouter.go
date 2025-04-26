package routes

import (
	controller "github.com/Dattt2k2/golang-project/auth-service/controller"
	service "github.com/Dattt2k2/golang-project/auth-service/service"
	// "github.com/Dattt2k2/golang-project/auth-service/middleware"
	"github.com/Dattt2k2/golang-project/auth-service/repository"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine){
	userRepo := repository.NewUserRepository()
	authService := service.NewAuthService(userRepo)
	authController := controller.NewAuthController(authService)
	// incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", authController.GetUsers())
	incomingRoutes.GET("/users//user_id", authController.GetUser())
	incomingRoutes.POST("/users/change-password", authController.ChangePassword())
	incomingRoutes.POST("/users/logout", authController.Logout())
	incomingRoutes.POST("/users/logout-all", authController.LogoutAll())
	incomingRoutes.GET("/users/devices", authController.GetDevices())
	incomingRoutes.POST("/admin/admin-change-password", authController.AdminChangePassword())
}



