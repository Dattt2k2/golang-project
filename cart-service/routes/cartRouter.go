package routes

import (
	"github.com/Dattt2k2/golang-project/middleware"
	"github.com/gin-gonic/gin"
)

func  CartRoutes(incomingRoutes *gin.Engine){
	authorized := incomingRoutes.Group("/")
	authorized.Use(middleware.Authenticate())

	authorized.POST("/cart/:userID/add")
}