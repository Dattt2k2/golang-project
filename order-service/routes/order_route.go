package routes

import (
	"order-service/controller"
	"order-service/database"
	"order-service/repositories"
	orderService "order-service/service"

	"github.com/gin-gonic/gin"
)

func SetupOrderController() *controller.OrderController {

	db := database.InitDB() // This returns *gorm.DB
	orderRepo := repositories.NewOrderRepository(db)
	orderSvc := orderService.NewOrderService(orderRepo)

	return controller.NewOrderController(orderSvc)
}

func OrderRoutes(incomming *gin.Engine) {
	orderController := SetupOrderController()

	authorized := incomming.Group("/")

	authorized.POST("order/cart", orderController.OrderFromCart())
	authorized.POST("order/direct", orderController.OrderDirectly())
	authorized.GET("order/user", orderController.GetUserOrders())
	authorized.GET("orders", orderController.GetOrdersByVendor())
	authorized.POST("orders/cancel/:order_id", orderController.CancelOrder())
	authorized.GET("order/:id", orderController.GetOrderByID())
	authorized.POST("orders/:id/update-status", orderController.UpdateOrderStatus())
	authorized.POST("orders/:id/vendor-update-status", orderController.VendorUpdateOrderStatus())
	authorized.GET("orders/statistics", orderController.GetOrderStatistics())

	authorized.POST("orders/:id/confirm-delivery", orderController.ConfirmDelivery())
	authorized.POST("orders/:id/mark-shipped", orderController.MarkAsShipped())
	authorized.GET("orders/:id/status", orderController.GetOrderStatus())
	authorized.POST("orders/:id/release-payment", orderController.ReleasePaymentManually())

	// Payment callback routes (for payment-service)
	authorized.POST("orders/:id/payment/success", orderController.HandlePaymentSuccess())
	authorized.POST("orders/:id/payment/failure", orderController.HandlePaymentFailure())
}
