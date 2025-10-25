package routes

import (
	"user-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, h *handlers.UserHandler) {
	users := r.Group("/users")
	{
		users.POST("/", h.CreateUser)
		users.GET("/", h.GetUser)
		users.PUT("/", h.UpdateUser)
		users.DELETE("/", h.DeleteUser)
	}
}
