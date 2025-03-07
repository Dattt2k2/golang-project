package routes

import (
	"github.com/Dattt2k2/golang-project/order-service/controller"
	"github.com/gin-gonic/gin"
)


func OrderRotes(incomming *gin.Engine){
	authorized := incomming.Group("/")

	authorized.POST("/order/:userId", controller.OrderFromCart())
	authorized.POST("/order-direct/:userId", controller.OrderDirectly())
}