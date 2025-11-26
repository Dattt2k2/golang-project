package routes

import (
	"user-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, h *handlers.UserHandler) {
	users := r.Group("/me")
	{
		users.POST("", h.CreateUser)
		users.GET("", h.GetUser)
		users.PUT("", h.UpdateUser)
		users.DELETE("", h.DeleteUser)

		addresses := users.Group("/addresses")
		{
			addresses.POST("", h.AddAddress)
			addresses.PUT("/:address_id", h.UpdateAddress)
			addresses.GET("", h.GetUserAddresses)
			addresses.DELETE("/:address_id", h.DeleteAddress)
		}
	}
	admin := r.Group("/admin")
	{
		admin.GET("/users", h.ListUsers)
		admin.GET("/users/:user_id", h.GetUserByID)
		admin.PATCH("/users/:user_id/status", h.UpdateUserStatus)
		admin.DELETE("/users/:user_id", h.AdminDeleteUser)
	}
}
