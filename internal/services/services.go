package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/internal/domain/storage"
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

type ServiceUser struct {
	cfg          *config.Config
	log          *slog.Logger
	tokenService TokenService
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

func New(log *slog.Logger, usrSaver UserSaver, usrProv UserProvider, appProv AppProvider, tokenService TokenService, cfg *config.Config) *ServiceUser {
	return &ServiceUser{
		cfg:          cfg,
		log:          log,
		userSaver:    usrSaver,
		userProvider: usrProv,
		tokenService: tokenService,
		tokenTTL:     cfg.JWT.TokenTTL,
		appProvider:  appProv,
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

func (s *ServiceUser) RegisterNewUser(ctx context.Context, login, password, email, fullName string) (userId int64, err error) {
	const op = "services.RegisterNewUser"

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
	const op = "services.Login"

	log := s.log.With(slog.String("op", op), slog.String("login", login))
	log.Info("logging in")

	user, err := s.userProvider.UserByLogin(ctx, login)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", err)
			return "", ErrUserNotFound
		}
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", err)
			return "", ErrAppNotFound
		}

		log.Error("failed to login", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		log.Warn("invalid password", err)
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in")

	var jwtModel models.TokenModel
	app, err := s.appProvider.App(ctx, appId)
	if err != nil {
		log.Error("failed to get app", err.Error())
		return "", fmt.Errorf("%s: %w", op, err)
	}
	jwtModel.AppId = app.Id
	jwtModel.Secret = app.Secret
	jwtModel.Uid = user.ID
	jwtModel.IsAdmin = user.IsAdmin

	token, err = s.tokenService.NewToken(ctx, jwtModel)
	if err != nil {
		log.Error("failed to generate token", err.Error())
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil

}

func (s *ServiceUser) IsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error) {
	const op = "services.IsAdmin"

	log := s.log.With(slog.String("op", op), slog.Int64("user_id", userId))
	log.Info("checking user is admin")

	isAdmin, err = s.userProvider.UserIsAdmin(ctx, userId)
	if err != nil {
		log.Error("failed to check user is admin", err)
		return false, err
	}

	return isAdmin, nil
}

func (s *ServiceUser) UserInfo(ctx context.Context, userId int64) (models.UserServices, error) {
	const op = "services.UserInfo"

	log := s.log.With(slog.String("op", op), slog.Int64("user_id", userId))
	log.Info("getting user info")

	user, err := s.userProvider.UserById(ctx, userId)
	if err != nil {
		log.Error("failed to get user info", err)
		return models.UserServices{}, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("UserServices: ", slog.Int64("id", user.ID), slog.String("login", user.Login))

	return models.UserServices{
		ID:        user.ID,
		Login:     user.Login,
		Email:     user.Email,
		FullName:  user.FullName,
		IsAdmin:   user.IsAdmin,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *ServiceUser) DeleteUser(ctx context.Context, userId int64) (err error) {
	const op = "services.DeleteUser"

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
	const op = "services.UpdatePassword"

	log := s.log.With(slog.String("op", op))
	log.Info("updating user password")

	passHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", err)
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

func (s *ServiceUser) UsersAll(ctx context.Context) ([]models.UserServices, error) {
	const op = "services.UsersAll"

	log := s.log.With(slog.String("op", op))
	log.Info("getting all users")

	usersRepos, err := s.userProvider.UsersAll(ctx)
	if err != nil {
		log.Error("failed to get all users", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	usersServices := make([]models.UserServices, len(usersRepos))
	for i, userRepo := range usersRepos {
		usersServices[i] = models.UserServices{
			ID:        userRepo.ID,
			Login:     userRepo.Login,
			Email:     userRepo.Email,
			FullName:  userRepo.FullName,
			IsAdmin:   userRepo.IsAdmin,
			CreatedAt: userRepo.CreatedAt,
			UpdatedAt: userRepo.UpdatedAt,
		}
	}
	return usersServices, nil
}

// UpdateTokenApp generates a new token for the given appId.
//
// Note: The function signature was changed to accept appId as a parameter,
// instead of extracting it from context metadata. This change was made to
// improve clarity and security by making the required appId explicit.
// Callers must now provide appId directly when calling this function.
// If you previously relied on context metadata for appId, update your code
// to pass appId as an argument. This change may affect existing callers.
func (s *ServiceUser) UpdateTokenApp(ctx context.Context, token string, appId int32) (string, error) {
	const op = "services.UpdateTokenApp"
	log := s.log.With(slog.String("op", op))

	jwtModel, err := s.SignIn(ctx, token, appId)
	if err != nil {
		log.Warn("failed to sign in: ", err.Error())
		return "", err
	}

	var appSecret string
	if appId == appIdSSO {
		appSecret = s.cfg.JWT.Secret
	} else {
		app, err := s.appProvider.App(ctx, appId)
		if err != nil {
			log.Warn("failed to get app: ", err.Error())
			return "", ErrAppNotFound
		}
		appSecret = app.Secret
	}

	tokenNew, err := s.tokenService.NewToken(
		ctx,
		models.TokenModel{
			AppId:  appId,
			Uid:    jwtModel.Uid,
			Secret: appSecret,
		})
	if err != nil {
		log.Warn("failed to generate token: ", err.Error())
		return "", ErrValidToken
	}

	return tokenNew, nil

}

// SignIn Функция, которая отвечает за валидацию токена, создана, чтобы снять ответственность
// за валидацию токенов с прикладного слоя
func (s *ServiceUser) SignIn(ctx context.Context, token string, appId int32) (tokenModel models.TokenInfo, err error) {
	const op = "services.SignIn"
	log := s.log.With(slog.String("op", op), slog.Int64("app_id", int64(appId)))
	appModel, err := s.appProvider.App(ctx, appId)
	if err != nil {
		log.Error("failed to get app: ", err.Error())
		return models.TokenInfo{}, err
	}
	secret := appModel.Secret
	tokenModel, err = s.tokenService.Parse(token, secret)
	if err != nil {
		s.log.Error("failed to parse token: ", err.Error())
		return models.TokenInfo{}, err
	}

	log.Info("user sing in", slog.Any("models.TokenInfo:", tokenModel))
	return tokenModel, nil
}
