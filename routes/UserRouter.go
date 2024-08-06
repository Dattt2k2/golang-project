package routes

import(
	controller "github.com/Dattt2k2/golang-project/controllers/admin"
	"github.com/Dattt2k2/golang-project/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users//user_id", controller.GetUser())
}



