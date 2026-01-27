package jwt

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
)

func NewServiceJwt(ttl time.Duration) *ServiceJwt {
	return &ServiceJwt{TokenTTL: ttl}
}

type ServiceJwt struct {
	TokenTTL time.Duration
}

// NewToken generate TokenJWT and have option params, secret, if func have params, generate Token
func (s *ServiceJwt) NewToken(ctx context.Context, jwtModel models.TokenModel) (string, error) {
	if jwtModel.Secret == "" {
		return "", errors.New("secret is empty")
	}
	if jwtModel.Uid == 0 {
		return "", errors.New("user id is empty")
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["uid"] = strconv.FormatInt(jwtModel.Uid, 10)
	claims["exp"] = time.Now().Add(s.TokenTTL).Unix()
	claims["app_id"] = jwtModel.AppId
	claims["is_admin"] = jwtModel.IsAdmin

	tokenString, err := token.SignedString([]byte(jwtModel.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *ServiceJwt) Parse(tokenString string, secret string) (models.TokenInfo, error) {
	fmt.Println(tokenString)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return models.TokenInfo{}, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return models.TokenInfo{}, errors.New("invalid token")
	}
	if claims["exp"] == nil {
		return models.TokenInfo{}, errors.New("token has no expiration date")
	}
	uid, ok := claims["uid"].(string)
	if !ok {
		return models.TokenInfo{}, errors.New("invalid uid")
	}
	uidInt64, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		return models.TokenInfo{}, err
	}

	isAdmin, ok := claims["is_admin"].(bool)
	if !ok {
		return models.TokenInfo{}, errors.New("invalid is_admin")
	}

	appId, ok := claims["app_id"].(int32)

	return models.TokenInfo{
		Uid:     uidInt64,
		IsAdmin: isAdmin,
		AppId:   appId,
	}, nil
}
