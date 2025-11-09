package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	usersTable = "users"
	appsTable  = "apps"
)

type Repository struct {
	db *sqlx.DB
}

func NewPostgresDB(cfg config.PostgresConfig) (Repository, error) {
	const op = "repository.NewPostgresDB"
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.Host, cfg.Port, cfg.UserName, cfg.Password, cfg.DB))
	if err != nil {
		return Repository{}, fmt.Errorf("%w: %s", err, op)
	}

	err = db.Ping()
	if err != nil {
		return Repository{}, fmt.Errorf("%w: %s", err, op)
	}
	return Repository{db: db}, nil
}

func (r *Repository) SaveUser(ctx context.Context, login string, password []byte, email, fullName string, userId int64) (err error) {
	const op = "repository.SaveUser"
	query := fmt.Sprintf("INSERT INTO users (id,login, email,full_name, pass_hash) VALUES($1,$2,$3,$4,$5)")
	_, err = r.db.ExecContext(ctx, query, userId, login, email, fullName, password)
	if err != nil {
		return fmt.Errorf("%w: %s", err, op)
	}
	return nil
}

func (r *Repository) User(ctx context.Context, login string) (user models.User, err error) {
	const op = "repository.User"
	const query = "SELECT * FROM users WHERE login = $1"

	if err = r.db.GetContext(ctx, &user, query, login); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errors.New("user not found")
		}
		fmt.Println(err)
		return models.User{}, fmt.Errorf("%w: %s", err, op)
	}
	return user, nil
}

func (r *Repository) UserIsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error) {
	const op = "repository.UserIsAdmin"
	query := fmt.Sprintf("SELECT is_admin FROM users WHERE id = $1")

	if err = r.db.GetContext(ctx, &isAdmin, query, userId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("user not found")
		}
		return false, fmt.Errorf("%w: %s", err, op)
	}

	return isAdmin, nil
}

func (r *Repository) App(ctx context.Context, appId int32) (app models.App, err error) {
	const op = "repository.App"
	query := fmt.Sprintf("SELECT * FROM apps WHERE id = $1")
	if err = r.db.GetContext(ctx, &app, query, appId); err != nil {
		return models.App{}, fmt.Errorf("%w: %s", err, op)
	}
	return app, nil
}
