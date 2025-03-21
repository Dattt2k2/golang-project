package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    First_name     *string            `json:"first_name" validate:"required,min=2,max=100"`
    Last_name      *string            `json:"last_name" validate:"required,min=2,max=100"`
    Password       *string            `json:"password" validate:"required,min=6"`
    Email          *string            `json:"email" validate:"email,required"`
    Phone          *string            `json:"phone" validate:"required"`
    // Token          *string            `json:"token,omitempty"`
    User_type      *string            `json:"user_type" validate:"required,eq=ADMIN|eq=USER|eq=SELLER"`
    // Refresh_token  *string            `json:"refresh_token,omitempty"`
    Created_at     time.Time          `json:"created_at,omitempty"`
    Updated_at     time.Time          `json:"updated_at,omitempty"`
    User_id        string             `json:"user_id,omitempty"`
}

type LoginResponse struct {
	Email         string `json:"email"`
	First_name     string `json:"first_name"`
	Last_name      string `json:"last_name"`
	User_type      string `json:"user_type"`
    Password       *string            `json:"password"`
	User_id        string `json:"user_id"`
	Token         string `json:"token"`
	RefreshToken  string `json:"refresh_token"`
}




// type User struct{
// 	ID				primitive.ObjectID 		`bson:"_id"`
// 	First_name		*string					`json:"first_name" validate:"required, min=2, max=100"`
// 	Last_name		*string					`json:"last_name" validate:"required, min=2, max=100"`
// 	Password		*string					`json:"password" validate:"required, min=6"`
// 	Email			*string					`json:"email" validate:"email, required"`
// 	Phone			*string					`json:"phone" validate:"required"`
// 	Token			*string					`json:"token"`
// 	User_type 		*string					`json:"user_type" validate:"required, eq=ADMIN|eq=USER"`
// 	Refresh_token	*string					`json:"refresh_token"`
// 	Created_at		time.Time				`json:"created_at"`
// 	Updated_at		time.Time				`json:"updated_at"`
// 	User_id			string					`json:"user_id"`
// }