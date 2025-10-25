package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	// "net/url"
	"strings"

	// "os"
	"time"

	// "github.com/Dattt2k2/golang-project/api-gateway/middleware"
	"api-gateway/logger"
	"api-gateway/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	// "google.golang.org/grpc/admin"
	// "golang.org/x/text/transform"
)

func transformProductResponse(c *gin.Context, responseBody []byte) ([]byte, error) {
	var response map[string]interface{}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return responseBody, err
	}

	// Handle product list responses
	if data, ok := response["data"].([]interface{}); ok {
		for i, item := range data {
			if product, ok := item.(map[string]interface{}); ok {
				if imagePath, ok := product["image_path"].(string); ok {
					// Extract the filename from the path
					parts := strings.Split(imagePath, "/")
					filename := parts[len(parts)-1]

					// Replace with public URL (notice /images/ path)
					product["image_url"] = fmt.Sprintf("http://%s/images/%s", c.Request.Host, filename)
					data[i] = product
				}
			}
		}
		response["data"] = data
	}

	return json.Marshal(response)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Cho phép tất cả origin (production nên cấu hình cẩn thận)
	},
}

func ForwardWebSocketToService(c *gin.Context, serviceURL string) {
	// Lấy thông tin user từ context
	userID, exists := c.Get("uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Err("Failed to upgrade to WebSocket", err)
		return
	}
	defer conn.Close()

	// Connect to backend service WebSocket
	serviceURL = strings.Replace(serviceURL, "ws://", "ws://", 1)
	serviceURL += "?user_id=" + userID.(string)

	backendConn, _, err := websocket.DefaultDialer.Dial(serviceURL, nil)
	if err != nil {
		logger.Err("Failed to connect to backend WebSocket", err)
		return
	}
	defer backendConn.Close()

	// Proxy messages bidirectionally
	go func() {
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			backendConn.WriteMessage(messageType, message)
		}
	}()

	for {
		messageType, message, err := backendConn.ReadMessage()
		if err != nil {
			break
		}
		conn.WriteMessage(messageType, message)
	}
}

var productIDRe = regexp.MustCompile(`/products/[^/]+$`)
var reviewIDRe = regexp.MustCompile(`/v1/products/[^/]+$`)

func ForwardRequestToService(c *gin.Context, serviceURL string, method string, contentType string) {
	// Handle public routes without auth
	if strings.HasSuffix(serviceURL, "/products/get") || strings.HasSuffix(serviceURL, "/search") || strings.HasSuffix(serviceURL, "/advanced-search") || productIDRe.MatchString(serviceURL) || reviewIDRe.MatchString(serviceURL) {
		client := &http.Client{Timeout: time.Second * 30}
		req, err := http.NewRequest(method, serviceURL, nil)
		if err != nil {
			logger.Err("Error creating request", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
			return
		}

		for k, v := range c.Request.Header {
			for _, vv := range v {
				req.Header.Add(k, vv)
			}
		}

		req.Header.Set("Content-Type", contentType)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Err("Error in request", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Err("Error reading response", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response"})
			return
		}

		if strings.Contains(serviceURL, "/products/get") {
			transformedBody, err := transformProductResponse(c, bodyBytes)
			if err == nil {
				bodyBytes = transformedBody
			}
		}

		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), bodyBytes)
		return
	}

	// Get user info from context
	email, exist := c.Get("email")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email not found in context"})
		return
	}

	role, exist := c.Get("role")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role not found in context"})
		return
	}

	uid, exist := c.Get("uid")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "UID not found in context"})
		return
	}

	// Convert to string
	uidStr := fmt.Sprint(uid)
	emailStr := fmt.Sprint(email)
	roleStr := fmt.Sprint(role)

	client := &http.Client{Timeout: time.Second * 30} // Đọc body một lần duy nhất
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
	}

	// Xử lý đặc biệt cho presigned URL - giữ nguyên format array
	if strings.Contains(serviceURL, "/upload/presigned-url") {
		req, err := http.NewRequest(method, serviceURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			logger.Err("Error creating request", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", uidStr)
		req.Header.Set("X-Email", emailStr)
		req.Header.Set("X-Role", roleStr)

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}

		resp, err := client.Do(req)
		if err != nil {
			logger.Err("Error in request", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
			return
		}
		defer resp.Body.Close()

		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Err("Error reading response", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response"})
			return
		}

		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBytes)
		return
	}

	// Xử lý bình thường cho các endpoint khác
	var requestBody map[string]interface{}
	if len(bodyBytes) > 0 {
		json.Unmarshal(bodyBytes, &requestBody)
	}

	if requestBody == nil {
		requestBody = make(map[string]interface{})
	}

	newBodyBytes, _ := json.Marshal(requestBody)

	req, err := http.NewRequest(method, serviceURL, bytes.NewBuffer(newBodyBytes))
	if err != nil {
		logger.Err("Error creating request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", uidStr)
	req.Header.Set("X-Email", emailStr)
	req.Header.Set("X-Role", roleStr)

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Err("Error in request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
		return
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Err("Error reading response", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response"})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBytes)
}

func SetupRouter(router *gin.Engine) {
	var client = &http.Client{}

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.DeviceInfoMiddleware())

	// Public routes - không cần auth
	auth := router.Group("/auth/users")
	{

		auth.POST("/register", func(c *gin.Context) {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			req, _ := http.NewRequest("POST", "http://auth-service:8081/auth/users/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))
			req.Header.Set("X-Device-Id", c.GetString("device_id"))
			req.Header.Set("User-Agent", c.GetString("user_agent"))
			req.Header.Set("X-Platform", c.GetString("platform"))

			resp, err := client.Do(req)
			if err != nil {
				logger.Err("Error creating new request", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
				return
			}
			defer resp.Body.Close()

			c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
		})

		auth.POST("/login", func(c *gin.Context) {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
				return
			}

			fmt.Println("[DEBUG] Body:", string(bodyBytes))
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			req, err := http.NewRequest("POST", "http://auth-service:8081/auth/users/login", bytes.NewReader(bodyBytes))
			fmt.Println("[DEBUG] Sending request to auth-service:8081/auth/users/login")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Device-Id", c.GetString("device_id"))
			req.Header.Set("User-Agent", c.GetString("user_agent"))
			req.Header.Set("X-Platform", c.GetString("platform"))

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				logger.Err("Error sending request to auth service", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
				return
			}
			defer resp.Body.Close()

			responseBytes, _ := io.ReadAll(resp.Body)

			if resp.StatusCode != http.StatusOK {
				c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBytes)
				return
			}

			// Check if response is empty
			if len(responseBytes) == 0 {
				logger.Err("Empty response from auth service", nil)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Empty response from auth service"})
				return
			}

			// Check if response is JSON
			if !json.Valid(responseBytes) {
				logger.Err("Invalid JSON response from auth service", nil, logger.Str("responseBody", string(responseBytes)))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid JSON response from auth service"})
				return
			}

			var loginResponse struct {
				Email        string `json:"email"`
				User_type    string `json:"user_type"`
				Uid          string `json:"user_id"`
				Token        string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			}
			if err := json.Unmarshal(responseBytes, &loginResponse); err != nil {
				logger.Err("Failed to parse auth response", err, logger.Str("responseBody", string(responseBytes)))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse auth response", "details": err.Error()})
				return
			}

			c.Set("email", loginResponse.Email)
			c.Set("role", loginResponse.User_type)
			c.Set("uid", loginResponse.Uid)

			// c.SetCookie("auth_token", loginResponse.Token, 60*60*24*7, "/", "", c.Request.TLS != nil, true)
			// c.SetCookie("refresh_token", loginResponse.RefreshToken, 60*60*24*30, "/", "", c.Request.TLS != nil, true)


			c.JSON(http.StatusOK, gin.H{
				"message":       "Login successful",
				"email":         loginResponse.Email,
				"role":          loginResponse.User_type,
				"uid":           loginResponse.Uid,
				"access_token":  loginResponse.Token,
				"refresh_token": loginResponse.RefreshToken,
			})
		})

		auth.POST("/refresh-token", func(c *gin.Context) {
			var reqBody struct {
				RefreshToken string `json:"refresh_token"`
			}
			if err := c.ShouldBindJSON(&reqBody); err != nil || reqBody.RefreshToken == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Missing refresh_token in body"})
				return
			}

			jsonBody, _ := json.Marshal(reqBody)
			req, err := http.NewRequest("POST", "http://auth-service:8081/auth/refresh-token", bytes.NewReader(jsonBody))
			if err != nil {
				logger.Err("Error creating refresh token request", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
				return
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				logger.Err("Error sending refresh token request", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
				return
			}
			defer resp.Body.Close()

			bodyBytes, _ := io.ReadAll(resp.Body)
			c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), bodyBytes)
		})

		auth.POST("/verify-otp", func(c *gin.Context) {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			req, _ := http.NewRequest("POST", "http://auth-service:8081/auth/verify-otp", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Device-Id", c.GetString("device_id"))
			req.Header.Set("User-Agent", c.GetString("user_agent"))
			req.Header.Set("X-Platform", c.GetString("platform"))

			resp, err := client.Do(req)
			if err != nil {
				logger.Err("Error creating new request", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
				return
			}
			defer resp.Body.Close()

			c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
		})

		auth.POST("/resend-otp", func(c *gin.Context) {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			req, _ := http.NewRequest("POST", "http://auth-service:8081/auth/resend-otp", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Device-Id", c.GetString("device_id"))
			req.Header.Set("User-Agent", c.GetString("user_agent"))
			req.Header.Set("X-Platform", c.GetString("platform"))

			resp, err := client.Do(req)
			if err != nil {
				logger.Err("Error creating new request", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
				return
			}
			defer resp.Body.Close()

			c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
		})
	}

	publicRoutes := router.Group("/api")
	{
		// Product routes
		publicRoutes.GET("/products/get", func(c *gin.Context) {
			ForwardRequestToService(c, "http://product-service:8082/products/get", "GET", "application/json")
		})
		publicRoutes.GET("/products/search", func(c *gin.Context) {
			ForwardRequestToService(c, "http://product-service:8082/products/search?name="+c.Query("name"), "GET", "application/json")
		})
		publicRoutes.GET("/products-info/:id", func(c *gin.Context) {
			ForwardRequestToService(c, "http://product-service:8082/products/"+c.Param("id"), "GET", "application/json")
		})
		publicRoutes.GET("/search", func(c *gin.Context) {
			ForwardRequestToService(c, "http://search-service:8086/search?query="+c.Query("query"), "GET", "application/json")
		})
		publicRoutes.GET("/advanced-search", func(c *gin.Context) {
			ForwardRequestToService(c, "http://search-service:8086/advanced-search?query="+c.Query("query")+"&category="+c.Query("category")+"&brand="+c.Query("brand"), "GET", "application/json")
		})
		publicRoutes.GET("/products/category/:category", func(c *gin.Context) {
			ForwardRequestToService(c, "http://product-service:8082/products/category/"+c.Param("category"), "GET", "application/json")
		})
	}

	// Protected routes - cần auth
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		userGroup := protected.Group("/user")
		userGroup.Use(middleware.RequireUserRole("USER"))
		userGroup.Use(middleware.RBACMiddleware("USER"))
		{
			// User routes
			userGroup.GET("/user/get", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/user", "GET", "application/json")
			})
			userGroup.POST("/users/change-password", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/users/change-password", "POST", "application/json")
			})
			userGroup.POST("/users/logout", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/users/logout", "GET", "application/json")
			})
			userGroup.POST("/users/logout-all", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/users/logout-all", "GET", "application/json")
			})

			userGroup.GET("/users", func(c *gin.Context) {
				ForwardRequestToService(c, "http://user-service:8085/users", "GET", "application/json")
			})

			userGroup.PUT("/users/update", func(c *gin.Context) {
				ForwardRequestToService(c, "http://user-service:8085/users/update", "PUT", "application/json")
			})

			userGroup.DELETE("/users/delete", func(c *gin.Context) {
				ForwardRequestToService(c, "http://user-service:8085/users/delete", "DELETE", "application/json")
			})

			// Cart routes
			userGroup.POST("/cart/add/:id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://cart-service:8083/cart/add/"+c.Param("id"), "POST", "application/json")
			})
			userGroup.GET("/cart/get", func(c *gin.Context) {
				ForwardRequestToService(c, "http://cart-service:8083/cart/user/get/", "GET", "application/json")
			})
			userGroup.DELETE("/cart/delete/:id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://cart-service:8083/cart/delete/"+c.Param("id"), "DELETE", "application/json")
			})

			// Order routes
			userGroup.POST("/order/cart", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/order/cart", "POST", "application/json")
			})
			userGroup.POST("/order/direct", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/order/direct", "POST", "application/json")
			})
			userGroup.GET("/order", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/order/user", "GET", "application/json")
			})
			userGroup.POST("/order/cancel/:order_id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/user/order/cancel/"+c.Param("order_id"), "POST", "application/json")
			})

			// Review routes
			userGroup.POST("/product/review/:product_id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://review-service:8089/v1/products/"+c.Param("product_id")+"/reviews", "POST", "application/json")
			})
			userGroup.GET("/product/review/:product_id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://review-service:8089/v1/products/"+c.Param("product_id")+"/reviews", "GET", "application/json")
			})
			userGroup.POST("/upload/presigned-url", func(c *gin.Context) {
				ForwardRequestToService(c, "http://product-service:8082/upload/presigned-url", "POST", "application/json")
			})
		}
		sellerGroup := protected.Group("/seller")
		// sellerGroup.Use(middleware.RBACMiddleware("SELLER"))
		sellerGroup.Use(middleware.RequireUserRole("USER"))
		{
			// User routes
			sellerGroup.GET("/get-users", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/admin/get-user", "GET", "application/json")
			})

			sellerGroup.POST("/change-password/", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/admin/change-password", "POST", "application/json")
			})

			// Product routes
			sellerGroup.POST("/products", func(c *gin.Context) {
				ForwardRequestToService(c, "http://product-service:8082/products/add", "POST", "application/json")
			})

			sellerGroup.GET("/products", func(c *gin.Context) {
				ForwardRequestToService(c, "http://product-service:8082/products/user", "GET", "application/json")
			})

			sellerGroup.DELETE("/products/delete/:id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://product-service:8082/products/delete/"+c.Param("id"), "DELETE", "application/json")
			})

			sellerGroup.PUT("/products/edit/:id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://product-service:8082/products/edit/"+c.Param("id"), "PUT", "application/json")
			})
			sellerGroup.GET("/products/images/:filename", func(ctx *gin.Context) {
				ForwardRequestToService(ctx, "http://product-service:8082/images/"+ctx.Param("filename"), "GET", "image/png")
			})

			// Cart routes
			sellerGroup.GET("/admin/orders", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/admin/orders", "GET", "application/json")
			})
		}

		adminGroup := protected.Group("/admin")
		adminGroup.Use(middleware.RBACMiddleware("ADMIN"))
		{
			adminGroup.GET("/get-users", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/admin/get-user", "GET", "application/json")
			})
			adminGroup.POST("/change-password/", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/admin/change-password", "POST", "application/json")
			})
			adminGroup.GET("/get-orders", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/admin/orders", "GET", "application/json")
			})
			adminGroup.DELETE("/delete-order/:order_id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/admin/delete-order/"+c.Param("order_id"), "DELETE", "application/json")
			})

		}

		// // Cart routes
		// protected.POST("/cart", func(c *gin.Context) {
		// 	ForwardRequestToService(c, "http://cart-service:8083/cart", "POST", "application/json")
		// })
		// protected.GET("/cart/get", func(c *gin.Context) {
		// 	ForwardRequestToService(c, "http://cart-service:8083/cart/get", "GET", "application/json")
		// })

		// protected.DELETE("/products/delete/:id", func(c *gin.Context) {
		// 	ForwardRequestToService(c, "http://product-service:8082/products/delete/"+c.Param("id"), "DELETE", "application/json")
		// })

		// Search routes

		// WebSocket routes - proxy to auth-service
		protected.GET("/ws", func(c *gin.Context) {
			// Proxy WebSocket connection to auth-service
			ForwardWebSocketToService(c, "ws://auth-service:8081/auth/ws")
		})
	}
}
