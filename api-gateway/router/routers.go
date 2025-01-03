package router

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"

	// "os"
	"time"

	// "github.com/Dattt2k2/golang-project/api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

// func ForwardRequestToService(c *gin.Context, serviceURL string, method string, contentType string) {
//     log.Printf("Starting ForwardRequestToService to %s with method %s", serviceURL, method)

//     email, exist := c.Get("email")
//     if !exist {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Email not found in context"})
//         return
//     }

//     role, exist := c.Get("role")
//     if !exist {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "Role not found in context"})
//         return
//     }

//     uid, exist := c.Get("uid")
//     if !exist {
//         c.JSON(http.StatusBadRequest, gin.H{"error": "UID not found in context"})
//         return
//     }

//     client := &http.Client{
//         Timeout: time.Second * 30,
//     }

//     if method == "GET" {
//         // Handle GET request with query params
//         reqURL, _ := url.Parse(serviceURL)
//         reqURL.RawQuery = c.Request.URL.RawQuery

//         req, err := http.NewRequest(method, reqURL.String(), nil)
//         if err != nil {
//             log.Printf("Error creating GET request: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
//             return
//         }

//         // Set headers
//         req.Header.Set("Content-Type", contentType)
//         req.Header.Set("email", email.(string))
//         req.Header.Set("user_type", role.(string))
//         req.Header.Set("user_id", uid.(string))

//         resp, err := client.Do(req)
//         if err != nil {
//             log.Printf("Error in GET request: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
//             return
//         }
//         defer resp.Body.Close()

//         bodyBytes, err := io.ReadAll(resp.Body)
//         if err != nil {
//             log.Printf("Error reading GET response: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response"})
//             return
//         }

//         c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), bodyBytes)
//         return
//     }

//     if contentType == "multipart/form-data" {
//         log.Printf("Processing multipart form data")
//         if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
//             log.Printf("Error parsing multipart form: %v", err)
//             c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing multipart form"})
//             return
//         }

//         // Tạo pipe và errChan để theo dõi lỗi từ goroutine
//         pr, pw := io.Pipe()
//         errChan := make(chan error)
//         writer := multipart.NewWriter(pw)

//         // Start goroutine để copy dữ liệu
//         go func() {
//             defer writer.Close()
//             defer pw.Close()

//             var copyError error
//             log.Printf("Form values before forwarding: %v", c.Request.Form)

//             // Copy form fields
//             for key, values := range c.Request.Form {
//                 log.Printf("Processing form field - Key: %s, Values: %v", key, values)
//                 for _, value := range values {
//                     if err := writer.WriteField(key, value); err != nil {
//                         copyError = fmt.Errorf("error writing field %s: %v", key, err)
//                         log.Printf("%v", copyError)
//                         break
//                     }
//                 }
//                 if copyError != nil {
//                     break
//                 }
//             }

//             // Copy file nếu không có lỗi từ form fields
//             if copyError == nil {
//                 if file, header, err := c.Request.FormFile("image"); err == nil {
//                     log.Printf("Processing file: %s", header.Filename)
//                     part, err := writer.CreateFormFile("image", header.Filename)
//                     if err != nil {
//                         copyError = fmt.Errorf("error creating form file: %v", err)
//                         log.Printf("%v", copyError)
//                     } else {
//                         if _, err := io.Copy(part, file); err != nil {
//                             copyError = fmt.Errorf("error copying file: %v", err)
//                             log.Printf("%v", copyError)
//                         } else {
//                             log.Printf("File processed successfully")
//                         }
//                     }
//                     file.Close()
//                 }
//             }

//             // Gửi lỗi (nếu có) qua channel
//             errChan <- copyError
//         }()

//         // Tạo request với pipe reader
//         req, err := http.NewRequest(method, serviceURL, pr)
//         if err != nil {
//             log.Printf("Error creating request: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
//             return
//         }

//         // Set headers
//         req.Header.Set("Content-Type", writer.FormDataContentType())
//         req.Header.Set("email", email.(string))
//         req.Header.Set("user_type", role.(string))
//         req.Header.Set("user_id", uid.(string))

//         // Tạo client với timeout
//         client := &http.Client{
//             Timeout: time.Second * 30, // Tăng timeout lên 30 giây
//         }

//         // Tạo channel để nhận response
//         respChan := make(chan *http.Response)
//         reqErrChan := make(chan error)

//         // Gửi request trong goroutine
//         go func() {
//             resp, err := client.Do(req)
//             if err != nil {
//                 reqErrChan <- err
//                 return
//             }
//             respChan <- resp
//         }()

//         // Đợi và xử lý kết quả
//         select {
//         case copyErr := <-errChan:
//             if copyErr != nil {
//                 log.Printf("Error in copy goroutine: %v", copyErr)
//                 c.JSON(http.StatusInternalServerError, gin.H{"error": "Error processing form data"})
//                 return
//             }
//         case reqErr := <-reqErrChan:
//             log.Printf("Error in request: %v", reqErr)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Error connecting to service"})
//             return
//         case resp := <-respChan:
//             defer resp.Body.Close()
//             log.Printf("Request successful, status: %d", resp.StatusCode)

//             // Copy response về client
//             bodyBytes, err := io.ReadAll(resp.Body)
//             if err != nil {
//                 log.Printf("Error reading response body: %v", err)
//                 c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading response"})
//                 return
//             }

//             resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
//             c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
//         case <-time.After(time.Second * 30):
//             log.Printf("Request timeout")
//             c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Request timeout"})
//             return
//         }
//     } else {
//         // Xử lý các request không phải multipart form như cũ
//         var bodyBytes []byte
//         if c.Request.Body != nil {
//             bodyBytes, _ = io.ReadAll(c.Request.Body)
//             c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
//         }

//         req, err := http.NewRequest(method, serviceURL, bytes.NewBuffer(bodyBytes))
//         if err != nil {
//             log.Printf("Error creating request: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to create request"})
//             return
//         }

//         req.Header.Set("Content-Type", contentType)
//         req.Header.Set("email", email.(string))
//         req.Header.Set("user_type", role.(string))
//         req.Header.Set("user_id", uid.(string))

//         resp, err := http.DefaultClient.Do(req)
//         if err != nil {
//             log.Printf("Error in request: %v", err)
//             c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to connect to service"})
//             return
//         }
//         defer resp.Body.Close()

//         c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
//     }
// }

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
    }
}

