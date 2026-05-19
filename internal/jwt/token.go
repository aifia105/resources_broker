package jwt

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrEmptySecret   = "JWT secret key is not set in environment variables"
	ErrEmptyExpiry   = "JWT expiry time is not set in environment variables"
	ErrExpiredToken  = "JWT token has expired"
	ErrInvalidToken  = "JWT token is invalid"
	ErrInvalidExpiry = "JWT token has invalid expiry time"
	ErrTokenNotFound = "JWT token not found in request"
	ErrInvalidClaims = "JWT token has invalid claims"
)

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GetTokens(user_id bson.ObjectID) (*Token, error) {
	accessToken, err := GenerateAccessToken(user_id)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := GenerateRefreshToken(user_id)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &Token{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func GenerateToken(user_id bson.ObjectID, secret string, expiryVar string) (string, error) {
	envSecret := os.Getenv(secret)
	if envSecret == "" {
		return "", errors.New(ErrEmptySecret)
	}

	expiryStr := os.Getenv(expiryVar)
	if expiryStr == "" {
		return "", errors.New(ErrEmptyExpiry)
	}

	tokenLifetime, err := time.ParseDuration(expiryStr)
	if err != nil {
		return "", errors.New(ErrInvalidExpiry)
	}

	claims := jwt.MapClaims{
		"authorized": true,
		"user_id":    user_id.Hex(),
		"exp":        time.Now().Add(tokenLifetime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(envSecret))
}

func GenerateRefreshToken(user_id bson.ObjectID) (string, error) {
	return GenerateToken(user_id, "REFRESH_TOKEN_SECRET", "REFRESH_TOKEN_EXPIRY")
}

func GenerateAccessToken(user_id bson.ObjectID) (string, error) {
	return GenerateToken(user_id, "ACCESS_TOKEN_SECRET", "ACCESS_TOKEN_EXPIRY")
}

func ValidToken(c *gin.Context) error {
	tokenString := ExtractToken(c)
	if tokenString == "" {
		return errors.New(ErrTokenNotFound)
	}

	envSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	if envSecret == "" {
		return errors.New(ErrEmptySecret)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(ErrInvalidToken)
		}
		return []byte(envSecret), nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return errors.New(ErrInvalidToken)
	}

	return nil
}

func ExtractToken(c *gin.Context) string {
	token, err := c.Cookie("access_token")
	if err == nil && token != "" {
		return token
	}

	bearerToken := c.Request.Header.Get("Authorization")
	if bearerToken != "" {
		parts := strings.Split(bearerToken, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	return ""
}

func ExtractTokenID(c *gin.Context) (bson.ObjectID, error) {
	tokenString := ExtractToken(c)
	if tokenString == "" {
		return bson.ObjectID{}, errors.New(ErrTokenNotFound)
	}

	secret := os.Getenv("ACCESS_TOKEN_SECRET")
	if secret == "" {
		return bson.ObjectID{}, errors.New(ErrEmptySecret)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(ErrInvalidToken)
		}
		return []byte(secret), nil
	})
	if err != nil {
		return bson.ObjectID{}, err
	}

	if !token.Valid {
		return bson.ObjectID{}, errors.New(ErrInvalidToken)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return bson.ObjectID{}, errors.New(ErrInvalidClaims)
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return bson.ObjectID{}, fmt.Errorf("%s: user_id not found or invalid type", ErrInvalidClaims)
	}

	objectID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		return bson.ObjectID{}, fmt.Errorf("%s: invalid ObjectID format", ErrInvalidClaims)
	}
	return objectID, nil
}

func ExtractRefreshToken(c *gin.Context) string {
	token, err := c.Cookie("refresh_token")
	if err == nil && token != "" {
		return token
	}

	bearerToken := c.Request.Header.Get("X-Refresh-Token")
	if bearerToken != "" {
		return bearerToken
	}

	return ""
}

func ValidateRefreshToken(tokenString string) (bson.ObjectID, error) {
	if tokenString == "" {
		return bson.ObjectID{}, errors.New(ErrTokenNotFound)
	}

	secret := os.Getenv("REFRESH_TOKEN_SECRET")
	if secret == "" {
		return bson.ObjectID{}, errors.New(ErrEmptySecret)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(ErrInvalidToken)
		}
		return []byte(secret), nil
	})
	if err != nil {
		return bson.ObjectID{}, err
	}

	if !token.Valid {
		return bson.ObjectID{}, errors.New(ErrInvalidToken)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return bson.ObjectID{}, errors.New(ErrInvalidClaims)
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return bson.ObjectID{}, fmt.Errorf("%s: user_id not found or invalid type", ErrInvalidClaims)
	}

	objectID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		return bson.ObjectID{}, fmt.Errorf("%s: invalid ObjectID format", ErrInvalidClaims)
	}
	return objectID, nil
}
