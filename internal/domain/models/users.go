package models

import "time"

type UserServices struct {
	ID           int64
	Login        string
	Email        string
	FullName     string
	PasswordHash []byte
	IsAdmin      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRepository struct {
	ID           int64     `db:"id"`
	Login        string    `db:"login"`
	Email        string    `db:"email"`
	FullName     string    `db:"full_name"`
	PasswordHash []byte    `db:"pass_hash"`
	IsAdmin      bool      `db:"is_admin"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
