package models

import "time"

type VerifyTokenServiceRequest struct {
	AppId int32
	Token string
}
type VerifyTokenServiceResponse struct {
	TokenInfo
}
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
