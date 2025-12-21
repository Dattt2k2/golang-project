package main

import (
	"log"
	"os"

	"auth-service/database/migration"
)

func main() {
	log.Println("Running admin user migration...")

	if err := migration.CreateAdminUserMigration(); err != nil {
		log.Fatalf("Migration failed: %v", err)
		os.Exit(1)
	}

	log.Println("✅ Migration completed successfully")
}

