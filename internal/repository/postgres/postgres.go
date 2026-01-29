package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"

	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/internal/domain/storage"
)

type Repository struct {
	db *sqlx.DB
}

func NewPostgresDB(cfg config.Config) (Repository, error) {
	const op = "repository.NewPostgresDB"

	connectionString := cfg.GetPostgresPath()
	db, err := sqlx.Open("postgres", connectionString)
	if err != nil {
		return Repository{}, fmt.Errorf("%w: %s", err, op)
	}
	db.SetMaxOpenConns(500)                 // Максимум открытых соединений
	db.SetMaxIdleConns(20)                  // Максимум простаивающих соединений
	db.SetConnMaxLifetime(5 * time.Minute)  // Время жизни соединения
	db.SetConnMaxIdleTime(10 * time.Minute) // Время простоя соединения

	err = db.Ping()
	if err != nil {
		return Repository{}, fmt.Errorf("%s: %w", op, err)
	}
	return Repository{db: db}, nil
}

func (r *Repository) SaveUser(ctx context.Context, login string, password []byte, email, fullName string, userId int64) (err error) {
	const op = "repository.SaveUser"
	const query = "INSERT INTO users (id, login, email, full_name, pass_hash) VALUES($1,$2,$3,$4,$5)"
	_, err = r.db.ExecContext(ctx, query, userId, login, email, fullName, password)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return storage.ErrUserNotUnique
		}
		return fmt.Errorf("%w: %s", err, op)
	}
	return nil
}

func (r *Repository) UserByLogin(ctx context.Context, login string) (user models.UserRepository, err error) {
	const op = "repository.UserByLogin"
	const query = "SELECT * FROM users WHERE login = $1"

	if err = r.db.GetContext(ctx, &user, query, login); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.UserRepository{}, storage.ErrUserNotFound
		}
		return models.UserRepository{}, fmt.Errorf("%w: %s", err, op)
	}
	return user, nil
}

func (r *Repository) UserIsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error) {
	const op = "repository.UserIsAdmin"
	const query = "SELECT is_admin FROM users WHERE id = $1"

	if err = r.db.GetContext(ctx, &isAdmin, query, userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, storage.ErrUserNotFound
		}
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

func (r *Repository) App(ctx context.Context, appId int32) (app models.AppRepos, err error) {
	const op = "repository.App"
	const query = "SELECT * FROM apps WHERE id = $1"
	if err = r.db.GetContext(ctx, &app, query, appId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.AppRepos{}, storage.ErrAppNotFound
		}
		return models.AppRepos{}, fmt.Errorf("%w: %s", err, op)
	}
	return app, nil
}

func (r *Repository) UserById(ctx context.Context, userId int64) (user models.UserRepository, err error) {
	const op = "repository.UserById"

	const query = "SELECT * FROM users WHERE id = $1"

	if err = r.db.GetContext(ctx, &user, query, userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.UserRepository{}, storage.ErrUserNotFound
		}
		return models.UserRepository{}, fmt.Errorf("%w: %s", err, op)
	}

	return user, nil
}

func (r *Repository) UserDelete(ctx context.Context, userId int64) (err error) {
	const op = "repository.UserDelete"
	const query = "DELETE FROM users WHERE id = $1"
	_, err = r.db.ExecContext(ctx, query, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrUserNotFound
		}
		return fmt.Errorf("%w: %s", err, op)
	}
	return nil
}

func (r *Repository) UserNewPassword(ctx context.Context, userId int64, newPassword []byte) (err error) {
	const op = "repository.UserNewPassword"
	const query = "UPDATE users SET pass_hash = $1, updated_at = NOW() WHERE id = $2"
	_, err = r.db.ExecContext(ctx, query, newPassword, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrUserNotFound
		}
		return fmt.Errorf("%w: %s", err, op)
	}
	return nil
}

func (r *Repository) UsersAll(ctx context.Context) (users []models.UserRepository, err error) {
	const op = "repository.UsersAll"
	const query = "SELECT * FROM users"
	if err = r.db.SelectContext(ctx, &users, query); err != nil {
		return nil, fmt.Errorf("%w: %s", err, op)
	}
	return users, nil
}
