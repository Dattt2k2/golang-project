package middleware

import (
	// "encoding/json"
	// "fmt"
	"log"
	"net/http"
	"strings"

	// "time"

	helper "github.com/Dattt2k2/golang-project/api-gateway/helpers"
	"github.com/google/uuid"

	// "github.com/Dattt2k2/golang-project/api-gateway/redisdb"
	// grpcClient "github.com/Dattt2k2/golang-project/api-gateway/grpc"
	"github.com/gin-gonic/gin"
)

// func CORSMiddleware() gin.HandlerFunc{
// 	return func(c *gin.Context){
// 		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
// 		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
// 		// c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
// 		if c.Request.Method == "OPTIONS"{
// 			c.AbortWithStatus(http.StatusOK)
// 			return
// 		}
// 		c.Next()
// 	}
// }

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Để cho phép Authorization

		// Nếu là yêu cầu OPTIONS, phản hồi thành công ngay lập tức
		if c.Request.Method == "OPTIONS" {
			// Chỉ cần xử lý cho OPTIONS
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// Nếu là đăng nhập hoặc đăng ký, không yêu cầu Authorization
		if c.FullPath() == "/auth/users/login" || c.FullPath() == "/auth/users/register" {
			c.Next()
			return
		}

		// // Kiểm tra token cho các route còn lại
		// tokenString := c.GetHeader("Authorization")
		// if tokenString == "" {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy token"})
		// 	c.Abort()
		// 	return
		// }

		// // Validate Bearer token
		// if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không đúng định dạng"})
		// 	c.Abort()
		// 	return
		// }
		// tokenString = tokenString[7:]

		// // Validate và lấy claims
		// claims, msg := helper.ValidateToken(tokenString)
		// if claims == nil {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
		// 	c.Abort()
		// 	return
		// }

		// // Set context từ claims
		// c.Set("email", claims.Email)
		// c.Set("role", claims.UserType)
		// c.Set("uid", claims.Uid)

		// log.Printf("Context đã được set: email=%s, role=%s, uid=%s",
		// 	claims.Email, claims.UserType, claims.Uid)

		c.Next()
	}
}

// func Authenticate() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// Bỏ qua routes không cần auth
// 		if c.FullPath() == "/auth/users/register" || c.FullPath() == "/auth/users/login" {
// 			c.Next()
// 			return
// 		}

// 		// var tokenString string

// 		// authHeader := c.GetHeader("Authorization")
// 		// if authHeader != "" {
// 		// 	if strings.HasPrefix(authHeader, "Bearer ") {
// 		// 		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
// 		// 	} else {
// 		// 		tokenString = authHeader
// 		// 	}
// 		// }

// 		// if tokenString == "" {
// 		// 	var err error
// 		// 	tokenString, err := c.Cookie("auth_token")
// 		// 	if err != nil || tokenString == "" {
// 		// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
// 		// 		c.Abort()
// 		// 		return
// 		// 	}
// 		// }

// 		// claims, msg := helper.ValidateToken(tokenString)
// 		// if claims == nil {
// 		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
// 		// 	c.Abort()
// 		// 	return
// 		// }

// 		// c.Set("email", claims.Email)
// 		// c.Set("role", claims.UserType)
// 		// c.Set("uid", claims.Uid)

// 		// log.Printf("User authenticated: email=%s, role=%s, uid=%s",
// 		// 	claims.Email, claims.UserType, claims.Uid)
// 		c.Next()

// 		// Kiểm tra token
// 		// tokenString := c.GetHeader("Authorization")
// 		// if tokenString == "" {
// 		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Không tìm thấy token"})
// 		// 	c.Abort()
// 		// 	return
// 		// }

// 		// // Validate Bearer token
// 		// if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
// 		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Token không đúng định dạng"})
// 		// 	c.Abort()
// 		// 	return
// 		// }
// 		// tokenString = tokenString[7:]

// 		// // Validate và lấy claims
// 		// claims, msg := helper.ValidateToken(tokenString)
// 		// if claims == nil {
// 		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
// 		// 	c.Abort()
// 		// 	return
// 		// }

// 		// // Set context từ claims
// 		// c.Set("email", claims.Email)
// 		// c.Set("role", claims.UserType)
// 		// c.Set("uid", claims.Uid)

// 		// log.Printf("Context đã được set: email=%s, role=%s, uid=%s",
// 		// 	claims.Email, claims.UserType, claims.Uid)

// 		// c.Next()
// 	}
// }



func DeviceInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceId := c.GetHeader("Device-Id")
		if deviceId == "" {
			deviceId = genterateDeviceId(c)
			c.Request.Header.Set("Device-Id", deviceId)
		}

		platform := c.GetHeader("X-Platform")
		if platform == "" {
			userAgent := c.GetHeader("User-Agent")
			platform = detectPlatform(userAgent)
			c.Request.Header.Set("X-Platform", platform)
		}

		c.Next()
	}
}

func genterateDeviceId(c *gin.Context) string {
	if deviceId, err := c.Cookie("device_id"); err == nil {
		return deviceId
	}

	deviceId := uuid.New().String()
	if !isMobileRequest(c.GetHeader("User-Agent")) {
		c.SetCookie("device_id", deviceId, 86400*365, "/", "", false, true)
	}

	return deviceId
}

func detectPlatform(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, "android") {
		return "android"
	} else if strings.Contains(userAgent, "ios") {
		return "ios"
	} else {
		return "web"
	}
}

func isMobileRequest(userAgent string) bool {
	userAgent = strings.ToLower(userAgent)
	return strings.Contains(userAgent, "android") || strings.Contains(userAgent, "ios")
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Try to get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				tokenString = authHeader
			}
			log.Printf("[DEBUG] Found token in Authorization header")
		}

		// If not in header, try cookie
		if tokenString == "" {
			var err error
			tokenString, err = c.Cookie("auth_token")
			if err == nil {
				log.Printf("[DEBUG] Found token in cookie")
			} else {
				log.Printf("[DEBUG] No token in cookie: %v", err)
			}
		}

		// If still no token, return error
		if tokenString == "" {
			log.Printf("[DEBUG] No token found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Validate token
		claims, msg := helper.ValidateToken(tokenString)
		if msg != "" {
			log.Printf("[DEBUG] Token validation failed: %s", msg)

			// If token is expired but we have a refresh token, try to refresh
			if msg == "token is expired" {
				refreshToken, err := c.Cookie("refresh_token")
				if err == nil && refreshToken != "" {
					newToken, refreshMsg := helper.RefreshToken(refreshToken)
					if refreshMsg == "" {
						// Set the new token in cookie
						c.SetCookie(
							"auth_token",
							newToken,
							60*60*24*7, // 7 days
							"/",
							"",
							c.Request.TLS != nil,
							true,
						)

						// Continue with new token
						tokenString = newToken
						claims, msg = helper.ValidateToken(newToken)
						log.Printf("[DEBUG] Token refreshed successfully")
					} else {
						log.Printf("[DEBUG] Failed to refresh token: %s", refreshMsg)
						c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired, please login again"})
						c.Abort()
						return
					}
				} else {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired, please login again"})
					c.Abort()
					return
				}
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
				c.Abort()
				return
			}
		}

		// Set user info in context
		c.Set("email", claims.Email)
		c.Set("role", claims.UserType)
		c.Set("uid", claims.Uid)

		c.Next()
	}
}
