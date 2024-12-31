package middleware

import (
	// "encoding/json"
	// "fmt"
	"log"
	"net/http"

	// "time"

	helper "github.com/Dattt2k2/golang-project/api-gateway/helpers"
	// "github.com/Dattt2k2/golang-project/api-gateway/redisdb"
	// grpcClient "github.com/Dattt2k2/golang-project/api-gateway/grpc"
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
        // Bỏ qua routes không cần auth
        if c.FullPath() == "/auth/users/register" || c.FullPath() == "/auth/users/login" {
            c.Next()
            return
        }

        // Kiểm tra token
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy token"})
            c.Abort()
            return
        }

        // Validate Bearer token
        if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không đúng định dạng"})
            c.Abort()
            return
        }
        tokenString = tokenString[7:]

        // Validate và lấy claims
        claims, msg := helper.ValidateToken(tokenString)
        if claims == nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
            c.Abort()
            return
        }

        // Set context từ claims
        c.Set("email", claims.Email)
        c.Set("role", claims.UserType)
        c.Set("uid", claims.Uid)

        log.Printf("Context đã được set: email=%s, role=%s, uid=%s", 
            claims.Email, claims.UserType, claims.Uid)

        c.Next()
    }
}



func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy token từ header Authorization
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is missing"})
			c.Abort()
			return
		}

		// Kiểm tra định dạng token: Bearer <token>
		if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// Loại bỏ từ "Bearer " và lấy token thực tế
		tokenString = tokenString[7:]

		// Validate token
		claims, msg := helper.ValidateToken(tokenString)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			c.Abort()
			return
		}

		// Lưu thông tin user vào context để sử dụng ở các handler tiếp theo
		c.Set("user_id", claims.Uid)
		c.Set("user_email", claims.Email)
        c.Set("role", claims.UserType)

		c.Next()
	}
}
