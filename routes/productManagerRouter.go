package routes

import (
	controller "github.com/Dattt2k2/golang-project/controllers/sellers"
	"github.com/Dattt2k2/golang-project/middleware"
	"github.com/gin-gonic/gin"
)

func ProductManagerRoutes(incomingRoutes *gin.Engine){

	authorized := incomingRoutes.Group("/")
	authorized.Use(middleware.Authenticate())

	// add product to database
	authorized.POST("/products", controller.AddProduct())
	// Edit product from database
	authorized.PUT("/products/:id", controller.EditProduct())
	// Delete product from databse
	authorized.DELETE("/products/:id", controller.DeleteProduct())
	// Get all product from database
	authorized.GET("/products", controller.GetAllProducts())
	// search product
	authorized.GET("/products/search", controller.GetProdctByNameHander())
}