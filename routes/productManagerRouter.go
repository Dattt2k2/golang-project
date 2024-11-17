package routes

import (
	controller "github.com/Dattt2k2/golang-project/controllers/sellers"
	"github.com/Dattt2k2/golang-project/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func ProductManagerRoutes(incomingRoutes *gin.Engine, db *mongo.Database){

	authorized := incomingRoutes.Group("/")
	authorized.Use(middleware.Authenticate())

	// add product to database
	authorized.POST("/products", controller.AddProduct(db))
	// Edit product from database
	authorized.PUT("/products/:id", controller.EditProduct())
	// Delete product from databse
	authorized.DELETE("/products/:id", controller.DeleteProduct())
	// Get all product from database
	authorized.GET("/products", controller.GetAllProducts(db))
	// search product
	authorized.GET("/products/search", controller.GetProdctByNameHander())
}