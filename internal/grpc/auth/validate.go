package authgrpc

import (
	"fmt"
	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	emptyValue = 0
)

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "login is empty")
	}
	if len(req.GetLogin()) > 255 {
		return status.Error(codes.InvalidArgument, "login is too long")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is empty")
	}
	if len(req.GetPassword()) > 255 {
		return status.Error(codes.InvalidArgument, "password is too long")
	}
	return nil

}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "login is empty")
	}
	if len(req.GetLogin()) > 255 {
		return status.Error(codes.InvalidArgument, "login is too long")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is empty")
	}
	if len(req.GetPassword()) > 255 {
		return status.Error(codes.InvalidArgument, "password is too long")
	}
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is empty")
	}
	if len(req.GetEmail()) > 255 {
		return status.Error(codes.InvalidArgument, "email is too long")
	}
	if req.GetFullName() == "" {
		return status.Error(codes.InvalidArgument, "full name is empty")
	}
	if len(req.GetFullName()) > 255 {
		return status.Error(codes.InvalidArgument, "full name is too long")
	}
	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "user id is empty")
	}
	return nil
}

func validateUpdatePassword(req *ssov1.UpdatePasswordRequest) error {
	fmt.Println(req.GetNewPassword(), req.NewPassword)
	if req.GetNewPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is empty")
	}
	if len(req.GetNewPassword()) > 255 {
		return status.Error(codes.InvalidArgument, "password is too long")
	}
	return nil
}
