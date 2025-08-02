package middleware

import (
	// "encoding/json"
	// "fmt"
	"net/http"
	"strings"

	// "time"

	helper "api-gateway/helpers"
	"api-gateway/logger"

	"api-gateway/models"
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

		

		c.Next()
	}
}


func RequireUserRole(role string) gin.HandlerFunc {
	return func (c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists || userRole != role {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You do not have permission"})
			c.Abort()
			return 
		}
	}
}


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
		// Bỏ qua routes không cần auth
		if c.FullPath() == "/auth/users/login" || c.FullPath() == "/auth/users/register" {
			c.Next()
			return
		}

		var tokenString string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				tokenString = authHeader
			}
			logger.Debug("Found token in Authorization header")
		}

		// If not in header, try cookie
		if tokenString == "" {
			var err error
			tokenString, err = c.Cookie("auth_token")
			if err == nil {
				logger.Debug("Found token in cookie")
			} else {
				logger.DebugE("Failed to get token from cookie", err)
			}
		}

		// If still no token, return error
		if tokenString == "" {
			logger.Debug("No token found in header or cookie")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Validate token
		claims, msg := helper.ValidateToken(tokenString)
		if msg != "" {
			logger.Debug("Token validation failed", logger.Str("error", msg))

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

						// Cập nhật header Authorization nếu client sử dụng Bearer token
						if authHeader != "" {
							c.Request.Header.Set("Authorization", "Bearer "+newToken)
						}

						// Tiếp tục với token mới
						tokenString = newToken
						claims, msg = helper.ValidateToken(newToken)
						
						if msg != "" {
							logger.Debug("Token validation failed after refresh", logger.Str("error", msg))
							c.JSON(http.StatusUnauthorized, gin.H{"error": "Your session has expired, please log in again"})
							c.Abort()
							return
						}
						
						logger.DebugE("Token refreshed successfully", nil, logger.Str("email", claims.Email))
					} else {
						logger.DebugE("Failed to refresh token", nil, logger.Str("error", refreshMsg))
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


func RBACMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if  !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error" : "Unauthorized"})
			return 
		}
		u := user.(models.User)
		for _, role := range allowedRoles {
			if u.Role != nil && *u.Role == role {
				c.Next()
				return 
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbiden"})
	}
}