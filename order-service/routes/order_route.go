package routes

import (
	"github.com/Dattt2k2/golang-project/order-service/controller"
	"github.com/Dattt2k2/golang-project/order-service/repositories"
	orderService "github.com/Dattt2k2/golang-project/order-service/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)
func SetupOrderController() *controller.OrderController {

	orderRepo := repositories.NewOrderRepository(&mongo.Collection{})
	// Import the service package and create a service instance
	orderSvc := orderService.NewOrderService(orderRepo)

	return controller.NewOrderController(orderSvc)
}



func OrderRoutes(incomming *gin.Engine){
	orderController := SetupOrderController()

	authorized := incomming.Group("/")

	authorized.POST("order/cart/:id", orderController.OrderFromCart())
	authorized.POST("order/direct/:id", orderController.OrderDirectly())
	// authorized.GET("order", controller.GetOrder())
}