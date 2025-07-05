package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	// "os"
	"time"

	// "github.com/Dattt2k2/golang-project/api-gateway/middleware"
	"api-gateway/logger"
	"api-gateway/middleware"
	"github.com/gin-gonic/gin"
	// "google.golang.org/grpc/admin"
	// "golang.org/x/text/transform"
)

func ForwardImageRequest(c *gin.Context, serviceURL string) {
	logger.InfoE("ForwardImageRequest starting for: %s", nil, logger.Str("serviceURL", serviceURL))

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	req, err := http.NewRequest("GET", serviceURL, nil)
	if err != nil {
		logger.Err("Error creating request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
		return
	}

	logger.InfoE("Sending request to: %s", nil, logger.Str("serviceURL", serviceURL))
	resp, err := client.Do(req)
	if err != nil {
		logger.Err("Error sending request", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
		return
	}
	defer resp.Body.Close()

	logger.InfoE("Response status code: %d", nil, logger.Int("statusCode", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		logger.Err("Error response from service", err, logger.Int("statusCode", resp.StatusCode))

		// Read the response body to log the error
		errorBody, _ := io.ReadAll(resp.Body)
		logger.Err("Error body: %s", err, logger.Str("errorBody", string(errorBody)))

		c.JSON(resp.StatusCode, gin.H{"error": "Image not found"})
		return
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		// Try to determine from filename
		filename := c.Param("filename")
		ext := strings.ToLower(filepath.Ext(filename))

		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".png":
			contentType = "image/png"
		case ".gif":
			contentType = "image/gif"
		case ".webp":
			contentType = "image/webp"
		default:
			contentType = "application/octet-stream"
		}

		logger.Err("No content type in response, using: %s", err, logger.Str("contentType", contentType))
	} else {
		logger.Info("Content type from response", logger.Str("contentType", contentType))
	}

	// Create a new response body since we've consumed the original
	logger.Info("Streaming image to client", logger.Str("contentType", contentType))
	c.Header("Content-Type", contentType)
	c.Status(http.StatusOK)
	io.Copy(c.Writer, resp.Body)
}

func GetImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.Param("filename")
		logger.Info("API Gateway GetImage", logger.Str("filename", filename))

		// Try multiple service URLs
		serviceURLs := []string{
			"http://product-service:8082/images/" + filename,
			"http://product-service:8082/static/product-service/uploads/images/" + filename,
			"http://product-service:8082/static/uploads/images/" + filename,
			"http://product-service:8082/uploads/images/" + filename,
		}

		var resp *http.Response
		var err error

		client := &http.Client{Timeout: time.Second * 30}

		// Try each URL
		for _, serviceURL := range serviceURLs {
			logger.Info("Trying URL", logger.Str("serviceURL", serviceURL))

			req, err := http.NewRequest("GET", serviceURL, nil)
			if err != nil {
				logger.Err("Error creating request", err, logger.Str("serviceURL", serviceURL))
				continue
			}

			resp, err = client.Do(req)
			if err != nil {
				logger.Err("Error accessing URL", err, logger.Str("serviceURL", serviceURL))
				continue
			}

			if resp.StatusCode == http.StatusOK {
				logger.Info("Successfully found image", logger.Str("serviceURL", serviceURL))
				break
			}

			resp.Body.Close()
		}

		if resp == nil || resp.StatusCode != http.StatusOK {
			logger.Err("Failed to find image at any URL", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
			return
		}
		defer resp.Body.Close()

		// Get content type
		contentType := resp.Header.Get("Content-Type")
		if contentType == "" {
			ext := strings.ToLower(filepath.Ext(filename))
			switch ext {
			case ".jpg", ".jpeg":
				contentType = "image/jpeg"
			case ".png":
				contentType = "image/png"
			case ".gif":
				contentType = "image/gif"
			default:
				contentType = "application/octet-stream"
			}
		}

		// Read the entire image into memory
		imageData, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Err("Error reading image data", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
			return
		}

		if len(imageData) == 0 {
			logger.Err("Empty image data received", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Empty image data"})
			return
		}
		logger.Info("Successfully read image data", logger.Int("length", len(imageData)))
		c.Header("Content-Type", contentType)
		c.Data(http.StatusOK, contentType, imageData)
	}
}

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
func ForwardRequestToService(c *gin.Context, serviceURL string, method string, contentType string) {
	logger.InfoE("ForwardRequestToService starting for: %s", nil, logger.Str("serviceURL", serviceURL))

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

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	if method == "GET" {
		// Handle GET request with query params
		reqURL, _ := url.Parse(serviceURL)
		reqURL.RawQuery = c.Request.URL.RawQuery

		req, err := http.NewRequest(method, reqURL.String(), nil)
		if err != nil {
			logger.Err("Error creating GET request", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
			return
		}

		// Set headers
		req.Header.Set("Content-Type", contentType)
		req.Header.Set("email", email.(string))
		req.Header.Set("user_type", role.(string))
		req.Header.Set("user_id", uid.(string))

		resp, err := client.Do(req)
		if err != nil {
			logger.Err("Error in GET request", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
			return
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Err("Error reading GET response", err)
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
	if contentType == "multipart/form-data" {
		err := c.Request.ParseMultipartForm(10 << 20)
		if err != nil {
			logger.Err("Error parsing multipart form", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing multipart form"})
			return
		}

		// Create a new buffer to store the multipart form data
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Copy all form fields
		for key, values := range c.Request.MultipartForm.Value {
			for _, value := range values {
				err := writer.WriteField(key, value)
				if err != nil {
					logger.Err("Error writing field", err, logger.Str("key", key))
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
					return
				}
			}
		}

		// Copy the file
		if file, header, err := c.Request.FormFile("image"); err == nil {
			part, err := writer.CreateFormFile("image", header.Filename)
			if err != nil {
				logger.Err("Error creating form file", err, logger.Str("filename", header.Filename))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
				return
			}

			if _, err := io.Copy(part, file); err != nil {
				logger.Err("Error copying file", err, logger.Str("filename", header.Filename))
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
				return
			}
			file.Close()
		}

		// Close the multipart writer
		writer.Close()

		// Create new request
		req, err := http.NewRequest(method, serviceURL, body)
		if err != nil {
			logger.Err("Error creating multipart request", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
			return
		}

		// Set headers
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("email", email.(string))
		req.Header.Set("user_type", role.(string))
		req.Header.Set("user_id", uid.(string))

		// Send request
		resp, err := client.Do(req)
		if err != nil {
			logger.Err("Error in multipart request", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
			return
		}
		defer resp.Body.Close()

		// Read and forward response
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Err("Error reading multipart response", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response"})
			return
		}

		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), bodyBytes)
	} else {
		// Xử lý các request không phải multipart form như cũ
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		req, err := http.NewRequest(method, serviceURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			logger.Err("Error creating request", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
			return
		}

		req.Header.Set("Content-Type", contentType)
		req.Header.Set("email", email.(string))
		req.Header.Set("user_type", role.(string))
		req.Header.Set("user_id", uid.(string))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Err("Error in request", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
			return
		}
		defer resp.Body.Close()

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	}
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

		// auth.POST("/login", func(c *gin.Context) {
		// 	bodyBytes, _ := io.ReadAll(c.Request.Body)
		// 	req, _ := http.NewRequest("POST", "http://auth-service:8081/auth/users/login", bytes.NewReader(bodyBytes))
		// 	req.Header.Set("Content-Type", "application/json")
		// 	req.Header.Set("X-Device-Id", c.GetString("device_id"))
		// 	req.Header.Set("User-Agent", c.GetString("user_agent"))
		// 	req.Header.Set("X-Platform", c.GetString("platform"))

		// 	resp, err := client.Do(req)
		// 	if err != nil {
		// 		logger.Err("Error creating new request", err)
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
		// 		return
		// 	}
		// 	logger.Info("Auth response:", logger.Int("statusCode", resp.StatusCode))
		// 	defer resp.Body.Close()

		// 	responseBytes, _ := io.ReadAll(resp.Body)
		// 	logger.Info("Auth response body:", logger.Str("responseBody", string(responseBytes)))

		// 	// Nếu status code không phải 200, forward trực tiếp response lỗi
		// 	if resp.StatusCode != http.StatusOK {
		// 		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), responseBytes)
		// 		return
		// 	}

		// 	var loginResponse struct {
		// 		Email        string `json:"email"`
		// 		User_type    string `json:"user_type"`
		// 		Uid          string `json:"user_id"`
		// 		Token        string `json:"token"`
		// 		RefreshToken string `json:"refresh_token"`
		// 	}
		// 	if err := json.Unmarshal(responseBytes, &loginResponse); err != nil {
		// 		logger.Err("Failed to parse auth response", err, logger.Str("responseBody", string(responseBytes)))
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse auth response"})
		// 		return
		// 	}

		// 	logger.Info("After unmarshalling:", logger.Str("email", loginResponse.Email), logger.Str("user_type", loginResponse.User_type), logger.Str("uid", loginResponse.Uid), logger.Str("token", loginResponse.Token))

		// 	// Lưu thông tin vào context
		// 	c.Set("email", loginResponse.Email)
		// 	c.Set("role", loginResponse.User_type)
		// 	c.Set("uid", loginResponse.Uid)

		// 	a, _ := c.Get("email")
		// 	logger.Info("Email from context:", logger.Str("email", a.(string)))

		// 	c.SetCookie(
		// 		"auth_token",
		// 		loginResponse.Token,
		// 		60*60*24*7, // 7 ngày
		// 		"/",
		// 		"",
		// 		c.Request.TLS != nil, // secure nếu là HTTPS
		// 		true,                 // httpOnly
		// 	)

		// 	// Lưu refresh_token vào cookie
		// 	c.SetCookie(
		// 		"refresh_token",
		// 		loginResponse.RefreshToken,
		// 		60*60*24*30, // 30 ngày
		// 		"/",
		// 		"",
		// 		c.Request.TLS != nil,
		// 		true,
		// 	)

		// 	logger.Info("Set auth_token and refresh_token cookies")

		// 	// Gửi phản hồi trở lại client (bao gồm token)
		// 	c.JSON(resp.StatusCode, gin.H{
		// 		"message":       "Login successful",
		// 		"email":         loginResponse.Email,
		// 		"role":          loginResponse.User_type,
		// 		"uid":           loginResponse.Uid,
		// 		"token":         loginResponse.Token,
		// 		"refresh_token": loginResponse.RefreshToken,
		// 	})
		// })
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
			logger.Info("Auth response:", logger.Int("statusCode", resp.StatusCode))
			logger.Info("Auth response body:", logger.Str("responseBody", string(responseBytes)))
			logger.Info("Auth response content-type:", logger.Str("contentType", resp.Header.Get("Content-Type")))

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
				Token        string `json:"token"`
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

			c.SetCookie("auth_token", loginResponse.Token, 60*60*24*7, "/", "", c.Request.TLS != nil, true)
			c.SetCookie("refresh_token", loginResponse.RefreshToken, 60*60*24*30, "/", "", c.Request.TLS != nil, true)

			logger.Info("Login successful", logger.Str("uid", loginResponse.Uid))

			c.JSON(http.StatusOK, gin.H{
				"message":       "Login successful",
				"email":         loginResponse.Email,
				"role":          loginResponse.User_type,
				"uid":           loginResponse.Uid,
				"token":         loginResponse.Token,
				"refresh_token": loginResponse.RefreshToken,
			})
		})

	}
	router.GET("/images/:filename", GetImage())
	// In your SetupRouter function
	router.GET("/static-images/:filename", func(c *gin.Context) {
		filename := c.Param("filename")
		logger.Info("Forwarding to static-images:", logger.Str("filename", filename))
		ForwardImageRequest(c, "http://product-service:8082/static-images/"+filename)
	})
	// Protected routes - cần auth
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		userGroup := protected.Group("/user")
		userGroup.Use(middleware.RequireUserRole("USER"))
		{
			// Usre routes
			protected.GET("/user/get", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/user", "GET", "application/json")
			})
			protected.POST("/users/change-password", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/users/change-password", "POST", "application/json")
			})
			protected.POST("/users/logout", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/users/logout", "GET", "application/json")
			})
			protected.POST("/users/logout-all", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/users/logout-all", "GET", "application/json")
			})

			// Cart routes
			protected.POST("/cart/add/:id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://cart-service:8083/cart/add/"+c.Param("id"), "POST", "application/json")
			})
			protected.GET("/cart/get/", func(c *gin.Context) {
				ForwardRequestToService(c, "http://cart-service:8083/cart/user/get/", "GET", "application/json")
			})
			protected.DELETE("/cart/delete/:id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://cart-service:8083/cart/delete/"+c.Param("id"), "DELETE", "application/json")
			})

			// Order routes
			protected.POST("/order/cart", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/order/cart", "POST", "application/json")
			})
			protected.POST("/order/direct", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/order/direct", "POST", "application/json")
			})
			protected.GET("/order", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/order/user", "GET", "application/json")
			})
			protected.POST("/user/order/cancel/:order_id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/user/order/cancel/"+c.Param("order_id"), "POST", "application/json")
			})
		}
		adminGroup := protected.Group("/admin")
		adminGroup.Use(middleware.RequireUserRole("SELLER"))
		{
			// User routes
			protected.GET("/admin/get-users", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/admin/get-user", "GET", "application/json")
			})

			protected.POST("/admin/change-password/", func(c *gin.Context) {
				ForwardRequestToService(c, "http://auth-service:8081/admin/change-password", "POST", "application/json")
			})

			// Product routes
			protected.POST("/products/add", func(c *gin.Context) {
				ForwardRequestToService(c, "http://product-service:8082/products/add", "POST", "multipart/form-data")
			})

			protected.PUT("/products/edit/:id", func(c *gin.Context) {
				ForwardRequestToService(c, "http://product-service:8082/products/edit/"+c.Param("id"), "PUT", "multipart/form-data")
			})
			protected.GET("/products/images/:filename", func(ctx *gin.Context) {
				ForwardRequestToService(ctx, "http://product-service:8082/images/"+ctx.Param("filename"), "GET", "image/png")
			})

			// Cart routes
			protected.GET("/admin/orders", func(c *gin.Context) {
				ForwardRequestToService(c, "http://order-service:8084/admin/orders", "GET", "application/json")
			})
		}

		// Product routes
		protected.GET("/products/get", func(c *gin.Context) {
			ForwardRequestToService(c, "http://product-service:8082/products/get", "GET", "application/json")
		})
		protected.GET("/products/search", func(c *gin.Context) {
			ForwardRequestToService(c, "http://product-service:8082/products/search?name"+c.Query("name"), "GET", "application/json")
		})

		// Cart routes
		protected.POST("/cart", func(c *gin.Context) {
			ForwardRequestToService(c, "http://cart-service:8083/cart", "POST", "application/json")
		})
		protected.GET("/cart/get", func(c *gin.Context) {
			ForwardRequestToService(c, "http://cart-service:8083/cart/get", "GET", "application/json")
		})

		protected.DELETE("/products/delete/:id", func(c *gin.Context) {
			ForwardRequestToService(c, "http://product-service:8082/products/delete/"+c.Param("id"), "DELETE", "application/json")
		})

		// Search routes
		protected.GET("/search", func(c *gin.Context) {
			ForwardRequestToService(c, "http://search-service:8086/search?name="+c.Query("name"), "GET", "application/json")
		})
		protected.GET("/advanced-search", func(c *gin.Context) {
			ForwardRequestToService(c, "http://search-service:8086/advanced-search?name="+c.Query("name")+"&category="+c.Query("category")+"&brand="+c.Query("brand"), "GET", "application/json")
		})
	}
}
