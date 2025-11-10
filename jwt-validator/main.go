package jwt_validator

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Secret    string `json:"secret"`
	HeaderName string `json:"headerName"`
}

func CreateConfig() *Config {
	return &Config{
		Secret:    os.Getenv("JWT_SECRET"),
		HeaderName: "X-User-ID",
	}
}


type JWTValidator struct {
    next       http.Handler
    secret     string
    headerName string
    name       string
}

// New tạo một instance mới của plugin
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
    if len(config.Secret) == 0 {
        return nil, errors.New("secret key cannot be empty")
    }
    if len(config.HeaderName) == 0 {
        return nil, errors.New("header name cannot be empty")
    }

    return &JWTValidator{
        next:       next,
        secret:     config.Secret,
        headerName: config.HeaderName,
        name:       name,
    }, nil
}

func (p *JWTValidator) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    authHeader := req.Header.Get("Authorization")
    if authHeader == "" {
        http.Error(rw, "Missing Authorization header", http.StatusUnauthorized)
        return
    }

    tokenString := strings.TrimPrefix(authHeader, "Bearer ")
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(p.secret), nil
    })

    if err != nil || !token.Valid {
        http.Error(rw, "Invalid token", http.StatusUnauthorized)
        return
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        http.Error(rw, "Invalid token claims", http.StatusUnauthorized)
        return
    }

    userID, ok := claims["uid"].(string)
    if !ok {
        http.Error(rw, "Missing user_id in token", http.StatusUnauthorized)
        return
    }

    req.Header.Set(p.headerName, userID)

    p.next.ServeHTTP(rw, req)
}