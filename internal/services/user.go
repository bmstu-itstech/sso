package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/internal/domain/storage"
	"golang.org/x/crypto/bcrypt"
)

func (s *Service) IsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error) {
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

func (s *Service) UserInfo(ctx context.Context, userId int64) (models.UserServices, error) {
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

func (s *Service) DeleteUser(ctx context.Context, userId int64) (err error) {
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

func (s *Service) UpdatePassword(ctx context.Context, userId int64, newPassword string) (err error) {
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

func (s *Service) UsersAll(ctx context.Context) ([]models.UserServices, error) {
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
