package routes

import (

	controller "github.com/Dattt2k2/golang-project/product-service/controller"
	"github.com/gin-gonic/gin"
	// "go.mongodb.org/mongo-driver/mongo"
)

func ProductManagerRoutes(incomingRoutes *gin.Engine){

	authorized := incomingRoutes.Group("/")

	// add product to database
	authorized.POST("/products/add", controller.AddProduct())
	// Edit product from database
	authorized.PUT("/products/edit/:id", controller.EditProduct())
	// Delete product from databse
	authorized.DELETE("/products/delete/:id", controller.DeleteProduct())
	// Get all product from database
	authorized.GET("/products/get", controller.GetAllProducts())
	// search product
	authorized.GET("/products/search", controller.GetProdctByNameHander())
	// get product image
	authorized.GET("/products/images/:filename", controller.GetProductImage())
	authorized.GET("images/:filename", controller.GetProductImage())
	authorized.GET("/verify", controller.VerifyImageExists())
	// In your ProductManagerRoutes function
incomingRoutes.StaticFS("/static-images", gin.Dir("product-service/uploads/images", true))
}