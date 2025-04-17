package routes

import (
	"github.com/Dattt2k2/golang-project/cart-service/database"
	"github.com/Dattt2k2/golang-project/order-service/controller"
	"github.com/Dattt2k2/golang-project/order-service/repositories"
	orderService "github.com/Dattt2k2/golang-project/order-service/service"
	"github.com/gin-gonic/gin"
)
func SetupOrderController() *controller.OrderController {

	orderRepo := repositories.NewOrderRepository(database.OpenCollection(database.Client, "order"))
	// Import the service package and create a service instance
	orderSvc := orderService.NewOrderService(orderRepo)

	return controller.NewOrderController(orderSvc)
}



func OrderRoutes(incomming *gin.Engine){
	orderController := SetupOrderController()

	authorized := incomming.Group("/")

	authorized.POST("order/cart", orderController.OrderFromCart())
	authorized.POST("order/direct/:id", orderController.OrderDirectly())
	authorized.GET("order/user", orderController.GetUserOrders())
	authorized.GET("admin/orders", orderController.AdminGetOrders())
}