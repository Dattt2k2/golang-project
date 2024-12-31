package middleware

import (
	// "encoding/json"
	// "fmt"
	"net/http"
    "log"

	// "time"

    helper "github.com/Dattt2k2/golang-project/auth-service/helpers"
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
        // Bỏ qua các route không cần xác thực
        if c.FullPath() == "/auth/users/register" || c.FullPath() == "/auth/users/login" {
            c.Next()
            return
        }

        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "No Authorization header provided"})
            c.Abort()
            return
        }

        // Kiểm tra định dạng token: Bearer <token>
        if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
            c.Abort()
            return
        }

        // Lấy token thực tế từ header
        tokenString = tokenString[7:]

        claims, msg := helper.ValidateToken(tokenString)
        if claims == nil && msg == "Token is expired" {
            // Nếu token hết hạn, yêu cầu client cung cấp refresh token
            refreshToken := c.GetHeader("Refresh-Token")
            if refreshToken == "" {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token is required"})
                c.Abort()
                return
            }

            claims, msg := helper.ValidateToken(refreshToken)
            if claims == nil {
                c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
                c.Abort()
                return
            }

            // Tạo lại access token mới từ refresh token
            _, newAccessToken, err := helper.GenerateAllToken(claims.Email, claims.FirstName, claims.LastName, claims.UserType, claims.Uid)
            if err != nil {
                c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to generate new access token"})
                c.Abort()
                return
            }

            // Trả về token mới
            c.JSON(http.StatusOK, gin.H{"access_token": newAccessToken})
            c.Abort() 
            return
        }

        // Nếu token hợp lệ, lưu thông tin vào context
        if claims != nil{
            log.Printf("Setting context values: email=%s, role=%s, uid=%s", claims.Email, claims.UserType, claims.Uid)
            c.Set("email", claims.Email)
            c.Set("role", claims.UserType)
            c.Set("uid", claims.Uid)
        }

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

		c.Next()
	}
}
