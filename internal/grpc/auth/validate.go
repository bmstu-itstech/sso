package authgrpc

import (
	"fmt"
	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	emptyValue    = 0
	maxLenStrings = 255
)

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "login is empty")
	}
	if len(req.GetLogin()) > maxLenStrings {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("login is too long, max length is %d, your login len is %d", maxLenStrings, len(req.GetLogin())))
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is empty")
	}
	if len(req.GetPassword()) > maxLenStrings {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("password is too long, max length is %d, your password len is %d", maxLenStrings, len(req.GetPassword())))
	}
	return nil

}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "login is empty")
	}
	if len(req.GetLogin()) > maxLenStrings {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("login is too long, max length is %d, your login len is %d", maxLenStrings, len(req.GetLogin())))
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is empty")
	}
	if len(req.GetPassword()) > maxLenStrings {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("password is too long, max length is %d, your password len is %d", maxLenStrings, len(req.GetPassword())))
	}
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is empty")
	}
	if len(req.GetEmail()) > maxLenStrings {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("email is too long, max length is %d, your email len is %d", maxLenStrings, len(req.GetEmail())))
	}
	if req.GetFullName() == "" {
		return status.Error(codes.InvalidArgument, "full name is empty")
	}
	if len(req.GetFullName()) > maxLenStrings {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("full name is too long, max length is %d, your full name len is %d", maxLenStrings, len(req.GetFullName())))
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
	if req.GetNewPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is empty")
	}
	if len(req.GetNewPassword()) > maxLenStrings {
		return status.Error(codes.InvalidArgument, fmt.Sprintf("password is too long, max length is %d, your password len is %d", maxLenStrings, len(req.GetNewPassword())))
	}
	return nil
}
