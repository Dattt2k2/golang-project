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



func ForwardRequestToService(c *gin.Context, serviceURL string, method string) {
    email, exists := c.Get("email")
    if !exists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Email not found in context"})
        return
    }

    role, exists := c.Get("user_type")
    if !exists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Role not found in context"})
        return
    }

    uid, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "UID not found in context"})
        return
    }

    bodyBytes, err := io.ReadAll(c.Request.Body)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    req, err := http.NewRequest(method, serviceURL, bytes.NewReader(bodyBytes))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
        return
    }

    req.Header.Add("email", email.(string))
    req.Header.Add("user_type", role.(string))
    req.Header.Add("user_id", uid.(string))
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
        return
    }

    defer resp.Body.Close()

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
            defer resp.Body.Close()

            responseBytes, _ := io.ReadAll(resp.Body)

            log.Println("Auth response:", string(responseBytes))

            var loginResponse struct {
                Email string `json:"email"`
                Role  string `json:"user_type"`
                Uid   string `json:"user_id"`
                Token string `json:"token"`
            }
            if err := json.Unmarshal(responseBytes, &loginResponse); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse auth response"})
                return
            }

            // Lưu thông tin vào context
            c.Set("email", loginResponse.Email)
            c.Set("role", loginResponse.Role)
            c.Set("uid", loginResponse.Uid)

            // Gửi phản hồi trở lại client (bao gồm token)
            c.JSON(resp.StatusCode, gin.H{
                "email": loginResponse.Email,
                "role":  loginResponse.Role,
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
            ForwardRequestToService(c, "http://product-service:8082/products/add", "POST")
        })

        protected.POST("/cart", func(c *gin.Context) {
            ForwardRequestToService(c, "http://cart-service:8083/cart", "POST")
        })

        protected.POST("/orders", func(c *gin.Context) {
            ForwardRequestToService(c, "http://order-service:8084/orders", "POST")
        })
    }
}

