package router

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	// "github.com/Dattt2k2/golang-project/api-gateway/middleware"
	"github.com/gin-gonic/gin"
)



// func ForwardRequestToService(c *gin.Context, serviceURL string, method string) {
//     // Debug logging
//     log.Printf("Starting ForwardRequestToService to %s", serviceURL)
    
//     // Get and validate context values with detailed logging
//     email, exists := c.Get("email")
//     log.Printf("Context email: %v, exists: %v", email, exists)
//     if !exists {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Email not found in context"})
//         return
//     }

//     role, exists := c.Get("role")
//     log.Printf("Context role: %v, exists: %v", role, exists)
//     if !exists {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Role not found in context"})
//         return
//     }

//     uid, exists := c.Get("uid")
//     log.Printf("Context uid: %v, exists: %v", uid, exists)
//     if !exists {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "UID not found in context"})
//         return
//     }

//     // Copy request body
//     var bodyBytes []byte
//     if c.Request.Body != nil {
//         bodyBytes, _ = io.ReadAll(c.Request.Body)
//         // Restore the body for subsequent reads
//         c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
//     }

//     // Create new request with detailed error handling
//     req, err := http.NewRequest(method, serviceURL, bytes.NewBuffer(bodyBytes))
//     if err != nil {
//         log.Printf("Error creating request: %v", err)
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
//         return
//     }

//     // Add headers with logging
//     log.Printf("Setting headers - email: %s, role: %s, uid: %s", email, role, uid)
//     req.Header.Set("email", email.(string))
//     req.Header.Set("user_type", role.(string))
//     req.Header.Set("user_id", uid.(string))
//     req.Header.Set("Content-Type", "application/json")

//     // Forward request with logging
//     resp, err := http.DefaultClient.Do(req)
//     if err != nil {
//         log.Printf("Error forwarding request: %v", err)
//         c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
//         return
//     }
//     defer resp.Body.Close()

//     // Debug response
//     log.Printf("Service response status: %d", resp.StatusCode)
    
//     // Forward response
//     c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
// }

func ForwardRequestToService(c *gin.Context, serviceURL string, method string, contentType string) {
    log.Printf("Starting ForwardRequestToService to %s", serviceURL)
    
    // Lấy thông tin từ context
    email, exists := c.Get("email")
    if !exists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Email not found in context"})
        return
    }

    role, exists := c.Get("role")
    if !exists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Role not found in context"})
        return
    }

    uid, exists := c.Get("uid")
    if !exists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "UID not found in context"})
        return
    }

    // Sao chép body từ request ban đầu
    var bodyBytes []byte
    if c.Request.Body != nil {
        bodyBytes, _ = io.ReadAll(c.Request.Body)
        c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
    }

    // Tạo request mới
    req, err := http.NewRequest(method, serviceURL, bytes.NewBuffer(bodyBytes))
    if err != nil {
        log.Printf("Error creating request: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
        return
    }

    // Thêm các header vào request
    req.Header.Set("email", email.(string))
    req.Header.Set("user_type", role.(string))
    req.Header.Set("user_id", uid.(string))

    // Set Content-Type theo tham số truyền vào
    if contentType != "" {
        req.Header.Set("Content-Type", contentType)
    } else {
        req.Header.Set("Content-Type", "application/json") // Mặc định
    }

    // Gửi request tới service
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Printf("Error forwarding request: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
        return
    }
    defer resp.Body.Close()

    log.Printf("Service response status: %d", resp.StatusCode)
    c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}




func SetupRouter(router *gin.Engine) {
    var client = &http.Client{}

    // Public routes - không cần auth
    auth := router.Group("/auth/users")
    {
        auth.POST("/register", func(c *gin.Context) {
            bodyBytes, _ := io.ReadAll(c.Request.Body)
            req, _ := http.NewRequest("POST", "http://auth-service:8081/users/register", bytes.NewReader(bodyBytes))
            req.Header.Set("Content-Type", "application/json")

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

            // Gửi phản hồi trở lại client (bao gồm token)
            c.JSON(resp.StatusCode, gin.H{
                "email": loginResponse.Email,
                "role":  loginResponse.User_type,
                "uid":   loginResponse.Uid,
                "token": loginResponse.Token,
            })
            })
        }

        // Protected routes - cần auth
        protected := router.Group("/api")
        // protected.Use(middleware.Authenticate())
    {
        protected.POST("/products/add", func(c *gin.Context) {
            ForwardRequestToService(c, "http://product-service:8082/products/add", "POST", "multipart/form-data")
        })

        protected.POST("/cart", func(c *gin.Context) {
            ForwardRequestToService(c, "http://cart-service:8083/cart", "POST", "application/json")
        })

        protected.POST("/orders", func(c *gin.Context) {
            ForwardRequestToService(c, "http://order-service:8084/orders", "POST" , "application/json")
        })
    }
}

