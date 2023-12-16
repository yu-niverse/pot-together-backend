package auth

import (
	"pottogether/config"
	"pottogether/pkg/errhandler"
	"pottogether/pkg/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

var jwtSecretKey []byte

type authClaims struct {
	UserID int `json:"userID"`
	jwt.StandardClaims
}

func SetJWTKey() {
	jwtSecretKey = []byte(config.Viper.GetString("JWT_SECRET_KEY"))
}

func GenerateToken(userID int, email string) (string, error) {
	// Set JWT claims fields
	expiresAt := time.Now().Add(24 * time.Hour).Unix() // 24 hours
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, authClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			Subject:   email,
			ExpiresAt: expiresAt,
		},
	})
	// Sign the token with our secret key
	tokenString, err := token.SignedString(jwtSecretKey)
	if err != nil {
		logger.Error("[AUTH] Error generating token: " + err.Error())
		return "", err
	}
	logger.Info("[AUTH] Generated token for user: " + email)
	return tokenString, nil
}

// Authentication middleware
func ValidateToken(c *gin.Context) {
	// Get token from header
	auth := c.GetHeader("Authorization")
	if auth == "" {
		errhandler.Unauthorized(c, nil, "Missing token")
		c.Abort()
		return
	}
	token := strings.Split(auth, "Bearer ")[1]
	// Parse token
	tokenClaims, err := jwt.ParseWithClaims(token, &authClaims{}, func(token *jwt.Token) (i interface{}, err error) {
		return jwtSecretKey, nil
	})
	// Check for token validation errors
	if err != nil {
		errhandler.Unauthorized(c, err, "Invalid token")
		c.Abort()
		return
	}
	// Check if token is valid -> continue
	if claims, ok := tokenClaims.Claims.(*authClaims); ok && tokenClaims.Valid {
		c.Set("id", claims.UserID)
		c.Next()
	} else {
		c.Abort()
		return
	}
}
