package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
)

var (
	ErrAppNotFound   = errors.New("app not found")
	ErrUserNotFound  = errors.New("user not found")
	ErrValidToken    = errors.New("token is not valid")
	ErrUserNotUnique = errors.New("user with this login already exists")
	ErrTokenExpired  = errors.New("token is expired")
)

const (
	appIdSSO = 0
)

type Service struct {
	cfg          *config.Config
	log          *slog.Logger
	tokenService TokenService
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	eventService EventService
	tokenTTL     time.Duration
}

func New(log *slog.Logger, usrSaver UserSaver, usrProv UserProvider, appProv AppProvider, tokenService TokenService, eventService EventService, cfg *config.Config) *Service {
	return &Service{
		cfg:          cfg,
		log:          log,
		userSaver:    usrSaver,
		userProvider: usrProv,
		tokenService: tokenService,
		tokenTTL:     cfg.JWT.TokenTTL,
		appProvider:  appProv,
		eventService: eventService,
	}
}

type TokenService interface {
	NewToken(ctx context.Context, model models.TokenModel) (token string, err error)
	Parse(tokenString string, secret string) (app models.TokenInfo, err error)
}

type UserSaver interface {
	SaveUser(ctx context.Context, login string, password []byte, email string, fullName string, userId int64) (err error)
	UserNewPassword(ctx context.Context, userId int64, newPassword []byte) (err error)
	UserDelete(ctx context.Context, userId int64) (err error)
}

type UserProvider interface {
	UserByLogin(ctx context.Context, login string) (user models.UserRepository, err error)
	UserById(ctx context.Context, userId int64) (user models.UserRepository, err error)
	UserIsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error)
	UsersAll(ctx context.Context) (users []models.UserRepository, err error)
}

type AppProvider interface {
	App(ctx context.Context, appId int32) (models.AppRepos, error)
}

type EventService interface {
	HasUserChanges(ctx context.Context, userID int64) (bool, error)
	PublishUserUpdated(ctx context.Context, userID int64) error
}
