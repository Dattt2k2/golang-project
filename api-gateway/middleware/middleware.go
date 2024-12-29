package middleware

import (
	// "encoding/json"
	"fmt"
	"net/http"

	// "time"

	// "github.com/Dattt2k2/golang-project/api-gateway/redisdb"
	grpcClient "github.com/Dattt2k2/golang-project/api-gateway/grpc"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc{
	return func(c *gin.Context){
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS"{
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	}
}



func Authenticate() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Bỏ qua các route không cần xác thực

        currrentPath := c.Request.URL.Path

        publicPath := []string{
            "auth/users/register",
            "auth/users/login",
        }
        
        fmt.Println("Current Path:", currrentPath)

        for _, path := range publicPath{
            if currrentPath == path{
                c.Next()
                return
            }
        }
        // if c.FullPath() == "/auth/users/register" || c.FullPath() == "/auth/users/login" {
        //     c.Next()
        //     return
        // }

        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "No Authorization header provided"})
            c.Abort()
            return
        }

        res, err := grpcClient.VerifyToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        c.Set("email", res.Email)
        c.Set("role", res.UserType)
        c.Set("uid", res.Uid)

        c.Next()
    }
}
