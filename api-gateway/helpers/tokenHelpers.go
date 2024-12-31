package helpers

import (
	"log"
	"os"
	"time"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

type SignedDetails struct {
    Email     string `json:"email"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    UserType  string `json:"user_type"` 
    Uid       string `json:"uid"`
    jwt.RegisteredClaims
}

var SECRECT_KEY string

// Hàm init để load SECRET_KEY từ .env
func InitDotEnv() {
	// Tải SECRET_KEY từ file .env
	err := godotenv.Load("./api-gateway/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Lấy giá trị SECRET_KEY từ biến môi trường
	SECRECT_KEY = os.Getenv("SECRET_KEY")
	if SECRECT_KEY == "" {
		log.Fatal("SECRET_KEY not found in .env")
	}
}

func ValidateToken(tokenString string) (*SignedDetails, string) {
    log.Printf("[Debug] Secret key length: %d", len(SECRECT_KEY))

    token, err := jwt.ParseWithClaims(
        tokenString,
        &SignedDetails{},
        func(token *jwt.Token) (interface{}, error) {
            // Verify signing method
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                log.Printf("[Error] Expected HMAC signing method but got %v", token.Header["alg"])
                return nil, fmt.Errorf("invalid signing method")
            }
            
            // Debug signing key
            signingKey := []byte(SECRECT_KEY)
            log.Printf("[Debug] Using signing key length: %d", len(signingKey))
            
            return signingKey, nil
        },
    )

    if err != nil {
        ve, ok := err.(*jwt.ValidationError)
        if ok {
            log.Printf("[Debug] Validation error type: %d", ve.Errors)
            log.Printf("[Debug] Raw error message: %s", ve.Error())

            switch {
            case ve.Errors&jwt.ValidationErrorSignatureInvalid != 0:
                log.Printf("[Error] Signature validation failed")
                return nil, "invalid signature"
            case ve.Errors&jwt.ValidationErrorExpired != 0:
                return nil, "token expired"
            case ve.Errors&jwt.ValidationErrorMalformed != 0:
                return nil, "token malformed"
            default:
                return nil, fmt.Sprintf("token validation error: %v", ve.Error())
            }
        }
        return nil, "token validation failed"
    }

    claims, ok := token.Claims.(*SignedDetails)
    if !ok || !token.Valid {
        log.Println("[Error] Invalid claims or token")
        return nil, "invalid token claims"
    }

    log.Printf("[Debug] Token validated successfully for: %s", claims.Email)
    return claims, ""
}


func GenerateAllToken(email, firstName, lastName, userType, uid string) (signedToken string, signedRefreshToken string, err error) {
    claims := &SignedDetails{
        Email:     email,
        FirstName: firstName,
        LastName:  lastName,
        UserType:  userType,
        Uid:       uid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
    }

    refreshClaims := &SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

    signedToken, err = token.SignedString([]byte(SECRECT_KEY))
    if err != nil {
        return
    }

    signedRefreshToken, err = refreshToken.SignedString([]byte(SECRECT_KEY))
    if err != nil {
        return
    }

    return signedToken, signedRefreshToken, nil
}