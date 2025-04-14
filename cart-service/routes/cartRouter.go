package routes

import (
	"github.com/Dattt2k2/golang-project/cart-service/controller"
	"github.com/gin-gonic/gin"
)

func  CartRoutes(incomingRoutes *gin.Engine){
	authorized := incomingRoutes.Group("/")

	authorized.POST("/cart/add/:id", controller.AddToCart())
	authorized.GET("/cart/get/:id", controller.GetCart())
	authorized.GET("/cart/get", controller.GetCartSeller())
	authorized.DELETE("/cart/delete/:id", controller.DeleteProductFromCart())
}