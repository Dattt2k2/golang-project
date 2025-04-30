package routes

import (
	"github.com/Dattt2k2/golang-project/search-service/controller"
	"github.com/gin-gonic/gin"
)

func SearchRoutes(router *gin.Engine, ctrl *controller.SearchController) {
	router.GET("/search", ctrl.BasicSearch("search"))
	router.GET("/search/advanced", ctrl.AdvancedSearch("advanced", map[string]interface{}{}))
}