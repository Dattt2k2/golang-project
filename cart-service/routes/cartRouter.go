package routes

import (
	"github.com/Dattt2k2/golang-project/cart-service/controller"
	"github.com/gin-gonic/gin"
)

func  CartRoutes(incomingRoutes *gin.Engine){
	authorized := incomingRoutes.Group("/")

	authorized.POST("/cart/add/:id", controller.AddToCart())
}