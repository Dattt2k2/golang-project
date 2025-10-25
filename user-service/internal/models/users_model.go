package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email     *string   `gorm:"unique;not null" json:"email" validate:"email,required"`
	FirstName *string   `gorm:"type:varchar(100)" json:"first_name"`
	LastName  *string   `gorm:"type:varchar(100)" json:"last_name"`
	UserType  *string   `gorm:"type:varchar(50);default:'USER'" json:"user_type" validate:"required,eq=ADMIN|eq=USER|eq=SELLER"`
	Phone     *string   `gorm:"type:varchar(15)" json:"phone"`
	Address   []UserAddress `gorm:"foreignKey:UserID" json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserAddress struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Street    string         `gorm:"type:varchar(100)" json:"street"`
	City      string         `gorm:"type:varchar(100)" json:"city"`
	State     string         `gorm:"type:varchar(100)" json:"state"`
	IsDefault bool           `gorm:"default:false" json:"is_default"`
}