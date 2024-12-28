package router

import (
	"net/http"

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
	router.POST("/auth/users/register", func(c *gin.Context){
		resp, err := http.Post("http://localhost:8081/auth/users/register", "application/json", c.Request.Body)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"Unable to connect to auth service"})
			return
		}
		defer resp.Body.Close()

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})

	// Login Route
	router.POST("/auth/users/login", func(c *gin.Context){
		resp, err := http.Post("http://localhost:8081/auth/users/login", "application/json", c.Request.Body)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error":"Unable to connect to auth service"})
			return
		}
		defer resp.Body.Close()

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)

	})

	// Product Route
	router.GET("/products", func(c *gin.Context){
		ForwardRequestToService(c, "http://localhost:8082/products")
	})

	// Cart Route
	router.GET("/cart", func(c *gin.Context){
		ForwardRequestToService(c, "http://localhost:8083/cart")
	})

	// Order Route
	router.GET("/orders", func(c *gin.Context){
		ForwardRequestToService(c, "http://localhost:8084/orders")
	})

}