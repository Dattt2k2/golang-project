package tests

import (
	"net/http"
	"testing"
)

var (
	httpClient *http.Client
	baseURL    string
)

// setupTest sets up the test environment.
func setupTest(t *testing.T) func() {
	httpClient = &http.Client{}
	baseURL = "http://localhost:8081/users" // Thay URL này bằng URL của service bạn

	return func() {
		// Clean up logic nếu cần, ví dụ đóng kết nối nếu bạn có mở file hay resource nào
	}
}
