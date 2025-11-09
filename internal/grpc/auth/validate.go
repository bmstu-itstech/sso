package authgrpc

import (
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

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is empty")
	}
	return nil

}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetLogin() == "" {
		return status.Error(codes.InvalidArgument, "login is empty")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is empty")
	}
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is empty")
	}
	if req.GetFullName() == "" {
		return status.Error(codes.InvalidArgument, "full name is empty")
	}
	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "user id is empty")
	}
	return nil
}
