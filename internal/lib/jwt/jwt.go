package jwt

import (
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func NewToken(user models.User, app models.App, tokenTTL time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
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
