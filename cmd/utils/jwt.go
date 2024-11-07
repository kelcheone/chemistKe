package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

var secretKey []byte

func int() {
	godotenv.Load()
	secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))
}

type jwtCustomClaims struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}

func CreateToken(id string, name string, role string) (string, error) {
	var admin bool
	if role == "USER" {
		admin = false
	} else {
		admin = true
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    id,
		"name":  name,
		"admin": admin,
		"exp":   time.Now().Add(time.Hour * 168).Unix(),
	})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ExtractClaims(tokenString string) (*jwtCustomClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwtCustomClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return secretKey, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Could not parse the token.")
	}

	if claims, ok := token.Claims.(*jwtCustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("Invalid JWT token")
	}
}

func ExtractClaimsFromRequest(c echo.Context) *jwtCustomClaims {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwtCustomClaims)

	return claims
}

func VerifyToken(tokenString string) error {
	token, err := jwt.Parse(
		tokenString,
		func(t *jwt.Token) (interface{}, error) {
			return secretKey, nil
		},
	)
	if err != nil {
		return err
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

var config = echojwt.Config{
	NewClaimsFunc: func(c echo.Context) jwt.Claims {
		return new(jwtCustomClaims)
	},

	SigningKey: secretKey,
}

var AuthMiddleware = echojwt.WithConfig(config)
