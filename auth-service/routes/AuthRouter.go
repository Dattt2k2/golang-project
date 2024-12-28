package routes

import(
	controller "github.com/Dattt2k2/golang-project/auth-service/controller"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.POST("users/register", controller.SignUp())
	incomingRoutes.POST("users/login", controller.Login())
}

