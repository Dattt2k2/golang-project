package migration

import (
	"log"
	"payment-service/models"

	"gorm.io/gorm"
)

// RunMigrations executes all database migrations
func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Auto-migrate all models
	err := db.AutoMigrate(
		&models.Payment{},
		&models.Refund{},
		&models.Transaction{},
		&models.VendorAccount{},
		&models.VendorPayout{},
		&models.VendorBalance{},
		&models.VendorTransaction{},
	)

	if err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}
