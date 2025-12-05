package storage

import "errors"

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrAppNotFound   = errors.New("app not found")
	ErrUserNotUnique = errors.New("user not unique")
)
