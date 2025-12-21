package migration

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// AdminUser represents the admin user structure
type AdminUser struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email     *string   `gorm:"unique;not null"`
	FirstName *string   `gorm:"type:varchar(100)"`
	LastName  *string   `gorm:"type:varchar(100)"`
	UserType  string    `gorm:"type:varchar(50);default:'ADMIN'"`
	Password  *string   `gorm:"not null"`
	IsVerify  bool      `gorm:"default:true"`
	Phone     *string   `gorm:"type:varchar(15)"`
}

func (AdminUser) TableName() string {
	return "users"
}

// UserServiceAdmin represents the admin user in user-service
type UserServiceAdmin struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Email     *string        `gorm:"unique;not null"`
	FirstName *string        `gorm:"type:varchar(100)"`
	LastName  *string        `gorm:"type:varchar(100)"`
	UserType  *string        `gorm:"type:varchar(50)"`
	Phone     *string        `gorm:"type:varchar(15)"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (UserServiceAdmin) TableName() string {
	return "users"
}

// CreateAdminUserMigration creates the admin user in both auth-service and user-service databases
func CreateAdminUserMigration() error {
	// Load environment variables - try root .env first (has USER_POSTGRES_*), then local .env
	if err := godotenv.Load("../../.env"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			if err := godotenv.Load(".env"); err != nil {
				log.Printf("Warning: Could not load .env file: %v\n", err)
			}
		}
	}
	// Also load local .env to override if needed
	_ = godotenv.Load(".env")

	ctx := context.Background()

	// Connect to auth-service database
	authDB, err := connectAuthDB()
	if err != nil {
		return fmt.Errorf("failed to connect to auth database: %v", err)
	}

	// Connect to user-service database
	userDB, err := connectUserDB()
	if err != nil {
		log.Printf("Warning: Failed to connect to user database: %v\n", err)
		log.Println("Admin user will only be created in auth-service")
		userDB = nil
	}

	// Create admin user
	if err := createAdminUser(ctx, authDB, userDB); err != nil {
		return fmt.Errorf("failed to create admin user: %v", err)
	}

	log.Println("✅ Admin user migration completed successfully")
	return nil
}

func connectAuthDB() (*gorm.DB, error) {
	host := getEnv("AUTH_POSTGRES_HOST", getEnv("POSTGRES_HOST", "localhost"))
	port := getEnv("AUTH_POSTGRES_PORT", getEnv("POSTGRES_PORT", "5432"))
	user := getEnv("AUTH_POSTGRES_USER", getEnv("POSTGRES_USER", "postgres"))
	password := getEnv("AUTH_POSTGRES_PASSWORD", getEnv("POSTGRES_PASSWORD", ""))
	dbname := getEnv("AUTH_POSTGRES_DB", getEnv("POSTGRES_DB", "auth_db"))
	sslmode := getEnv("AUTH_POSTGRES_SSLMODE", getEnv("POSTGRES_SSLMODE", "disable"))

	if password == "" {
		return nil, fmt.Errorf("AUTH_POSTGRES_PASSWORD or POSTGRES_PASSWORD environment variable is required")
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// If connection fails and host is not localhost, try localhost as fallback
		// This handles the case where .env has Docker service names but running locally
		if host != "localhost" && host != "127.0.0.1" {
			log.Printf("⚠️  Failed to connect to %s, trying localhost...\n", host)
			dsnLocalhost := fmt.Sprintf("host=localhost port=%s user=%s password=%s dbname=%s sslmode=%s",
				port, user, password, dbname, sslmode)
			db, err = gorm.Open(postgres.Open(dsnLocalhost), &gorm.Config{})
			if err == nil {
				log.Printf("✅ Connected to auth database using localhost\n")
				return db, nil
			}
		}
		return nil, fmt.Errorf("failed to connect to auth database: %v", err)
	}

	return db, nil
}

func connectUserDB() (*gorm.DB, error) {
	// Explicitly check USER_POSTGRES_* first, don't fallback to POSTGRES_* to avoid using wrong database
	host := getEnv("USER_POSTGRES_HOST", "")
	if host == "" {
		host = "localhost"
	}
	
	port := getEnv("USER_POSTGRES_PORT", "")
	if port == "" {
		port = "5433"
	}
	
	user := getEnv("USER_POSTGRES_USER", "")
	if user == "" {
		user = "postgres"
	}
	
	password := getEnv("USER_POSTGRES_PASSWORD", "")
	if password == "" {
		// Only fallback to POSTGRES_PASSWORD if USER_POSTGRES_PASSWORD is not set
		password = getEnv("POSTGRES_PASSWORD", "")
	}
	
	// Explicitly use USER_POSTGRES_DB, don't fallback to POSTGRES_DB to avoid using wrong database
	dbname := getEnv("USER_POSTGRES_DB", "")
	if dbname == "" {
		// Try DB_NAME as fallback (user-service uses this), but not POSTGRES_DB
		dbname = getEnv("DB_NAME", "")
		if dbname == "" || dbname == "auth_db" {
			// If still empty or wrong, use default
			dbname = "user_db"
		}
	}
	
	sslmode := getEnv("USER_POSTGRES_SSLMODE", "disable")

	if password == "" {
		return nil, fmt.Errorf("USER_POSTGRES_PASSWORD or POSTGRES_PASSWORD environment variable is required")
	}

	log.Printf("🔌 Connecting to user-service database: host=%s, port=%s, dbname=%s, user=%s\n", host, port, dbname, user)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// If connection fails and host is not localhost, try localhost as fallback
		// This handles the case where .env has Docker service names but running locally
		if host != "localhost" && host != "127.0.0.1" {
			log.Printf("⚠️  Failed to connect to %s, trying localhost...\n", host)
			dsnLocalhost := fmt.Sprintf("host=localhost port=%s user=%s password=%s dbname=%s sslmode=%s",
				port, user, password, dbname, sslmode)
			db, err = gorm.Open(postgres.Open(dsnLocalhost), &gorm.Config{})
			if err == nil {
				log.Printf("✅ Connected to user database using localhost: %s\n", dbname)
				return db, nil
			}
		}
		return nil, fmt.Errorf("failed to connect to user database: %v", err)
	}

	log.Printf("✅ Connected to user database: %s\n", dbname)
	return db, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func createAdminUser(ctx context.Context, authDB, userDB *gorm.DB) error {
	// Admin user details
	adminEmail := "admin@ecomo.com"
	adminFirstName := "Admin"
	adminLastName := ""
	adminPassword := "Aminumina1!"
	adminUserType := "ADMIN"

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), 14)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}
	hashedPasswordStr := string(hashedPassword)

	// Check if admin already exists in auth-service
	var existingUser AdminUser
	if err := authDB.WithContext(ctx).Where("email = ?", adminEmail).First(&existingUser).Error; err == nil {
		log.Printf("Admin user %s already exists, updating information...\n", adminEmail)

		// Update existing admin user
		firstNamePtr := &adminFirstName
		lastNamePtr := &adminLastName

		updates := map[string]interface{}{
			"first_name": firstNamePtr,
			"user_type":  adminUserType,
			"password":   hashedPasswordStr,
			"is_verify":  true,
		}
		if existingUser.LastName == nil || (existingUser.LastName != nil && *existingUser.LastName == "") {
			updates["last_name"] = lastNamePtr
		}

		authDB.WithContext(ctx).Model(&existingUser).Updates(updates)

		// Update in user-service if connected - ensure ID matches auth-service
		if userDB != nil {
			// Always delete any user with this email but different ID (including soft-deleted) to prevent conflicts
			var deletedCount int64
			// Try to delete addresses, but ignore error if table doesn't exist
			_ = userDB.WithContext(ctx).Unscoped().Exec("DELETE FROM user_addresses WHERE user_id IN (SELECT id FROM users WHERE email = $1 AND id != $2)", adminEmail, existingUser.ID)
			result := userDB.WithContext(ctx).Unscoped().Exec("DELETE FROM users WHERE email = $1 AND id != $2", adminEmail, existingUser.ID)
			if result.Error == nil {
				deletedCount = result.RowsAffected
				if deletedCount > 0 {
					log.Printf("⚠️  Deleted %d old admin user(s) with mismatched ID\n", deletedCount)
				}
			}
			
			// Always use raw SQL to ensure correct ID and handle password column
			// Check if password column exists
			var hasPassword bool
			userDB.WithContext(ctx).Raw(`
				SELECT COUNT(*) > 0 
				FROM information_schema.columns 
				WHERE table_name = 'users' AND column_name = 'password' AND table_schema = 'public'
			`).Scan(&hasPassword)
			
			log.Printf("📝 Inserting admin into user-service: ID=%s, Email=%s, HasPassword=%v\n", existingUser.ID.String(), adminEmail, hasPassword)
			
			if hasPassword {
				dummyPassword := "N/A"
				result := userDB.WithContext(ctx).Exec(`
					INSERT INTO users (id, email, first_name, last_name, user_type, password, created_at, updated_at, deleted_at)
					VALUES ($1, $2, $3, $4, $5, $6, COALESCE((SELECT created_at FROM users WHERE id = $1), NOW()), NOW(), NULL)
					ON CONFLICT (id) DO UPDATE SET
						email = EXCLUDED.email,
						first_name = EXCLUDED.first_name,
						last_name = EXCLUDED.last_name,
						user_type = EXCLUDED.user_type,
						deleted_at = NULL,
						updated_at = NOW()
				`, existingUser.ID, adminEmail, adminFirstName, adminLastName, adminUserType, dummyPassword)
				if result.Error != nil {
					log.Printf("❌ ERROR: Failed to create/update admin in user-service: %v\n", result.Error)
				} else {
					log.Printf("✅ Created/updated admin in user-service: RowsAffected=%d\n", result.RowsAffected)
				}
			} else {
				result := userDB.WithContext(ctx).Exec(`
					INSERT INTO users (id, email, first_name, last_name, user_type, created_at, updated_at, deleted_at)
					VALUES ($1, $2, $3, $4, $5, COALESCE((SELECT created_at FROM users WHERE id = $1), NOW()), NOW(), NULL)
					ON CONFLICT (id) DO UPDATE SET
						email = EXCLUDED.email,
						first_name = EXCLUDED.first_name,
						last_name = EXCLUDED.last_name,
						user_type = EXCLUDED.user_type,
						deleted_at = NULL,
						updated_at = NOW()
				`, existingUser.ID, adminEmail, adminFirstName, adminLastName, adminUserType)
				if result.Error != nil {
					log.Printf("❌ ERROR: Failed to create/update admin in user-service: %v\n", result.Error)
				} else {
					log.Printf("✅ Created/updated admin in user-service: RowsAffected=%d\n", result.RowsAffected)
				}
			}
			
			// Verify user exists with detailed query
			var verifyUser UserServiceAdmin
			if err := userDB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", existingUser.ID).First(&verifyUser).Error; err != nil {
				log.Printf("⚠️  Warning: Admin user not found after creation (deleted_at IS NULL). Error: %v\n", err)
				// Try to find even if soft-deleted
				if err2 := userDB.WithContext(ctx).Unscoped().Where("id = ?", existingUser.ID).First(&verifyUser).Error; err2 != nil {
					log.Printf("❌ ERROR: Admin user not found even with Unscoped. Error: %v\n", err2)
				} else {
					log.Printf("⚠️  Found admin user but it's soft-deleted (deleted_at=%v)\n", verifyUser.DeletedAt)
				}
			} else {
				log.Printf("✅ Verified: Admin user exists in user-service with ID: %s, Email: %s\n", verifyUser.ID.String(), *verifyUser.Email)
			}
			
			// Also check by email
			var userByEmail UserServiceAdmin
			if err := userDB.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", adminEmail).First(&userByEmail).Error; err != nil {
				log.Printf("⚠️  Warning: Admin user not found by email (deleted_at IS NULL). Error: %v\n", err)
			} else {
				log.Printf("✅ Verified by email: Admin user exists with ID: %s\n", userByEmail.ID.String())
			}
		} else {
			log.Printf("⚠️  Warning: userDB is nil, skipping user-service creation\n")
		}

		log.Printf("✅ Updated admin user: %s (ID: %s)\n", adminEmail, existingUser.ID.String())
		log.Printf("   Admin Email: %s\n", adminEmail)
		log.Printf("   Admin Password: %s\n", adminPassword)
		log.Printf("   Admin User Type: %s\n", adminUserType)
		return nil
	}

	// Create new admin user
	adminID := uuid.New()
	firstNamePtr := &adminFirstName
	lastNamePtr := &adminLastName

	// Create admin in auth-service
	adminUser := AdminUser{
		ID:        adminID,
		Email:     &adminEmail,
		FirstName: firstNamePtr,
		LastName:  lastNamePtr,
		UserType:  adminUserType,
		Password:  &hashedPasswordStr,
		IsVerify:  true,
	}

	if err := authDB.WithContext(ctx).Create(&adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user in auth-service: %v", err)
	}
	log.Printf("✅ Created admin user in auth-service: %s (ID: %s)\n", adminEmail, adminID.String())

	// Create admin in user-service if connected - use raw SQL to ensure correct ID
	if userDB != nil {
		// Always delete any user with this email but different ID (including soft-deleted) to prevent conflicts
		var deletedCount int64
		// Try to delete addresses, but ignore error if table doesn't exist
		_ = userDB.WithContext(ctx).Unscoped().Exec("DELETE FROM user_addresses WHERE user_id IN (SELECT id FROM users WHERE email = $1 AND id != $2)", adminEmail, adminID)
		result := userDB.WithContext(ctx).Unscoped().Exec("DELETE FROM users WHERE email = $1 AND id != $2", adminEmail, adminID)
		if result.Error == nil {
			deletedCount = result.RowsAffected
			if deletedCount > 0 {
				log.Printf("⚠️  Deleted %d old admin user(s) with mismatched ID\n", deletedCount)
			}
		}
		
		// Always use raw SQL to ensure correct ID and handle password column
		var hasPassword bool
		userDB.WithContext(ctx).Raw(`
			SELECT COUNT(*) > 0 
			FROM information_schema.columns 
			WHERE table_name = 'users' AND column_name = 'password' AND table_schema = 'public'
		`).Scan(&hasPassword)
		
		log.Printf("📝 Inserting admin into user-service: ID=%s, Email=%s, HasPassword=%v\n", adminID.String(), adminEmail, hasPassword)
		
		if hasPassword {
			dummyPassword := "N/A"
			result := userDB.WithContext(ctx).Exec(`
				INSERT INTO users (id, email, first_name, last_name, user_type, password, created_at, updated_at, deleted_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW(), NULL)
				ON CONFLICT (id) DO UPDATE SET
					email = EXCLUDED.email,
					first_name = EXCLUDED.first_name,
					last_name = EXCLUDED.last_name,
					user_type = EXCLUDED.user_type,
					deleted_at = NULL,
					updated_at = NOW()
			`, adminID, adminEmail, adminFirstName, adminLastName, adminUserType, dummyPassword)
			if result.Error != nil {
				log.Printf("❌ ERROR: Failed to create/update admin in user-service: %v\n", result.Error)
			} else {
				log.Printf("✅ Created/updated admin in user-service: RowsAffected=%d\n", result.RowsAffected)
			}
		} else {
			result := userDB.WithContext(ctx).Exec(`
				INSERT INTO users (id, email, first_name, last_name, user_type, created_at, updated_at, deleted_at)
				VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), NULL)
				ON CONFLICT (id) DO UPDATE SET
					email = EXCLUDED.email,
					first_name = EXCLUDED.first_name,
					last_name = EXCLUDED.last_name,
					user_type = EXCLUDED.user_type,
					deleted_at = NULL,
					updated_at = NOW()
			`, adminID, adminEmail, adminFirstName, adminLastName, adminUserType)
			if result.Error != nil {
				log.Printf("❌ ERROR: Failed to create/update admin in user-service: %v\n", result.Error)
			} else {
				log.Printf("✅ Created/updated admin in user-service: RowsAffected=%d\n", result.RowsAffected)
			}
		}
		
		// Verify user exists with detailed query
		var verifyUser UserServiceAdmin
		if err := userDB.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", adminID).First(&verifyUser).Error; err != nil {
			log.Printf("⚠️  Warning: Admin user not found after creation (deleted_at IS NULL). Error: %v\n", err)
			// Try to find even if soft-deleted
			if err2 := userDB.WithContext(ctx).Unscoped().Where("id = ?", adminID).First(&verifyUser).Error; err2 != nil {
				log.Printf("❌ ERROR: Admin user not found even with Unscoped. Error: %v\n", err2)
			} else {
				log.Printf("⚠️  Found admin user but it's soft-deleted (deleted_at=%v)\n", verifyUser.DeletedAt)
			}
		} else {
			log.Printf("✅ Verified: Admin user exists in user-service with ID: %s, Email: %s\n", verifyUser.ID.String(), *verifyUser.Email)
		}
		
		// Also check by email
		var userByEmail UserServiceAdmin
		if err := userDB.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", adminEmail).First(&userByEmail).Error; err != nil {
			log.Printf("⚠️  Warning: Admin user not found by email (deleted_at IS NULL). Error: %v\n", err)
		} else {
			log.Printf("✅ Verified by email: Admin user exists with ID: %s\n", userByEmail.ID.String())
		}
	} else {
		log.Printf("⚠️  Warning: userDB is nil, skipping user-service creation\n")
	}

	log.Printf("   Admin Email: %s\n", adminEmail)
	log.Printf("   Admin Password: %s\n", adminPassword)
	log.Printf("   Admin User Type: %s\n", adminUserType)
	return nil
}

