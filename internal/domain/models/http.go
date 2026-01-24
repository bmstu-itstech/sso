package models

import (
	"time"
)

type LoginRequest struct {
	AppId    int32  `json:"appId" validate:"required"`
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"fullName" validate:"required"`
}

type RegisterResponse struct {
	UserId int64 `json:"userId"`
}

type IsAdminRequest struct {
	UserId int64 `json:"userId" validate:"required"`
}

type IsAdminResponse struct {
	IsAdmin bool `json:"isAdmin"`
}

type UserInfoRequest struct {
	UserId int64 `json:"userId" validate:"required"`
}

type User struct {
	UserId    int64     `json:"userId"`
	Login     string    `json:"login"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	IsAdmin   bool      `json:"isAdmin"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
}

type Users struct {
	Users []User `json:"users"`
}

type UpdateTokenRequest struct {
	AppId int32 `json:"appId" validate:"required"`
}

type UpdateTokenResponse struct {
	Token string `json:"token"`
}

type UpdatePasswordRequest struct {
	UserId      int64  `json:"userId" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required,min=8"`
}

type UpdatePasswordResponse struct {
	Message string `json:"message"`
}

type RemoveUserRequest struct {
	UserId int64 `json:"userId" validate:"required"`
}

type RemoveUserResponse struct {
	Message string `json:"message"`
}
