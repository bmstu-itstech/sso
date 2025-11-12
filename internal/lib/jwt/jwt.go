package jwt

import (
	"errors"
	"fmt"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

var (
	ErrTokenNoValid = errors.New("token no valid")
	ErrTokenExpired = errors.New("token is expired")
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

func NewTokenSSO(user models.User, secret string, tokenTTL time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = strconv.FormatInt(user.ID, 10)
	claims["exp"] = time.Now().Add(tokenTTL).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseSSOJwtToken(tokenString string, secret string) (int64, error) {
	tokenParsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrTokenNoValid
	}

	unixTime := int64(claims["exp"].(float64))
	timeExp := time.Unix(unixTime, 0)
	fmt.Println(timeExp)

	if timeExp.Before(time.Now()) {
		return 0, ErrTokenExpired
	}

	userId, err := strconv.ParseInt(claims["uid"].(string), 10, 64)
	if err != nil {
		return 0, ErrTokenNoValid
	}

	return userId, nil
}
