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

func OrderRoutes(incomming *gin.Engine) {
	orderController := SetupOrderController()

	authorized := incomming.Group("/")

	// Existing routes
	authorized.POST("order/cart", orderController.OrderFromCart())
	authorized.POST("order/direct", orderController.OrderDirectly())
	authorized.GET("order/user", orderController.GetUserOrders())
	authorized.GET("admin/orders", orderController.AdminGetOrders())
	authorized.POST("user/order/cancel/:order_id", orderController.CancelOrder())

	// API routes for Kong gateway
	authorized.POST("/api/orders/cart", orderController.OrderFromCart())
	authorized.POST("/api/orders/direct", orderController.OrderDirectly())
	authorized.GET("/api/orders/user", orderController.GetUserOrders())
	authorized.GET("/api/orders/admin/orders", orderController.AdminGetOrders())
	authorized.POST("/api/orders/user/order/cancel/:order_id", orderController.CancelOrder())
}
