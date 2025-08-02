package models

import (
	"time"

)


type User struct {
    ID         string     `gorm:"type:uuid;primaryKey"`
    FirstName  *string    `gorm:"type:varchar(100)" json:"first_name"`
    LastName   *string    `gorm:"type:varchar(100)" json:"last_name"`
    Password   *string    `gorm:"not null" json:"password" validate:"required,min=6"`
    Email      *string    `gorm:"unique;not null" json:"email" validate:"email,required"`
    Phone      *string    `gorm:"type:varchar(15)" json:"phone"`
    Role       *string    `gorm:"type:varchar(50);default:'USER'" json:"role"`
    CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
    UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
    UserID     string     `gorm:"type:varchar(50)" json:"user_id"`
    IsVerify   bool       `gorm:"default:false" json:"is_verify"`
}