package models

import "time"

type User struct {
	ID           int64
	Login        string
	Email        string
	FullName     string
	PasswordHash []byte
	IsAdmin      bool
	CreatedAt    time.Duration
	UpdatedAt    time.Duration
}
