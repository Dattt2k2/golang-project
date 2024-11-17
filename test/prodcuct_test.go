package test

import (
	"bytes"
	"context"
	"fmt"

	"github.com/Dattt2k2/golang-project/controllers/sellers"  // Thay đổi theo package của bạn
	// "github.com/Dattt2k2/golang-project/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	// "time"

	"mime/multipart"
	"path/filepath"

	// "github.com/Dattt2k2/golang-project/controllers/admin"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAddProduct(t *testing.T) {
	// Thiết lập MongoDB giả lập (hoặc sử dụng MongoDB thực)
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Connect(context.Background())
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database("testdb")

	// Khởi tạo Gin
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Mô phỏng route
	router.POST("/add-product", controllers.AddProduct(db))

	// Mô phỏng dữ liệu file và request form
	formData := new(bytes.Buffer)
	writer := multipart.NewWriter(formData)

	// Mô phỏng file image
	imagePath := filepath.Join("test_images", "test.jpg")
	file, err := os.Open(imagePath)
	if err != nil {
		t.Fatalf("Failed to open test image: %v", err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("image", "test.jpg")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Copy image file content vào form
	_, err = fmt.Fprintf(part, "test image content")
	if err != nil {
		t.Fatalf("Failed to write image content to form: %v", err)
	}

	// Thêm các field khác vào form
	_ = writer.WriteField("name", "Test Product")
	_ = writer.WriteField("description", "A test product description")
	_ = writer.WriteField("quantity", "10")
	_ = writer.WriteField("price", "99.99")

	// Kết thúc multipart
	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Tạo request HTTP
	req := httptest.NewRequest(http.MethodPost, "/add-product", formData)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("uid", "60a7b527f61e2f001f1f1f1f")  // Thay đổi với ID người dùng hợp lệ

	// Gửi request và nhận response
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Kiểm tra kết quả
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "Product added successfully")
}

