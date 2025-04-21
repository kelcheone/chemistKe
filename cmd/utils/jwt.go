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

func init() {
	godotenv.Load()
	secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

	if len(secretKey) == 0 {
		panic("JWT_SECRET_KEY environment variable is not set")
	}
}

type jwtCustomClaims struct {
	Id     string `json:"id"`
	Name   string `json:"name" example:"John Doe" binding:"required"`
	Email  string `json:"email" example:"john.doe@example.com" binding:"required,email"`
	Phone  string `json:"phone" example:"+1234567890" binding:"required,phone"`
	Admin  bool   `json:"admin"`
	Author bool   `json:"author"`
	jwt.RegisteredClaims
}

func CreateToken(id string, email string, name string, phone string, role string) (string, error) {
	var admin bool
	var author bool

	switch role {
	case "USER":
		admin = false
	case "AUTHOR":
		author = true
	default:
		admin = true
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":     id,
		"name":   name,
		"email":  email,
		"phone":  phone,
		"admin":  admin,
		"author": author,
		"exp":    time.Now().Add(time.Hour * 168).Unix(),
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
		return nil, fmt.Errorf("could not parse the token")
	}

	if claims, ok := token.Claims.(*jwtCustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid JWT token")
	}
}

func ExtractClaimsFromRequest(c echo.Context) *jwtCustomClaims {
	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		return nil
	}
	claims, ok := user.Claims.(*jwtCustomClaims)
	if !ok {
		return nil
	}
	return claims
}

// extract claims from Cookie
func ExtractClaimsFromCookie(c echo.Context) *jwtCustomClaims {
	cookie, err := c.Cookie("token")
	if err != nil {
		return nil
	}
	tokenString := cookie.Value
	token, err := jwt.ParseWithClaims(
		tokenString,
		&jwtCustomClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return secretKey, nil
		},
	)
	if err != nil {
		return nil
	}

	if claims, ok := token.Claims.(*jwtCustomClaims); ok && token.Valid {
		return claims
	} else {
		return nil
	}
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
