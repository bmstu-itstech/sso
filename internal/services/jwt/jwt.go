package jwt

import (
	"context"
	"errors"
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
	_ = 0
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["uid"] = strconv.FormatInt(jwtModel.Uid, 10)
	claims["exp"] = time.Now().Add(s.TokenTTL).Unix()
	claims["app_id"] = jwtModel.AppId

	tokenString, err := token.SignedString([]byte(jwtModel.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
