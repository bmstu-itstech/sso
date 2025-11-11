package jwt

import (
	"fmt"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

func NewToken(user models.User, app models.App, tokenTTL time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = strconv.FormatInt(user.ID, 10)
	claims["login"] = user.Login
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(tokenTTL).Unix()
	claims["app_id"] = app.Id

	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func NewTokenSSO(user models.User, secret string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = strconv.FormatInt(user.ID, 10)

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseIdJwtToken(tokenString string, secret string) (int64, error) {
	tokenParsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("token is required")
	}
	userId, err := strconv.ParseInt(claims["uid"].(string), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("token is invalid")
	}
	return userId, nil
}
