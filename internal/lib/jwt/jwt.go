package jwt

import (
	"errors"
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

func NewTokenSSO(userId int64, secret string, tokenTTL time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = strconv.FormatInt(userId, 10)
	claims["exp"] = time.Now().Add(tokenTTL).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ParseTokenSSO(tokenString string, secret string) (int64, error) {
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

	if timeExp.Before(time.Now()) {
		return 0, ErrTokenExpired
	}

	userId, err := strconv.ParseInt(claims["uid"].(string), 10, 64)
	if err != nil {
		return 0, ErrTokenNoValid
	}

	return userId, nil
}

func PaseTokenApp(tokenString string, appSecret string) (models.AppJWT, error) {
	tokenParsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	if err != nil {
		return models.AppJWT{}, err
	}

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	if !ok {
		return models.AppJWT{}, ErrTokenNoValid
	}

	unixTime := int64(claims["exp"].(float64))
	timeExp := time.Unix(unixTime, 0)

	if timeExp.Before(time.Now()) {
		return models.AppJWT{}, ErrTokenExpired
	}

	userId, err := strconv.ParseInt(claims["uid"].(string), 10, 64)
	if err != nil {
		return models.AppJWT{}, ErrTokenNoValid
	}
	return models.AppJWT{
		Uid:   userId,
		Login: claims["login"].(string),
		Email: claims["email"].(string),
		AppId: int32(claims["app_id"].(float64)),
	}, nil
}

func ParseAppId(tokenString string) (int32, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())

	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if appID, exists := claims["app_id"]; exists {
			return int32(appID.(float64)), nil
		}
	}

	return 0, errors.New("app_id claim not found")
}
