package utils

import (
	"hello-fiber/app/model"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "your-secret-key"
}

func GetJWTSecret() []byte {
	return jwtSecret
}

type Claims struct {
	UserID string `json:"user_id"` // Using json tags (not bson) because JWT is JSON Web Token
	Email  string `json:"email"`
	RoleID string `json:"role_id"` // Changed from int to string to store ObjectID hex
	jwt.RegisteredClaims
}

// // GenerateJWT generates a JWT token for authenticated user
// func GenerateJWT(user model.User) (string, error) {
// 	uidStr := user.ID.Hex() // Convert ObjectID to hex string
// 	roleIDStr := user.RoleID.Hex() // Convert RoleID ObjectID to hex string
// 	claims := Claims{
// 		UserID: uidStr,
// 		Email:  user.Email,
// 		RoleID: roleIDStr,
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
// 			IssuedAt:  jwt.NewNumericDate(time.Now()),
// 			Subject:   uidStr,
// 		},
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	return token.SignedString(jwtSecret)
// }

// GenerateJWTPostgres generates a JWT token for user from PostgreSQL
func GenerateJWTPostgres(user *model.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		Email:  user.Email,
		RoleID: user.RoleID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
