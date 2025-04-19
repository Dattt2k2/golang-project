package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	// "os"
	"time"

	// "github.com/Dattt2k2/golang-project/api-gateway/middleware"
	"github.com/Dattt2k2/golang-project/api-gateway/middleware"
	"github.com/gin-gonic/gin"
	// "golang.org/x/text/transform"
)

func ForwardImageRequest(c *gin.Context, serviceURL string) {
    log.Printf("ForwardImageRequest starting for: %s", serviceURL)
    
    client := &http.Client{
        Timeout: time.Second * 30,
    }
    
    req, err := http.NewRequest("GET", serviceURL, nil)
    if err != nil {
        log.Printf("Error creating request: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
        return
    }
    
    log.Printf("Sending request to: %s", serviceURL)
    resp, err := client.Do(req)
    if err != nil {
        log.Printf("Error sending request: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
        return
    }
    defer resp.Body.Close()
    
    log.Printf("Response status code: %d", resp.StatusCode)
    
    if resp.StatusCode != http.StatusOK {
        log.Printf("Error response: %d", resp.StatusCode)
        
        // Read the response body to log the error
        errorBody, _ := io.ReadAll(resp.Body)
        log.Printf("Error body: %s", string(errorBody))
        
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
        
        log.Printf("No content type in response, using: %s", contentType)
    } else {
        log.Printf("Content type from response: %s", contentType)
    }
    
    // Create a new response body since we've consumed the original
    log.Printf("Streaming image to client with content type: %s", contentType)
    c.Header("Content-Type", contentType)
    c.Status(http.StatusOK)
    io.Copy(c.Writer, resp.Body)
}

func GetImage() gin.HandlerFunc {
    return func(c *gin.Context) {
        filename := c.Param("filename")
        log.Printf("API Gateway GetImage: %s", filename)
        
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
            log.Printf("Trying URL: %s", serviceURL)
            
            req, err := http.NewRequest("GET", serviceURL, nil)
            if err != nil {
                log.Printf("Error creating request for %s: %v", serviceURL, err)
                continue
            }
            
            resp, err = client.Do(req)
            if err != nil {
                log.Printf("Error accessing %s: %v", serviceURL, err)
                continue
            }
            
            if resp.StatusCode == http.StatusOK {
                log.Printf("Successfully found image at: %s", serviceURL)
                break
            }
            
            resp.Body.Close()
        }
        
        if resp == nil || resp.StatusCode != http.StatusOK {
            log.Printf("Failed to find image at any URL")
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
            log.Printf("Error reading image data: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
            return
        }
        
        if len(imageData) == 0 {
            log.Printf("Empty image data received")
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Empty image data"})
            return
        }
        
        log.Printf("Successfully read %d bytes of image data", len(imageData))
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
    log.Printf("Starting ForwardRequestToService to %s with method %s", serviceURL, method)

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
            log.Printf("Error creating GET request: %v", err)
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
            log.Printf("Error in GET request: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
            return
        }
        defer resp.Body.Close()
        
        bodyBytes, err := io.ReadAll(resp.Body)
        if err != nil {
            log.Printf("Error reading GET response: %v", err)
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
            log.Printf("Error parsing original multipart form: %v", err)
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
                    log.Printf("Error writing field %s: %v", key, err)
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
                    return
                }
            }
        }

        // Copy the file
        if file, header, err := c.Request.FormFile("image"); err == nil {
            part, err := writer.CreateFormFile("image", header.Filename)
            if err != nil {
                log.Printf("Error creating form file: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating request"})
                return
            }
            
            if _, err := io.Copy(part, file); err != nil {
                log.Printf("Error copying file: %v", err)
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
            log.Printf("Error creating request: %v", err)
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
            log.Printf("Error in request: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
            return
        }
        defer resp.Body.Close()

        // Read and forward response
        bodyBytes, err := io.ReadAll(resp.Body)
        if err != nil {
            log.Printf("Error reading response: %v", err)
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
                    log.Printf("Error creating request: %v", err)
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
                    return
                }
        
                req.Header.Set("Content-Type", contentType)
                req.Header.Set("email", email.(string))
                req.Header.Set("user_type", role.(string))
                req.Header.Set("user_id", uid.(string))
        
                resp, err := http.DefaultClient.Do(req)
                if err != nil {
                    log.Printf("Error in request: %v", err)
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
        // auth.POST("/register", func(c *gin.Context) {
        //     bodyBytes, err := io.ReadAll(c.Request.Body)
        //     if err != nil{
        //         log.Printf("Error reading requets body: %v", err)
        //         c.JSON(http.StatusBadRequest, gin.H{"error": "Error reading requets"})
        //         return
        //     }

        //     log.Printf("Body bytes: %v", bodyBytes)
        //     req, err := http.NewRequest("POST", "http://auth-service:8081/users/register", bytes.NewBuffer(bodyBytes))
        //     if err != nil{
        //         log.Println("Error creating new request:", err)
        //         c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create new request"})
        //         return
        //     }
        //     req.Header.Set("Content-Type", "application/json")
        //     req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))

        //     client := &http.Client{
        //         Timeout: 10 * time.Second,
        //     }

        //     resp, err := client.Do(req)
        //     if err != nil {
		// 		log.Println("Error reading request body:", err)
        //         c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
        //         return
        //     }
        //     respBody, err := io.ReadAll(resp.Body)
        //     if err != nil{
        //         log.Printf("Error reading response body: %v", err)
        //         c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading request body"})
        //         return
        //     }
        //     c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
        //     // defer resp.Body.Close()

        //     // c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
        // })
        auth.POST("/register", func(c *gin.Context) {
            bodyBytes, _ := io.ReadAll(c.Request.Body)
            req, _ := http.NewRequest("POST", "http://auth-service:8081/users/register", bytes.NewReader(bodyBytes))
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyBytes)))
            req.Header.Set("X-Device-Id", c.GetString("device_id"))
            req.Header.Set("User-Agent", c.GetString("user_agent"))
            req.Header.Set("X-Platform", c.GetString("platform"))

            resp, err := client.Do(req)
            if err != nil {
				log.Println("Error reading request body:", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
                return
            }
            defer resp.Body.Close()

            c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
        })

        auth.POST("/login", func(c *gin.Context) {
            bodyBytes, _ := io.ReadAll(c.Request.Body)
            req, _ := http.NewRequest("POST", "http://auth-service:8081/users/login", bytes.NewReader(bodyBytes))
            req.Header.Set("Content-Type", "application/json")
            req.Header.Set("X-Device-Id", c.GetString("device_id"))
            req.Header.Set("User-Agent", c.GetString("user_agent"))
            req.Header.Set("X-Platform", c.GetString("platform"))

            resp, err := client.Do(req)
            if err != nil {
				log.Println("Error creating new request:", err)
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
                return
            }
            log.Println("Auth response:", resp)
            defer resp.Body.Close()

            responseBytes, _ := io.ReadAll(resp.Body)

            log.Println("Auth response:", string(responseBytes))

            var loginResponse struct {
                Email string `json:"email"`
                User_type  string `json:"user_type"`
                Uid   string `json:"user_id"`
                Token string `json:"token"`
            }
            if err := json.Unmarshal(responseBytes, &loginResponse); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse auth response"})
                return
            }

            log.Println("After unmashalling:", loginResponse.Email, loginResponse.User_type, loginResponse.Uid, loginResponse.Token)

            // Lưu thông tin vào context
            c.Set("email", loginResponse.Email)
            c.Set("role", loginResponse.User_type)
            c.Set("uid", loginResponse.Uid)

            a, _ := c.Get("email")
            log.Println("Email from context:", a)

            c.SetCookie(
                "auth_token",
                loginResponse.Token,
                60*60*24*7,
                "/",
                "",
                c.Request.TLS != nil,
                true,
            )

            log.Printf("Set auth_token cokkies")

            // Gửi phản hồi trở lại client (bao gồm token)
            c.JSON(resp.StatusCode, gin.H{
                "email": loginResponse.Email,
                "role":  loginResponse.User_type,
                "uid":   loginResponse.Uid,
                "token": loginResponse.Token,
            })
            })
        }
        router.GET("/images/:filename", GetImage())
        // In your SetupRouter function
        router.GET("/static-images/:filename", func(c *gin.Context) {
            filename := c.Param("filename")
            log.Printf("Forwarding to static-images: %s", filename)
            ForwardImageRequest(c, "http://product-service:8082/static-images/" + filename)
        })
        // Protected routes - cần auth
        protected := router.Group("/api")
        protected.Use(middleware.AuthMiddleware())
    {
        protected.POST("/products/add", func(c *gin.Context) {
            ForwardRequestToService(c, "http://product-service:8082/products/add", "POST", "multipart/form-data")
        })

        protected.POST("/cart", func(c *gin.Context) {
            ForwardRequestToService(c, "http://cart-service:8083/cart", "POST", "application/json")
        })

        // protected.POST("/orders", func(c *gin.Context) {
        //     ForwardRequestToService(c, "http://order-service:8084/orders", "POST" , "application/json")
        // })


        // Product routes
        protected.GET("/products/get", func(c *gin.Context){
            ForwardRequestToService(c, "http://product-service:8082/products/get", "GET", "application/json")
        })
        protected.DELETE("/products/delete/:id", func(c *gin.Context){
            ForwardRequestToService(c, "http://product-service:8082/products/delete/" + c.Param("id"), "DELETE", "application/json")
        })
        protected.PUT("/products/edit/:id", func(c *gin.Context){
            ForwardRequestToService(c, "http://product-service:8082/products/edit/" + c.Param("id"), "PUT", "multipart/form-data")
        })
        protected.GET("/products/search", func(c *gin.Context){
            ForwardRequestToService(c, "http://product-service:8082/products/search?name" + c.Query("name"), "GET", "application/json")
        })
        protected.GET("/products/images/:filename", func(ctx *gin.Context) {
            ForwardRequestToService(ctx, "http://product-service:8082/images/" + ctx.Param("filename"), "GET", "image/png")
        })

        // Cart routes
        protected.POST("/cart/add/:id", func(c *gin.Context){
            ForwardRequestToService(c,"http://cart-service:8083/cart/add/" + c.Param("id"), "POST", "application/json")
        })
        protected.GET("cart/get", func(c *gin.Context) {
            ForwardRequestToService(c, "http://cart-service:8083/cart/get", "GET", "application/json")
        })
        protected.GET("/cart/get/:id", func(c *gin.Context){
            ForwardRequestToService(c, "http://cart-service:8083/cart/get/" + c.Param("id"), "GET", "application/json")
        })
        protected.DELETE("cart/delete/:id", func(c *gin.Context){
            ForwardRequestToService(c, "http://cart-service:8083/cart/delete/" + c.Param("id"), "DELETE", "application/json")
        })


        // Order routes
        protected.POST("order/cart/", func(c *gin.Context){
            ForwardRequestToService(c, "http://order-service:8084/order/cart/", "POST", "application/json")
        })
        protected.POST("order/direct", func(c *gin.Context){
            ForwardRequestToService(c, "http://order-service:8084/order/direct/", "POST", "application/json")
        })
        protected.GET("order/user", func(c *gin.Context){
            ForwardRequestToService(c, "http://order-service:8084/order/user", "GET", "application/json")
        })
        protected.GET("admin/orders", func(c *gin.Context){
            ForwardRequestToService(c, "http://order-service:8084/admin/orders", "GET", "application/json")
        })
    }
}

