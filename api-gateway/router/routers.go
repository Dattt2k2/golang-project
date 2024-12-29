package router

import (
	"bytes"
	"io"
	"net/http"

	// "github.com/Dattt2k2/golang-project/api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

func ForwardRequestToService(c *gin.Context, serviceURL string){
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	uid, _ := c.Get("uid")

	req, err  := http.NewRequest("GET", serviceURL, nil)
	if err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
		return
	}

	req.Header.Add("email", email.(string))
	req.Header.Add("role", role.(string))
	req.Header.Add("uid", uid.(string))

	resp, err := http.DefaultClient.Do(req)
	if err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error":"Unable to connect to service"})
		return
	}

	defer resp.Body.Close()

	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}

func SetupRouter(router *gin.Engine){

	// Auth Route
	var client = &http.Client{}

	router.POST("/auth/users/register", func(c *gin.Context) {
    	bodyBytes, _ := io.ReadAll(c.Request.Body)
    	req, _ := http.NewRequest("POST", "http://auth-service:8081/auth/users/register", bytes.NewReader(bodyBytes))
    	req.Header.Set("Content-Type", "application/json")

    	resp, err := client.Do(req)
    	if err != nil {
        	c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
        	return
    	}
    	defer resp.Body.Close()

    	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})


	// Login Route
	router.POST("/auth/users/login", func(c *gin.Context) {
    	bodyBytes, _ := io.ReadAll(c.Request.Body)
    	req, _ := http.NewRequest("POST", "http://auth-service:8081/auth/users/login", bytes.NewReader(bodyBytes))
    	req.Header.Set("Content-Type", "application/json")

    	resp, err := client.Do(req)
    	if err != nil {
        	c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
        	return
    	}
    	defer resp.Body.Close()

    	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})

	// Product Route
	router.GET("/products", func(c *gin.Context){
		ForwardRequestToService(c, "http://product-service:8082/products")
	})

	// Cart Route
	router.GET("/cart", func(c *gin.Context){
		ForwardRequestToService(c, "http://cart-service:8083/cartt")
	})

	// Order Route
	router.GET("/orders", func(c *gin.Context){
		ForwardRequestToService(c, "http://order-service:8084/orders")
	})

	

}

// func SetupRouter(router *gin.Engine){

// 	var client = &http.Client{}

// 	publicRoutes := router.Group("/auth/users")
// 	{
// 		publicRoutes.POST("/register", func(c * gin.Context){
// 			bodyBytes, _ := io.ReadAll(c.Request.Body)
// 			req, _ := http.NewRequest("POST", "http://auth-service:8081/auth/users/register", bytes.NewReader(bodyBytes))
// 			req.Header.Set("Content-Type", "application/json")

// 			resp, err := client.Do(req)
// 			if err != nil{
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to conect to auth  service"})
// 				return
// 			}
// 			defer resp.Body.Close()

// 			c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
// 		})

// 		publicRoutes.POST("/login", func(c *gin.Context) {
// 			bodyBytes, _ := io.ReadAll(c.Request.Body)
// 			req, _ := http.NewRequest("POST", "http://auth-service:8081/auth/users/login", bytes.NewReader(bodyBytes))
// 			req.Header.Set("Content-Type", "application/json")

// 			resp, err := client.Do(req)
// 			if err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to auth service"})
// 				return
// 			}
// 			defer resp.Body.Close()

// 			c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
// 		})
// 	}

// 	protectedRoutes := router.Group("/")
// 	protectedRoutes.Use(middleware.Authenticate())
// 	{
// 		protectedRoutes.GET("/products", func(c *gin.Context){
// 			ForwardRequestToService(c, "http://product-service:8082/products")
// 		})

// 		protectedRoutes.GET("/cart", func(c *gin.Context) {
// 			ForwardRequestToService(c, "http://cart-service:8083/cart")
// 		})

// 		protectedRoutes.GET("/orders", func(c *gin.Context) {
// 			ForwardRequestToService(c, "http://order-service:8084/orders")
// 		})
// 	}

// }