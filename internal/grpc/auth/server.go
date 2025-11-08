package authgrpc

import (
	"context"
	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"google.golang.org/grpc"
)

type serverApi struct {
	ssov1.UnimplementedAuthServer
}

func Register(gRPC *grpc.Server) {
	ssov1.RegisterAuthServer(gRPC, &serverApi{})
}

func (s *serverApi) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	return &ssov1.LoginResponse{
		Token: "token",
	}, nil
}
func (s *serverApi) Register(ctx context.Context, req *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	panic("implement me")
}
func (s *serverApi) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	panic("implement me")
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
