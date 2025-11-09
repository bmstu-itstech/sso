package authgrpc

import (
	"context"
	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth struct {
	Login           func(ctx context.Context, appId int32, login string, password string) (token string, err error)
	RegisterNewUser func(ctx context.Context, login, password, email, fullName string) (userId int64, err error)
	IsAdmin         func(ctx context.Context, userId int64) (isAdmin bool, err error)
	//UserInfo        func(ctx context.Context, id int64) (user *ssov1.User, err error)
	//UpdatePassword  func(ctx context.Context, id int64, newPassword string) (*ssov1.User, error)
	//DeleteUser      func(ctx context.Context, req *ssov1.RemoveUserRequest) error
}
type serverApi struct {
	ssov1.UnimplementedAuthServer
	auth *Auth
}

func RegisterServer(gRPC *grpc.Server, auth *Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverApi{auth: auth})
}

func (s *serverApi) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetAppId(), req.GetLogin(), req.GetPassword())
	if err != nil {
		// TODO: обработка ошибок
		return nil, status.Error(codes.Internal, "login or password is incorrect")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverApi) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}
	userId, err := s.auth.RegisterNewUser(ctx, req.GetLogin(), req.GetPassword(), req.GetEmail(), req.GetFullName())
	if err != nil {
		// TODO: ...
		return nil, status.Error(codes.Internal, "register failed")
	}
	return &ssov1.RegisterResponse{UserId: userId}, nil
}

func (s *serverApi) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}
	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		// TODO: ...
		return nil, status.Error(codes.Internal, "is admin failed")
	}
	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *serverApi) UserInfo(ctx context.Context, req *ssov1.UserInfoRequest) (*ssov1.User, error) {
	panic("implement me")
}

func (s *serverApi) UpdatePassword(ctx context.Context, req *ssov1.UpdatePasswordRequest) (*ssov1.User, error) {
	panic("implement me")
}

func (s *serverApi) RemoveUser(ctx context.Context, req *ssov1.RemoveUserRequest) (*ssov1.RemoveUserResponse, error) {
	panic("implement me")
}
