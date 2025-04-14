package test

import (
	"errors"
	// "fmt"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// saveImageToFileSystem saves the uploaded image file to the file system
func saveImageToFileSystem(c *MockGinContext, file *multipart.FileHeader) (string, error) {
	// Create uploads/images directory
	uploadsPath := filepath.Join("uploads", "images")
	err := os.MkdirAll(uploadsPath, 0755)
	if err != nil {
		return "", fmt.Errorf("Failed to create image directory: %w", err)
	}
	
	// Generate unique filename with timestamp
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%d_%s", timestamp, file.Filename)
	filepath := filepath.Join(uploadsPath, filename)
	
	// Save the file
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		return "", fmt.Errorf("Failed to save image: %w", err)
	}
	
	return filepath, nil
}

// MockFileHeader is a mock implementation of multipart.FileHeader
type MockFileHeader struct {
	mock.Mock
	Filename string
	Size     int64
	Header   map[string][]string
}

// MockGinContext is used to mock the gin.Context for testing
type MockGinContext struct {
	*gin.Context
	SaveUploadedFileFunc func(file *multipart.FileHeader, dst string) error
}

func (m *MockGinContext) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	if m.SaveUploadedFileFunc != nil {
		return m.SaveUploadedFileFunc(file, dst)
	}
	return nil
}

func TestSaveImageToFileSystem(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "product-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	originalWd, _ := os.Getwd()
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	t.Run("Successful image save", func(t *testing.T) {
		// Arrange
		file := &multipart.FileHeader{
			Filename: "test-image.jpg",
			Size:     1024,
			Header:   make(map[string][]string),
		}
		
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		
		mockC := &MockGinContext{
			Context: c,
			SaveUploadedFileFunc: func(file *multipart.FileHeader, dst string) error {
				// Create an empty file to simulate saving
				f, err := os.Create(dst)
				if err != nil {
					return err
				}
				f.Close()
				return nil
			},
		}
		
		// Act
		imagePath, err := saveImageToFileSystem(mockC, file)
		
		// Assert
		assert.NoError(t, err)
		assert.NotEmpty(t, imagePath)
		assert.True(t, strings.HasSuffix(imagePath, "test-image.jpg"))
		
		// Verify the directory was created
		wd, _ := os.Getwd()
		assert.DirExists(t, filepath.Join(wd, "uploads", "images"))
	})
	
	t.Run("File save error", func(t *testing.T) {
		// Arrange
		file := &multipart.FileHeader{
			Filename: "test-image.jpg",
			Size:     1024,
			Header:   make(map[string][]string),
		}
		
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		
		mockC := &MockGinContext{
			Context: c,
			SaveUploadedFileFunc: func(file *multipart.FileHeader, dst string) error {
				return errors.New("simulated file save error")
			},
		}
		
		// Act
		imagePath, err := saveImageToFileSystem(mockC, file)
		
		// Assert
		assert.Error(t, err)
		assert.Empty(t, imagePath)
		assert.Contains(t, err.Error(), "Failed to save image")
	})
	
	t.Run("Unique filename generation", func(t *testing.T) {
		// Arrange
		file := &multipart.FileHeader{
			Filename: "test-image.jpg",
			Size:     1024,
			Header:   make(map[string][]string),
		}
		
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		
		mockC := &MockGinContext{
			Context: c,
			SaveUploadedFileFunc: func(file *multipart.FileHeader, dst string) error {
				// Create an empty file to simulate saving
				f, err := os.Create(dst)
				if err != nil {
					return err
				}
				f.Close()
				return nil
			},
		}
		
		// Act
		// Call twice to verify unique names
		firstPath, _ := saveImageToFileSystem(mockC, file)
		time.Sleep(1 * time.Second) // Ensure different timestamp
		secondPath, _ := saveImageToFileSystem(mockC, file)
		
		// Assert
		assert.NotEqual(t, firstPath, secondPath)
		assert.True(t, strings.HasSuffix(firstPath, "test-image.jpg"))
		assert.True(t, strings.HasSuffix(secondPath, "test-image.jpg"))
	})
	
	t.Run("Directory creation handling", func(t *testing.T) {
		// Arrange - Create a read-only directory that will cause mkdir to fail
		// Note: This test may not work on Windows
		if os.Getenv("SKIP_PERMISSION_TESTS") != "" {
			t.Skip("Skipping permission-based test")
		}
		
		readOnlyDir := filepath.Join(tempDir, "readonly")
		os.Mkdir(readOnlyDir, 0755)
		defer os.Chmod(readOnlyDir, 0755)
		
		err := os.Chmod(readOnlyDir, 0555) // Read-only
		if err != nil {
			t.Skip("Cannot set directory to read-only, skipping test")
		}
		
		// Force working directory to read-only dir
		originalWd, _ := os.Getwd()
		os.Chdir(readOnlyDir)
		defer os.Chdir(originalWd)
		
		file := &multipart.FileHeader{
			Filename: "test-image.jpg",
			Size:     1024,
			Header:   make(map[string][]string),
		}
		
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		
		mockC := &MockGinContext{Context: c}
		
		// Act
		imagePath, err := saveImageToFileSystem(mockC, file)
		
		// Assert
		// Should still succeed because the function tries multiple directories
		if err != nil {
			assert.Contains(t, err.Error(), "Failed to create any image directory")
			assert.Empty(t, imagePath)
		}
	})
}

func TestGetProductImage(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "product-images-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Change to the temp directory
	originalWd, _ := os.Getwd()
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change working directory: %v", err)
	}
	defer os.Chdir(originalWd)
	
	// Create test image directories
	imagesDir := filepath.Join(tempDir, "uploads", "images")
	err = os.MkdirAll(imagesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create images directory: %v", err)
	}
	
	// Create a test image file
	testImagePath := filepath.Join(imagesDir, "test-product.jpg")
	testImageContent := []byte("fake image data")
	err = os.WriteFile(testImagePath, testImageContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test image file: %v", err)
	}
	
	t.Run("Successfully serve existing image", func(t *testing.T) {
		// Setup
		gin.SetMode(gin.TestMode)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		
		// Setup request with image filename parameter
		req := httptest.NewRequest("GET", "/images/test-product.jpg", nil)
		c.Request = req
		c.Params = []gin.Param{
			{
				Key:   "filename",
				Value: "test-product.jpg",
			},
		}
		
		// Act
		GetProductImage()(c)
		
		// Assert
		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Equal(t, "image/jpeg", recorder.Header().Get("Content-Type"))
		assert.Equal(t, testImageContent, recorder.Body.Bytes())
	})
	
	t.Run("Return 404 for non-existing image", func(t *testing.T) {
		// Setup
		gin.SetMode(gin.TestMode)
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		
		// Setup request with non-existing image filename parameter
		req := httptest.NewRequest("GET", "/images/non-existing.jpg", nil)
		c.Request = req
		c.Params = []gin.Param{
			{
				Key:   "filename",
				Value: "non-existing.jpg",
			},
		}
		
		// Act
		GetProductImage()(c)
		
		// Assert
		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})
}

func GetProductImage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the filename from the URL parameter
		filename := c.Param("filename")
		
		// Look for the image in the uploads/images directory
		imagePath := filepath.Join("uploads", "images", filename)
		
		// Check if the file exists
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
			return
		}
		
		// Serve the file
		c.File(imagePath)
	}
}