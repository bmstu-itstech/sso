package jwt

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"

	"github.com/bmstu-itstech/sso/internal/domain/models"
)

var (
	ErrTokenSpecified = errors.New("token specified")
	ErrTokenEmpty     = errors.New("token empty")
	ErrTokenNoValid   = errors.New("token no valid")
	ErrTokenExpired   = errors.New("token is expired")
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

	tokenString, err := token.SignedString([]byte(jwtModel.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *ServiceJwt) ParseTokenApp(ctx context.Context, appSecret string) (models.TokenInfo, error) {
	tokenJwt, err := s.getToken(ctx)
	if err != nil {
		return models.TokenInfo{}, err
	}

	tokenParsed, err := jwt.Parse(tokenJwt, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	if err != nil {
		return models.TokenInfo{}, ErrTokenNoValid
	}

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	if !ok {
		return models.TokenInfo{}, ErrTokenNoValid
	}

	unixTime := int64(claims["exp"].(float64))
	timeExp := time.Unix(unixTime, 0)
	if timeExp.Before(time.Now()) {
		return models.TokenInfo{}, ErrTokenExpired
	}

	userId, err := strconv.ParseInt(claims["uid"].(string), 10, 64)
	if err != nil {
		return models.TokenInfo{}, ErrTokenNoValid
	}
	return models.TokenInfo{
		Uid:   userId,
		AppId: int32(claims["app_id"].(float64)),
	}, nil
}

func (s *ServiceJwt) getToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrTokenSpecified
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", ErrTokenSpecified
	}

	jwtString := authHeaders[0]
	jwtString = strings.TrimPrefix(jwtString, "Bearer ")
	if jwtString == "" {
		return "", ErrTokenEmpty
	}
	return jwtString, nil
}
