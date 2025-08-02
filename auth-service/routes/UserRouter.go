package routes

import (
	controller "auth-service/controller"
	service "auth-service/service"

	// "auth-service/middleware"
	"auth-service/repository"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	userRepo := repository.NewUserRepository()
	authService := service.NewAuthService(userRepo)
	authController := controller.NewAuthController(authService)
	// incomingRoutes.Use(middleware.Authenticate())

	// Existing routes
	incomingRoutes.GET("/admin/get-users", authController.GetUsers())
	incomingRoutes.GET("/users/user_id", authController.GetUser())
	incomingRoutes.POST("/users/change-password", authController.ChangePassword())
	incomingRoutes.POST("/users/logout", authController.Logout())
	incomingRoutes.POST("/users/logout-all", authController.LogoutAll())
	incomingRoutes.GET("/users/devices", authController.GetDevices())
	incomingRoutes.POST("/admin/change-password", authController.AdminChangePassword())

	// API routes for Kong gateway
	incomingRoutes.GET("/api/auth/admin/get-users", authController.GetUsers())
	incomingRoutes.GET("/api/auth/users/user_id", authController.GetUser())
	incomingRoutes.POST("/api/auth/users/change-password", authController.ChangePassword())
	incomingRoutes.POST("/api/auth/users/logout", authController.Logout())
	incomingRoutes.POST("/api/auth/users/logout-all", authController.LogoutAll())
	incomingRoutes.GET("/api/auth/users/devices", authController.GetDevices())
	incomingRoutes.POST("/api/auth/admin/change-password", authController.AdminChangePassword())

	incomingRoutes.POST("/admin/update-role", authController.UpdateUserRole())
}
