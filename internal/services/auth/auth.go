package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/internal/lib"
	"github.com/bmstu-itstech/sso/internal/lib/jwt"
	"github.com/bmstu-itstech/sso/internal/services/storage"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

var (
	errAppNotFoud = errors.New("app not found")
)

type Auth struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

type UserSaver interface {
	SaveUser(ctx context.Context, login string, password []byte, email string, fullName string, userId int64) (err error)
}

type UserProvider interface {
	User(ctx context.Context, login string) (user models.User, err error)
	UserIsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error)
}

type AppProvider interface {
	App(ctx context.Context, appId int32) (models.App, error)
}

func New(log *slog.Logger, usrSaver UserSaver, usrProv UserProvider, appProv AppProvider, tokenTTL time.Duration) *Auth {
	return &Auth{
		log:          log,
		userSaver:    usrSaver,
		userProvider: usrProv,
		tokenTTL:     tokenTTL,
		appProvider:  appProv,
	}
}

func (a *Auth) RegisterNewUser(ctx context.Context, login, password, email, fullName string) (userId int64, err error) {
	const op = "auth.RegisterNewUser"
	log := a.log.With(slog.String("op", op), slog.String("login", login))
	log.Info("registering user")
	// TODO: что по безопасности
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to geterate password hash", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	id := lib.RandoInt64()
	log.Info("generated user id", slog.Int64("user_id", id))

	if err = a.userSaver.SaveUser(ctx, login, passHash, email, fullName, id); err != nil {
		log.Error("failed to save user", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (a *Auth) Login(ctx context.Context, appId int32, login string, password string) (token string, err error) {
	const op = "auth.Login"
	log := a.log.With(slog.String("op", op), slog.String("login", login))

	log.Info("logging in")

	user, err := a.userProvider.User(ctx, login)
	fmt.Println(user)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", err)
			return "", fmt.Errorf("%s: %w", op, err)
		}

		log.Error("failed to login", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Warn("invalid password", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in")

	app, err := a.appProvider.App(ctx, appId)
	if err != nil {
		log.Error("failed to get app", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err = jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil

}

func (a *Auth) IsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error) {
	const op = "auth.IsAdmin"
	log := a.log.With(slog.String("op", op), slog.Int64("user_id", userId))

	log.Info("checking user is admin")

	isAdmin, err = a.userProvider.UserIsAdmin(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("user not found", err)
			return false, fmt.Errorf("%s: %w", op, errAppNotFoud)
		}
		log.Error("failed to check user is admin", err)
		return false, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}
