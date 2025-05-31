package routes

import (
	"github.com/Dattt2k2/golang-project/search-service/controller"
	"github.com/gin-gonic/gin"
)

func SearchRoutes(router *gin.Engine, ctrl *controller.SearchController) {
	// Existing routes
	router.GET("/search", ctrl.BasicSearch("search"))
	router.GET("/search/advanced", ctrl.AdvancedSearch("advanced", map[string]interface{}{}))

	// API routes for Kong gateway
	router.GET("/api/search", ctrl.BasicSearch("search"))
	router.GET("/api/search/advanced", ctrl.AdvancedSearch("advanced", map[string]interface{}{}))
}
