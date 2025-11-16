package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/internal/domain/storage"
	"github.com/bmstu-itstech/sso/internal/lib/jwt"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/metadata"
	"log/slog"
	"math/rand"
	"strings"
	"time"
)

var (
	ErrAppNotFoud    = errors.New("app not found")
	ErrUserNotFound  = errors.New("user not found")
	ErrNoToken       = errors.New("you have not jwt token")
	ErrValidToken    = errors.New("token is not valid")
	ErrUserNotUnique = errors.New("user with this login already exists")
)

const (
	appIdSSO = 0
)

type ServiceUser struct {
	cfg          *config.Config
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

func New(log *slog.Logger, usrSaver UserSaver, usrProv UserProvider, appProv AppProvider, cfg *config.Config) *ServiceUser {
	return &ServiceUser{
		cfg:          cfg,
		log:          log,
		userSaver:    usrSaver,
		userProvider: usrProv,
		tokenTTL:     cfg.JWT.TokenTTL,
		appProvider:  appProv,
	}
}

type UserSaver interface {
	SaveUser(ctx context.Context, login string, password []byte, email string, fullName string, userId int64) (err error)
	UserNewPassword(ctx context.Context, userId int64, newPassword []byte) (err error)
	UserDelete(ctx context.Context, userId int64) (err error)
}

type UserProvider interface {
	UserByLogin(ctx context.Context, login string) (user models.User, err error)
	UserById(ctx context.Context, userId int64) (user models.User, err error)
	UserIsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error)
	UsersAll(ctx context.Context) (users []models.User, err error)
}

type AppProvider interface {
	App(ctx context.Context, appId int32) (models.App, error)
}

func (s *ServiceUser) RegisterNewUser(ctx context.Context, login, password, email, fullName string) (userId int64, err error) {
	const op = "auth.RegisterNewUser"

	log := s.log.With(slog.String("op", op), slog.String("login", login))
	log.Info("registering user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to create password", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id := rand.Int63()
	log.Info("create user id", slog.Int64("user_id", id))

	if err = s.userSaver.SaveUser(ctx, login, passHash, email, fullName, id); err != nil {
		if errors.Is(err, storage.ErrUserNotUnique) {
			return 0, ErrUserNotUnique
		}
		log.Error("failed to save user", err)
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *ServiceUser) Login(ctx context.Context, appId int32, login string, password string) (token string, err error) {
	const op = "auth.Login"
	log := s.log.With(slog.String("op", op), slog.String("login", login))

	log.Info("logging in")

	user, err := s.userProvider.UserByLogin(ctx, login)
	fmt.Println(user)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", err)
			return "", ErrUserNotFound
		}
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", err)
			return "", ErrAppNotFoud
		}

		log.Error("failed to login", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Warn("invalid password", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in")

	if appId == appIdSSO {
		token, err = jwt.NewToken(user, models.App{Id: 0, Secret: s.cfg.JWT.Secret}, s.cfg.JWT.TokenTTL)
		fmt.Println(token)
		return token, err
	}

	app, err := s.appProvider.App(ctx, appId)
	if err != nil {
		log.Error("failed to get app", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err = jwt.NewToken(user, app, s.tokenTTL)
	if err != nil {
		log.Error("failed to generate token", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil

}

func (s *ServiceUser) IsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error) {
	const op = "auth.IsAdmin"

	log := s.log.With(slog.String("op", op), slog.Int64("user_id", userId))

	log.Info("checking user is admin")

	isAdmin, err = s.userProvider.UserIsAdmin(ctx, userId)
	// Поведение без ошибок
	if err != nil {
		log.Error("failed to check user is admin", err)
		return false, nil
	}

	return isAdmin, nil
}

func (s *ServiceUser) GetUserInfo(ctx context.Context, userId int64) (user models.User, err error) {
	const op = "auth.GetUserInfo"

	log := s.log.With(slog.String("op", op), slog.Int64("user_id", userId))
	log.Info("getting user info")

	user, err = s.userProvider.UserById(ctx, userId)
	log.Info("User: ", user)
	if err != nil {
		log.Error("failed to get user info", err)
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *ServiceUser) DeleteUser(ctx context.Context, userId int64) (err error) {
	const op = "auth.DeleteUser"
	log := s.log.With(slog.String("op", op))
	log.Info("deleting user")

	err = s.userSaver.UserDelete(ctx, userId)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", err)
			return ErrUserNotFound
		}
		log.Error("failed to delete user", err)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *ServiceUser) UpdatePassword(ctx context.Context, userId int64, newPassword string) (err error) {
	const op = "auth.UpdatePassword"
	log := s.log.With(slog.String("op", op))
	log.Info("updating user password")

	passHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to geterate password hash", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.userSaver.UserNewPassword(ctx, userId, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", err)
			return ErrUserNotFound
		}
		log.Error("failed to update user password", err)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *ServiceUser) GetAllUsers(ctx context.Context) (users []models.User, err error) {
	const op = "auth.GetAllUsers"
	log := s.log.With(slog.String("op", op))
	log.Info("getting all users")

	users, err = s.userProvider.UsersAll(ctx)
	if err != nil {
		log.Error("failed to get all users", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return users, nil
}

func (s *ServiceUser) UpdateTokenApp(ctx context.Context) (string, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		s.log.Warn("JWT token not found")
		return "", ErrNoToken
	}

	jwtString := authHeader[0]
	jwtString = strings.TrimPrefix(jwtString, "Bearer ")
	appId, err := jwt.ParseAppId(jwtString)
	if err != nil {
		s.log.Warn("failed to parse token: %w", err)
		return "", ErrValidToken
	}

	app, err := s.appProvider.App(ctx, appId)
	if err != nil {
		s.log.Warn("failed to get app: %w", err)
		return "", ErrValidToken
	}

	tokenMap, err := jwt.ParseTokenApp(jwtString, app.Secret)
	if err != nil {
		s.log.Warn("failed to parse token: %w", err)
		return "", ErrValidToken
	}

	tokenNew, err := jwt.NewToken(
		models.User{ID: tokenMap.Uid},
		app, s.tokenTTL)

	if err != nil {
		s.log.Warn("failed to generate token: %w", err)
		return "", ErrValidToken
	}

	return tokenNew, nil

}
