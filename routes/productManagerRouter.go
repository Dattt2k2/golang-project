package routes

import (
	controller "github.com/Dattt2k2/golang-project/controllers/sellers"
	"github.com/Dattt2k2/golang-project/middleware"
	"github.com/gin-gonic/gin"
)

func ProductManagerRoutes(incomingRoutes *gin.Engine){

	authorized := incomingRoutes.Group("/")
	authorized.Use(middleware.Authenticate())

	authorized.POST("/products", controller.AddProduct())
	// authorized.PUT("/products/:id", controller.EditProduct())
	authorized.DELETE("/products/:id", controller.DeleteProduct())
	authorized.GET("/products", controller.GetAllProducts())
}