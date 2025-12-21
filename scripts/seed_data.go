package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ProductResponse represents the response from product service
type ProductResponse struct {
	Data     []Product `json:"data"`
	Total    int64     `json:"total"`
	Page     int64     `json:"page"`
	Pages    int       `json:"pages"`
	HasNext  bool      `json:"has_next"`
	HasPrev  bool      `json:"has_prev"`
}

type Product struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Category    string         `json:"category"`
	Variants    []ProductVariant `json:"variants"`
	UserID      string         `json:"user_id"`
	Status      string         `json:"status"`
}

type ProductVariant struct {
	ID        string  `json:"id"`
	Size      string  `json:"size"`
	Color     string  `json:"color"`
	Material  string  `json:"material"`
	CostPrice float64 `json:"cost_price"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	VariantID string  `json:"variant_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	CostPrice float64 `json:"cost_price"`
	Price     float64 `json:"price"`
	VendorID  string  `json:"vendor_id"`
}

// UserAddress model matching user-service
type UserAddress struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Street    string         `gorm:"type:varchar(255)" json:"street"`
	City      string         `gorm:"type:varchar(100)" json:"city"`
	State     string         `gorm:"type:varchar(100)" json:"state"`
	IsDefault bool           `gorm:"default:false" json:"is_default"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for UserAddress
func (UserAddress) TableName() string {
	return "user_addresses"
}

// User model matching auth-service
type User struct {
	gorm.Model
	ID         uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Email      *string        `gorm:"unique;not null" json:"email"`
	FirstName  *string        `gorm:"type:varchar(100)" json:"first_name"`
	LastName   *string        `gorm:"type:varchar(100)" json:"last_name"`
	UserType   string         `gorm:"type:varchar(50);default:'USER'" json:"user_type"`
	Password   *string        `gorm:"not null" json:"password"`
	Phone      *string        `gorm:"type:varchar(15)" json:"phone"`
	IsVerify   bool           `gorm:"default:false" json:"is_verify"`
	IsDisabled bool           `gorm:"default:false" json:"is_disabled"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserServiceUser model matching user-service (without password)
type UserServiceUser struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Email     *string        `gorm:"unique;not null" json:"email"`
	FirstName *string        `gorm:"type:varchar(100)" json:"first_name"`
	LastName  *string        `gorm:"type:varchar(100)" json:"last_name"`
	UserType  *string        `gorm:"type:varchar(50);default:'USER'" json:"user_type"`
	Phone     *string        `gorm:"type:varchar(15)" json:"phone"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for UserServiceUser
func (UserServiceUser) TableName() string {
	return "users"
}

// Order model matching order-service
type Order struct {
	gorm.Model
	OrderID         string         `gorm:"uniqueIndex;not null"`
	UserID          string         `gorm:"not null"`
	Items           datatypes.JSON `gorm:"type:jsonb;not null"`
	Status          string         `gorm:"not null;default:'pending'"`
	Source          string         `gorm:"not null;default:'web'"`
	TotalPrice      float64        `gorm:"not null"`
	TotalCost       float64        `gorm:"not null;default:0"`
	TotalRevenue    float64        `gorm:"not null;default:0"`
	PaymentMethod   string         `gorm:"not null;default:'cod'"`
	PaymentStatus   string         `gorm:"not null;default:'unpaid'"`
	PaymentIntentID *string        `gorm:"column:payment_intent_id" json:"payment_intent_id,omitempty"`
	ShippingStatus  string         `gorm:"not null;default:'pending'"`
	ShippingAddress string         `gorm:"not null"`
	ShippingInfo    datatypes.JSON `gorm:"type:jsonb;default:'{}'"`
	PlatformFee     float64        `gorm:"not null;default:0"`
	VendorAmount    float64        `gorm:"not null;default:0"`
	DeliveryDate     *time.Time     `json:"delivery_date"`
	PaymentReleaseDate *time.Time   `json:"payment_release_date"`
}

func main() {
	// Load environment variables from root project directory
	// Try multiple paths: current dir, parent dir (root), and explicit .env path
	envPaths := []string{
		"../.env",                 // Parent directory (root project) - most common
		".env",                    // Current directory
		"../../.env",              // Two levels up
	}
	
	var envLoaded bool
	var loadedPath string
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			loadedPath = path
			envLoaded = true
			break
		}
	}
	
	if envLoaded {
		log.Printf("✅ Loaded .env from: %s\n", loadedPath)
	} else {
		log.Println("⚠️  Warning: Could not load .env file, using environment variables")
		log.Println("   Tried paths:", envPaths)
		log.Println("   Make sure .env file exists in root project directory")
	}

	// Initialize databases
	authDB, err := initAuthDB()
	if err != nil {
		log.Fatalf("Failed to connect to auth database: %v", err)
	}
	
	userDB, err := initUserDB()
	if err != nil {
		log.Fatalf("Failed to connect to user database: %v", err)
	}
	
	orderDB, err := initOrderDB()
	if err != nil {
		log.Fatalf("Failed to connect to order database: %v", err)
	}

	ctx := context.Background()

	// Step 1: Create 20 users
	log.Println("Step 1: Creating 20 users...")
	users, userAddresses, err := createUsers(ctx, authDB, userDB, 20)
	if err != nil {
		log.Fatalf("Failed to create users: %v", err)
	}
	log.Printf("✅ Created %d users successfully\n", len(users))

	// Step 2: Fetch products from product-service
	log.Println("Step 2: Fetching products from product-service...")
	products, err := fetchProducts()
	if err != nil {
		log.Fatalf("Failed to fetch products: %v", err)
	}
	if len(products) == 0 {
		log.Fatalf("No products found. Please ensure product-service is running and has products.")
	}
	log.Printf("✅ Fetched %d products successfully\n", len(products))

	// Step 3: Create orders for all months in 2025
	log.Println("Step 3: Creating orders for all months in 2025...")
	totalOrders := 0
	for month := 1; month <= 12; month++ {
		ordersCreated, err := createOrdersForMonth(ctx, orderDB, users, userAddresses, products, 2025, month, 10)
		if err != nil {
			log.Printf("Warning: Failed to create orders for month %d: %v\n", month, err)
			continue
		}
		totalOrders += ordersCreated
		log.Printf("✅ Created %d orders for month %d/2025\n", ordersCreated, month)
	}

	log.Printf("\n🎉 Seed data completed successfully!")
	log.Printf("   - Users created: %d", len(users))
	log.Printf("   - Products fetched: %d", len(products))
	log.Printf("   - Total orders created: %d", totalOrders)
}

func initAuthDB() (*gorm.DB, error) {
	host := getEnv("AUTH_POSTGRES_HOST", getEnv("POSTGRES_HOST", "localhost"))
	port := getEnv("AUTH_POSTGRES_PORT", getEnv("POSTGRES_PORT", "5432"))
	user := getEnv("AUTH_POSTGRES_USER", getEnv("POSTGRES_USER", "postgres"))
	password := getEnv("AUTH_POSTGRES_PASSWORD", getEnv("POSTGRES_PASSWORD", ""))
	dbname := getEnv("AUTH_POSTGRES_DB", getEnv("POSTGRES_DB", "authdb"))
	sslmode := getEnv("AUTH_POSTGRES_SSLMODE", getEnv("POSTGRES_SSLMODE", "disable"))

	if password == "" {
		return nil, fmt.Errorf("AUTH_POSTGRES_PASSWORD or POSTGRES_PASSWORD environment variable is required for auth database")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	log.Printf("Connecting to auth database: host=%s, port=%s, user=%s, dbname=%s", host, port, user, dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth database: %v. Please check your database credentials", err)
	}

	return db, nil
}

func initUserDB() (*gorm.DB, error) {
	host := getEnv("USER_POSTGRES_HOST", 
		getEnv("AUTH_POSTGRES_HOST", 
			getEnv("POSTGRES_HOST", "localhost")))
	port := getEnv("USER_POSTGRES_PORT", 
		getEnv("AUTH_POSTGRES_PORT", 
			getEnv("POSTGRES_PORT", "5432")))
	user := getEnv("USER_POSTGRES_USER", getEnv("POSTGRES_USER", "postgres"))
	password := getEnv("USER_POSTGRES_PASSWORD", 
		getEnv("AUTH_POSTGRES_PASSWORD", 
			getEnv("POSTGRES_PASSWORD", "")))
	dbname := getEnv("USER_POSTGRES_DB", getEnv("DB_NAME", "user_service_db"))
	sslmode := getEnv("USER_POSTGRES_SSLMODE", 
		getEnv("AUTH_POSTGRES_SSLMODE", 
			getEnv("POSTGRES_SSLMODE", "disable")))

	if password == "" {
		return nil, fmt.Errorf("USER_POSTGRES_PASSWORD, AUTH_POSTGRES_PASSWORD, or POSTGRES_PASSWORD environment variable is required for user database")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	log.Printf("Connecting to user database: host=%s, port=%s, user=%s, dbname=%s", host, port, user, dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user database: %v. Please check your database credentials", err)
	}

	return db, nil
}

func initOrderDB() (*gorm.DB, error) {
	// Try ORDER_POSTGRES_* first, then fallback to AUTH or USER config
	host := getEnv("ORDER_POSTGRES_HOST", 
		getEnv("AUTH_POSTGRES_HOST", 
			getEnv("USER_POSTGRES_HOST", 
				getEnv("POSTGRES_HOST", "localhost"))))
	port := getEnv("ORDER_POSTGRES_PORT", 
		getEnv("AUTH_POSTGRES_PORT", 
			getEnv("USER_POSTGRES_PORT", 
				getEnv("POSTGRES_PORT", "5432"))))
	user := getEnv("ORDER_POSTGRES_USER", 
		getEnv("AUTH_POSTGRES_USER", 
			getEnv("USER_POSTGRES_USER", 
				getEnv("POSTGRES_USER", "postgres"))))
	// Try multiple fallbacks for password
	password := getEnv("ORDER_POSTGRES_PASSWORD", 
		getEnv("AUTH_POSTGRES_PASSWORD", 
			getEnv("USER_POSTGRES_PASSWORD", 
				getEnv("POSTGRES_PASSWORD", 
					getEnv("DB_PASSWORD", "")))))
	dbname := getEnv("ORDER_POSTGRES_DB", 
		getEnv("POSTGRES_DB", "orderdb"))
	sslmode := getEnv("ORDER_POSTGRES_SSLMODE", 
		getEnv("AUTH_POSTGRES_SSLMODE", 
			getEnv("USER_POSTGRES_SSLMODE", 
				getEnv("POSTGRES_SSLMODE", "disable"))))

	if password == "" {
		return nil, fmt.Errorf("ORDER_POSTGRES_PASSWORD, AUTH_POSTGRES_PASSWORD, USER_POSTGRES_PASSWORD, POSTGRES_PASSWORD, or DB_PASSWORD environment variable is required")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	log.Printf("Connecting to order database: host=%s, port=%s, user=%s, dbname=%s", host, port, user, dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order database: %v. Please check your database credentials", err)
	}

	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createUsers(ctx context.Context, authDB *gorm.DB, userDB *gorm.DB, count int) ([]User, map[string]string, error) {
	users := make([]User, 0, count)
	userAddresses := make(map[string]string) // Map userID -> full address string
	
	// Common password for all users (can be configured via environment variable)
	commonPassword := getEnv("SEED_USER_PASSWORD", "Aminumina1!")
	
	// Hash password once for all users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(commonPassword), 14)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %v", err)
	}
	hashedPasswordStr := string(hashedPassword)
	
	// Predefined addresses for users (street, city, state) - Hà Nội
	addresses := []struct {
		Street string
		City   string
		State  string
	}{
		{"123 Hoàn Kiếm", "Quận Hoàn Kiếm", "Hà Nội"},
		{"456 Tràng Tiền", "Quận Hoàn Kiếm", "Hà Nội"},
		{"789 Lý Thái Tổ", "Quận Hoàn Kiếm", "Hà Nội"},
		{"321 Nguyễn Du", "Quận Hai Bà Trưng", "Hà Nội"},
		{"654 Bà Triệu", "Quận Hai Bà Trưng", "Hà Nội"},
		{"987 Lê Duẩn", "Quận Hai Bà Trưng", "Hà Nội"},
		{"147 Cầu Giấy", "Quận Cầu Giấy", "Hà Nội"},
		{"258 Trần Duy Hưng", "Quận Cầu Giấy", "Hà Nội"},
		{"369 Hoàng Đạo Thúy", "Quận Cầu Giấy", "Hà Nội"},
		{"741 Nguyễn Chí Thanh", "Quận Ba Đình", "Hà Nội"},
		{"852 Đội Cấn", "Quận Ba Đình", "Hà Nội"},
		{"963 Giảng Võ", "Quận Ba Đình", "Hà Nội"},
		{"159 Láng Hạ", "Quận Đống Đa", "Hà Nội"},
		{"357 Tây Sơn", "Quận Đống Đa", "Hà Nội"},
		{"468 Khâm Thiên", "Quận Đống Đa", "Hà Nội"},
		{"579 Nguyễn Lương Bằng", "Quận Đống Đa", "Hà Nội"},
		{"680 Láng Thượng", "Quận Đống Đa", "Hà Nội"},
		{"791 Kim Mã", "Quận Ba Đình", "Hà Nội"},
		{"802 Điện Biên Phủ", "Quận Ba Đình", "Hà Nội"},
		{"913 Hoàng Hoa Thám", "Quận Ba Đình", "Hà Nội"},
	}
	
	// Base email for seed users
	baseEmail := getEnv("SEED_BASE_EMAIL", "datttne123@gmail.com")
	
	// Extract email prefix and domain
	emailParts := strings.Split(baseEmail, "@")
	if len(emailParts) != 2 {
		return nil, nil, fmt.Errorf("invalid base email format: %s", baseEmail)
	}
	emailPrefix := emailParts[0]
	emailDomain := emailParts[1]
	
	// Vietnamese first names and last names for seed users
	firstNames := []string{"Nguyễn", "Trần", "Lê", "Phạm", "Hoàng", "Huỳnh", "Phan", "Vũ", "Võ", "Đặng", "Bùi", "Đỗ", "Hồ", "Ngô", "Dương", "Lý", "Đinh", "Đào", "Mai", "Tạ"}
	lastNames := []string{"Văn", "Thị", "Minh", "Anh", "Hùng", "Dũng", "Lan", "Hương", "Thảo", "Linh", "Nam", "Hải", "Tuấn", "Huy", "Khang", "Phương", "Nga", "Trang", "Vy", "Quang"}
	
	for i := 1; i <= count; i++ {
		// Create unique emails: datttne1231@gmail.com, datttne1232@gmail.com, etc.
		email := fmt.Sprintf("%s%d@%s", emailPrefix, i, emailDomain)
		phone := fmt.Sprintf("090%08d", i) // Vietnamese phone format
		
		// Generate Vietnamese name
		firstName := firstNames[(i-1)%len(firstNames)] + " " + lastNames[(i-1)%len(lastNames)]
		lastName := fmt.Sprintf("User %d", i) // Simple last name
		firstNamePtr := &firstName
		lastNamePtr := &lastName
		
		// Get address (cycle through addresses if more users than addresses)
		addr := addresses[(i-1)%len(addresses)]
		fullAddress := fmt.Sprintf("%s, %s, %s", addr.Street, addr.City, addr.State)
		
		// Create user in auth-service database
		user := User{
			Email:     &email,
			Password:  &hashedPasswordStr,
			Phone:     &phone,
			FirstName: firstNamePtr,
			LastName:  lastNamePtr,
			UserType:  "USER",
			IsVerify:  true,
		}
		
		// Check if user already exists, update if needed
		var existingUser User
		if err := authDB.WithContext(ctx).Where("email = ?", email).First(&existingUser).Error; err == nil {
			log.Printf("User %s already exists, updating if needed\n", email)
			// Update FirstName and LastName if they are nil or empty
			// This is critical to avoid nil pointer dereference in Login function
			updates := make(map[string]interface{})
			if existingUser.FirstName == nil || (existingUser.FirstName != nil && *existingUser.FirstName == "") {
				updates["first_name"] = firstNamePtr
			}
			// Always ensure LastName is set (even if empty string, not nil)
			// This prevents nil pointer dereference in Login function
			if existingUser.LastName == nil {
				// If lastNamePtr is nil, set to empty string to avoid nil pointer
				if lastNamePtr == nil {
					emptyStr := ""
					updates["last_name"] = &emptyStr
				} else {
					updates["last_name"] = lastNamePtr
				}
			} else if *existingUser.LastName == "" && lastNamePtr != nil {
				updates["last_name"] = lastNamePtr
			}
			// Update IsVerify to true if needed
			if !existingUser.IsVerify {
				updates["is_verify"] = true
			}
			// Update password if needed (in case it changed)
			if existingUser.Password == nil || *existingUser.Password != hashedPasswordStr {
				updates["password"] = hashedPasswordStr
			}
			// Update phone if needed
			if existingUser.Phone == nil || *existingUser.Phone != phone {
				updates["phone"] = &phone
			}
			
			if len(updates) > 0 {
				if err := authDB.WithContext(ctx).Model(&existingUser).Updates(updates).Error; err != nil {
					log.Printf("Warning: Failed to update user %s: %v\n", email, err)
				} else {
					log.Printf("Updated user %s with missing fields\n", email)
				}
				// Reload user to get updated values
				authDB.WithContext(ctx).Where("email = ?", email).First(&existingUser)
			}
			// Use existing user
			user = existingUser
		} else if err := authDB.WithContext(ctx).Create(&user).Error; err != nil {
			log.Printf("Warning: Failed to create user %s: %v\n", email, err)
			continue
		}
		
		userID := user.ID
		userTypeStr := user.UserType
		
		// Check if user already exists in user-service database (by ID or email)
		var existingUserServiceUser UserServiceUser
		userExistsInUserService := false
		
		// First try to find by ID (correct case - IDs match) - including soft deleted
		if err := userDB.WithContext(ctx).Unscoped().Where("id = ?", userID).First(&existingUserServiceUser).Error; err == nil {
			userExistsInUserService = true
			// If user was soft deleted, restore it
			if existingUserServiceUser.DeletedAt.Valid {
				userDB.WithContext(ctx).Unscoped().Model(&existingUserServiceUser).
					Update("deleted_at", nil)
				log.Printf("Restored soft-deleted user %s in user-service\n", email)
			}
			// Update first_name and last_name if they are null
			if existingUserServiceUser.FirstName == nil || existingUserServiceUser.LastName == nil {
				userDB.WithContext(ctx).Model(&existingUserServiceUser).
					Updates(map[string]interface{}{
						"first_name": firstNamePtr,
						"last_name":  lastNamePtr,
					})
				log.Printf("Updated name for user %s in user-service\n", email)
			}
		} else {
			// Try to find by email (in case ID doesn't match - need to fix) - including soft deleted
			if err := userDB.WithContext(ctx).Unscoped().Where("email = ?", email).First(&existingUserServiceUser).Error; err == nil {
				oldUserID := existingUserServiceUser.ID
				// IDs don't match - need to fix this
				if oldUserID != userID {
					log.Printf("⚠️  User %s has mismatched IDs: auth=%s, user=%s. Fixing...\n", 
						email, userID.String(), oldUserID.String())
					
					// Delete all addresses of the old user (they reference old ID) - hard delete
					userDB.WithContext(ctx).Unscoped().Where("user_id = ?", oldUserID).Delete(&UserAddress{})
					
					// Hard delete the old user with wrong ID (Unscoped() bypasses soft delete)
					userDB.WithContext(ctx).Unscoped().Where("id = ?", oldUserID).Delete(&UserServiceUser{})
					
					// Verify deletion and check if user with correct ID already exists (including soft deleted)
					var checkUser UserServiceUser
					if err := userDB.WithContext(ctx).Unscoped().Where("id = ?", userID).First(&checkUser).Error; err == nil {
						// User with correct ID already exists, restore if soft deleted and update info
						if checkUser.DeletedAt.Valid {
							userDB.WithContext(ctx).Unscoped().Model(&checkUser).
								Update("deleted_at", nil)
							log.Printf("Restored soft-deleted user %s with correct ID\n", email)
						}
						userDB.WithContext(ctx).Model(&checkUser).Updates(map[string]interface{}{
							"first_name": firstNamePtr,
							"last_name":  lastNamePtr,
							"user_type":  userTypeStr,
							"phone":      &phone,
						})
						log.Printf("✅ Fixed user %s: updated existing user with correct ID %s\n", email, userID.String())
						userExistsInUserService = true
					} else {
						// Create new user with correct ID from auth-service
						userServiceUser := UserServiceUser{
							ID:        userID,
							Email:     &email,
							FirstName: firstNamePtr,
							LastName:  lastNamePtr,
							UserType:  &userTypeStr,
							Phone:     &phone,
						}
						
						if err := userDB.WithContext(ctx).Create(&userServiceUser).Error; err != nil {
							log.Printf("Warning: Failed to create user in user-service database %s: %v\n", email, err)
							userExistsInUserService = false
						} else {
							log.Printf("✅ Fixed user %s: recreated with correct ID %s\n", email, userID.String())
							userExistsInUserService = true
							// Verify user was created successfully
							var verifyUser UserServiceUser
							if err := userDB.WithContext(ctx).Where("id = ?", userID).First(&verifyUser).Error; err != nil {
								log.Printf("Warning: User %s created but not found immediately, will skip address creation\n", email)
								userExistsInUserService = false
							}
						}
					}
				} else {
					// IDs match, just update info
					userExistsInUserService = true
					// If user was soft deleted, restore it first
					if existingUserServiceUser.DeletedAt.Valid {
						userDB.WithContext(ctx).Unscoped().Model(&existingUserServiceUser).
							Update("deleted_at", nil)
						log.Printf("Restored soft-deleted user %s in user-service\n", email)
					}
					updates := map[string]interface{}{
						"first_name": firstNamePtr,
						"last_name":  lastNamePtr,
					}
					if existingUserServiceUser.UserType == nil || *existingUserServiceUser.UserType != userTypeStr {
						updates["user_type"] = userTypeStr
					}
					userDB.WithContext(ctx).Model(&existingUserServiceUser).Updates(updates)
				}
			}
		}
		
		if !userExistsInUserService {
			// Create user in user-service database with same ID (required for foreign key)
			userServiceUser := UserServiceUser{
				ID:        userID,
				Email:     &email,
				FirstName: firstNamePtr, // Add first name
				LastName:  lastNamePtr,  // Add last name
				UserType:  &userTypeStr,
				Phone:     &phone,
			}
			
			if err := userDB.WithContext(ctx).Create(&userServiceUser).Error; err != nil {
				log.Printf("Warning: Failed to create user in user-service database %s: %v\n", email, err)
				// Skip address creation if user creation failed
				userExistsInUserService = false
			} else {
				userExistsInUserService = true
			}
		}
		
		// Create address in user-service database (only if user exists)
		// Verify user exists in user-service before creating address
		if userExistsInUserService {
			var verifyUser UserServiceUser
			if err := userDB.WithContext(ctx).Where("id = ?", userID).First(&verifyUser).Error; err != nil {
				log.Printf("Warning: User %s (ID: %s) not found in user-service database after creation, skipping address creation\n", email, userID.String())
				userExistsInUserService = false
			} else {
				// User exists, proceed with address creation
				// Check if address already exists
				var existingAddress UserAddress
				if err := userDB.WithContext(ctx).Where("user_id = ? AND is_default = ?", userID, true).First(&existingAddress).Error; err == nil {
					log.Printf("Default address already exists for user %s, skipping\n", email)
				} else {
					userAddress := UserAddress{
						UserID:    userID,
						Street:    addr.Street,
						City:      addr.City,
						State:     addr.State,
						IsDefault: true, // Set as default address
					}
					
					if err := userDB.WithContext(ctx).Create(&userAddress).Error; err != nil {
						log.Printf("Warning: Failed to create address for user %s: %v\n", email, err)
						// Continue anyway, we'll use the full address string
					} else {
						log.Printf("✅ Created address for user %s\n", email)
					}
				}
			}
		} else {
			log.Printf("⚠️  User %s does not exist in user-service, skipping address creation\n", email)
		}
		
		// Store full address string for order creation
		userAddresses[userID.String()] = fullAddress
		
		users = append(users, user)
	}
	
	log.Printf("✅ All users use common password: %s\n", commonPassword)
	return users, userAddresses, nil
}

func fetchProducts() ([]Product, error) {
	// Get product service URL from environment or use default
	// Note: If running in Docker, use service name (e.g., http://product-service:8082)
	// If running locally, use http://localhost:8082
	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
	if productServiceURL == "" {
		// Try Docker service name first, then localhost
		productServiceURL = "http://product-service:8082"
	}
	
	// Fetch products with pagination
	allProducts := make([]Product, 0)
	page := int64(1)
	limit := int64(100) // Fetch 100 products per page
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	for {
		url := fmt.Sprintf("%s/products/get/all?page=%d&limit=%d", productServiceURL, page, limit)
		
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
		
		// Try without auth first (if accessing service directly in Docker network)
		resp, err := client.Do(req)
		if err != nil {
			// If Docker service name fails, try localhost
			if productServiceURL == "http://product-service:8082" {
				log.Println("⚠️  Failed to connect to product-service, trying localhost...")
				productServiceURL = "http://localhost:8082"
				url = fmt.Sprintf("%s/products/get/all?page=%d&limit=%d", productServiceURL, page, limit)
				req, _ = http.NewRequest("GET", url, nil)
				resp, err = client.Do(req)
			}
			if err != nil {
				return nil, fmt.Errorf("failed to fetch products: %v. Make sure product-service is running.", err)
			}
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return nil, fmt.Errorf("authentication required. Please ensure product-service allows direct access or add authentication token to the script")
		}
		
		if resp.StatusCode != http.StatusOK {
			bodyBytes := make([]byte, 200)
			resp.Body.Read(bodyBytes)
			return nil, fmt.Errorf("product service returned status %d: %s", resp.StatusCode, string(bodyBytes))
		}
		
		var productResp ProductResponse
		if err := json.NewDecoder(resp.Body).Decode(&productResp); err != nil {
			return nil, fmt.Errorf("failed to decode product response: %v", err)
		}
		
		// Filter only products with variants and status "onsale"
		for _, product := range productResp.Data {
			if product.Status == "onsale" && len(product.Variants) > 0 {
				allProducts = append(allProducts, product)
			}
		}
		
		if !productResp.HasNext {
			break
		}
		
		page++
		
		// Safety limit to prevent infinite loops
		if page > 100 {
			log.Println("⚠️  Reached page limit (100), stopping pagination")
			break
		}
	}
	
	return allProducts, nil
}

func createOrdersForMonth(ctx context.Context, db *gorm.DB, users []User, userAddresses map[string]string, products []Product, year, month int, minOrdersPerMonth int) (int, error) {
	if len(users) == 0 {
		return 0, fmt.Errorf("no users available")
	}
	if len(products) == 0 {
		return 0, fmt.Errorf("no products available")
	}
	
	// Create orders with some randomness (10-15 orders per month)
	numOrders := minOrdersPerMonth + randomInt(0, 6)
	ordersCreated := 0
	
	log.Printf("  Creating %d orders for %d/%d...", numOrders, month, year)
	
	// Create start and end dates for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	
	// Calculate days in month
	daysInMonth := int(endDate.Sub(startDate).Hours() / 24)
	
	for i := 0; i < numOrders; i++ {
		if (i+1)%5 == 0 {
			log.Printf("  Progress: %d/%d orders created for month %d", i+1, numOrders, month)
		}
		// Random day in the month
		day := randomInt(1, daysInMonth+1)
		hour := randomInt(9, 22) // Business hours
		minute := randomInt(0, 60)
		second := randomInt(0, 60)
		
		orderDate := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
		
		// Random user
		user := users[randomInt(0, len(users))]
		userID := user.ID.String()
		
		// Skip if user ID is empty
		if userID == "" || userID == "00000000-0000-0000-0000-000000000000" {
			continue
		}
		
		// Random number of items (1-5 items per order)
		numItems := randomInt(1, 6)
		orderItems := make([]OrderItem, 0, numItems)
		var totalPrice, totalCost, totalRevenue float64
		
		// Select random products with retry limit to avoid infinite loop
		selectedProducts := make(map[string]bool)
		maxRetries := 100 // Prevent infinite loop
		retryCount := 0
		
		for len(orderItems) < numItems && retryCount < maxRetries {
			retryCount++
			product := products[randomInt(0, len(products))]
			
			// Avoid duplicate products in same order
			if selectedProducts[product.ID] {
				// If we've tried all products and still don't have enough items, break
				if len(selectedProducts) >= len(products) {
					break
				}
				continue
			}
			selectedProducts[product.ID] = true
			
			// Select a random variant
			if len(product.Variants) == 0 {
				continue
			}
			
			variant := product.Variants[randomInt(0, len(product.Variants))]
			
			// Skip if variant has no stock
			if variant.Quantity <= 0 {
				continue
			}
			
			// Random quantity (1-3 items)
			quantity := randomInt(1, min(4, variant.Quantity+1))
			
			itemPrice := variant.Price
			itemCost := variant.CostPrice
			
			orderItem := OrderItem{
				ProductID: product.ID,
				VariantID: variant.ID,
				Name:      product.Name,
				Quantity:  quantity,
				CostPrice: itemCost,
				Price:     itemPrice,
				VendorID:  product.UserID,
			}
			
			orderItems = append(orderItems, orderItem)
			totalPrice += itemPrice * float64(quantity)
			totalCost += itemCost * float64(quantity)
			totalRevenue += (itemPrice - itemCost) * float64(quantity)
		}
		
		if len(orderItems) == 0 {
			log.Printf("  Warning: Could not create order items for user %s (no available products/variants), skipping\n", userID)
			continue
		}
		
		// Convert order items to JSON
		itemsJSON, err := json.Marshal(orderItems)
		if err != nil {
			log.Printf("Warning: Failed to marshal order items: %v\n", err)
			continue
		}
		
		// Generate order ID
		orderID := generateOrderID(orderDate)
		
		// Random payment method
		paymentMethods := []string{"COD", "STRIPE"}
		paymentMethod := paymentMethods[randomInt(0, len(paymentMethods))]
		
		// Set all orders to SHIPPED status as requested
		status := "SHIPPED"
		shippingStatus := "SHIPPED"
		// Payment status: if shipped, payment should be completed
		paymentStatus := "PAID"
		if paymentMethod == "COD" {
			paymentStatus = "COMPLETED"
		}
		
		// Get shipping address from user addresses map
		shippingAddress := "123 Main Street, Ho Chi Minh City" // Default fallback
		if addr, exists := userAddresses[userID]; exists && addr != "" {
			shippingAddress = addr
		}
		
		// Create order with SHIPPED status
		order := Order{
			OrderID:         orderID,
			UserID:          userID,
			Items:           datatypes.JSON(itemsJSON),
			TotalPrice:      totalPrice,
			TotalCost:       totalCost,
			TotalRevenue:    totalRevenue,
			Status:          status, // SHIPPED
			PaymentMethod:   paymentMethod,
			PaymentStatus:   paymentStatus,
			ShippingAddress: shippingAddress,
			ShippingStatus:  shippingStatus, // SHIPPED
			Source:          "web",
			ShippingInfo:    datatypes.JSON([]byte("{}")),
		}
		
		// Set CreatedAt manually to the order date
		order.CreatedAt = orderDate
		order.UpdatedAt = orderDate
		
		// Save order
		if err := db.WithContext(ctx).Create(&order).Error; err != nil {
			log.Printf("Warning: Failed to create order: %v\n", err)
			continue
		}
		
		// Verify and ensure status is SHIPPED after creation
		// Update status and created_at to ensure they are set correctly
		updateResult := db.Model(&Order{}).
			Where("order_id = ?", order.OrderID).
			Updates(map[string]interface{}{
				"status":      "SHIPPED",
				"shipping_status": "SHIPPED",
				"created_at":  orderDate,
			})
		if updateResult.Error != nil {
			log.Printf("Warning: Failed to update order status for %s: %v\n", order.OrderID, updateResult.Error)
		}
		
		ordersCreated++
		if ordersCreated%10 == 0 {
			log.Printf("  Created %d orders so far for month %d...", ordersCreated, month)
		}
	}
	
	log.Printf("  ✅ Completed creating %d orders for month %d/%d", ordersCreated, month, year)
	return ordersCreated, nil
}

func generateOrderID(orderDate time.Time) string {
	dateStr := orderDate.Format("20060102")
	timeStr := orderDate.Format("150405")
	randStr := generateRandomString(5)
	return fmt.Sprintf("ORD-%s-%s-%s", dateStr, timeStr, randStr)
}

func generateRandomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ""
		}
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

func randomInt(min, max int) int {
	if min >= max {
		return min
	}
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	return int(num.Int64()) + min
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

