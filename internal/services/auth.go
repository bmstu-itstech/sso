package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/internal/domain/storage"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) RegisterNewUser(ctx context.Context, login, password, email, fullName string) (userId int64, err error) {
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

func (s *Service) Login(ctx context.Context, appId int32, login string, password string) (token string, err error) {
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

// SignIn Функция, которая отвечает за валидацию токена, создана, чтобы снять ответственность
// за валидацию токенов с прикладного слоя
func (s *Service) SignIn(ctx context.Context, token string, appId int32) (tokenModel models.TokenInfo, err error) {
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

func (s *Service) VerifyToken(ctx context.Context, req *models.VerifyTokenServiceRequest) (models.VerifyTokenServiceResponse, error) {
	const op = "services.VerifyToken"
	log := s.log.With(slog.String("op", op))
	app, err := s.appProvider.App(ctx, req.AppId)
	if err != nil {
		log.Error("failed to get app: ", err.Error())
		return models.VerifyTokenServiceResponse{}, err
	}

	modelToken, err := s.tokenService.Parse(req.Token, app.Secret)
	if err != nil {
		log.Error("failed to parse token: ", err.Error())
		return models.VerifyTokenServiceResponse{}, err
	}

	return models.VerifyTokenServiceResponse{
		TokenInfo: modelToken,
	}, nil
}

// UpdateTokenApp generates a new token for the given appId.
func (s *Service) UpdateTokenApp(ctx context.Context, token string, appId int32) (string, error) {
	const op = "services.UpdateTokenApp"
	log := s.log.With(slog.String("op", op))

	jwtModel, err := s.SignIn(ctx, token, appId)
	if err != nil {
		log.Error("failed to sign in: ", err.Error())
		return "", err
	}

	app, err := s.appProvider.App(ctx, appId)
	if err != nil {
		log.Error("failed to get app: ", err.Error())
		return "", ErrAppNotFound
	}

	tokenNew, err := s.tokenService.NewToken(
		ctx,
		models.TokenModel{
			AppId:  appId,
			Uid:    jwtModel.Uid,
			Secret: app.Secret,
		})
	if err != nil {
		log.Error("failed to generate token: ", err.Error())
		return "", ErrValidToken
	}

	return tokenNew, nil
}
