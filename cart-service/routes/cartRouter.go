package routes

import (
	"github.com/gin-gonic/gin"
)

func  CartRoutes(incomingRoutes *gin.Engine){
	authorized := incomingRoutes.Group("/")

	authorized.POST("/cart/:userID/add")
}