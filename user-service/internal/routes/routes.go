package routes

import (
	"user-service/internal/handlers"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, h *handlers.UserHandler) {
	users := r.Group("/users")
	{
		users.POST("/", h.CreateUser)
		users.GET(":id", h.GetUser)
		users.PUT(":id", h.UpdateUser)
		users.DELETE(":id", h.DeleteUser)
	}
}
