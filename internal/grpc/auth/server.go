package authgrpc

import (
	"context"
	"errors"
	"fmt"

	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/internal/services"
)

const ssoAppId = 0

type Auth interface {
	Login(ctx context.Context, appId int32, login string, password string) (token string, err error)
	RegisterNewUser(ctx context.Context, login, password, email, fullName string) (userId int64, err error)

	UpdatePassword(ctx context.Context, id int64, newPassword string) error
	DeleteUser(ctx context.Context, userId int64) error

	SignIn(ctx context.Context, appId int32) (userId int64, err error)
	IsAdmin(ctx context.Context, userId int64) (isAdmin bool, err error)

	UserInfo(ctx context.Context, id int64) (user models.UserServices, err error)
	UsersAll(ctx context.Context) ([]models.UserServices, error)

	UpdateTokenApp(ctx context.Context, appId int32) (newToken string, err error)
}

type serverApi struct {
	ssov1.UnimplementedAuthServer
	auth Auth
	cfg  *config.Config
}

func RegisterServer(gRPC *grpc.Server, auth Auth, cfg *config.Config) {
	ssov1.RegisterAuthServer(gRPC, &serverApi{auth: auth, cfg: cfg})
}

func (s *serverApi) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *serverApi) Login(ctx context.Context, req *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	token, err := s.auth.Login(ctx, req.GetAppId(), req.GetLogin(), req.GetPassword())
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		if errors.Is(err, services.ErrAppNotFound) {
			return nil, status.Error(codes.NotFound, "app not found")
		}
		return nil, status.Error(codes.Internal, fmt.Sprintf("login failed"))
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
		if errors.Is(err, services.ErrUserNotUnique) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.Internal, "register failed")
	}

	return &ssov1.RegisterResponse{UserId: userId}, nil
}

func (s *serverApi) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}
	userId, err := s.SignIn(ctx, ssoAppId) // Id пользователя, который обращается к серверу
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, "x-user-id", userId)

	isAdmin, err := s.auth.IsAdmin(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, "isAdmin failed")
	}

	if !isAdmin && userId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	isAdmin, err = s.auth.IsAdmin(ctx, req.GetUserId())

	if err != nil {
		return nil, status.Error(codes.NotFound, "id not found")
	}

	return &ssov1.IsAdminResponse{IsAdmin: isAdmin}, nil
}

func (s *serverApi) UserInfo(ctx context.Context, req *ssov1.UserInfoRequest) (*ssov1.User, error) {
	userId, err := s.SignIn(ctx, ssoAppId) // Id пользователя, который обращается к серверу
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, "x-user-id", userId)

	isAdmin, err := s.auth.IsAdmin(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, "is admin failed")
	}

	if !isAdmin && userId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	userInfo, err := s.auth.UserInfo(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "get user info failed")
	}

	return &ssov1.User{
		UserId:   userInfo.ID,
		Login:    userInfo.Login,
		Email:    userInfo.Email,
		FullName: userInfo.FullName,
		IsAdmin:  userInfo.IsAdmin,
		CreateAt: timestamppb.New(userInfo.CreatedAt),
		UpdateAt: timestamppb.New(userInfo.UpdatedAt),
	}, nil
}

func (s *serverApi) UpdateToken(ctx context.Context, req *ssov1.UpdateTokenRequest) (*ssov1.UpdateTokenResponse, error) {
	userId, err := s.SignIn(ctx, req.AppId) // Id пользователя, который обращается к серверу
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, "x-user-id", userId)
	token, err := s.auth.UpdateTokenApp(ctx, req.GetAppId())
	if err != nil {
		return nil, status.Error(codes.Internal, "update token failed")
	}
	return &ssov1.UpdateTokenResponse{
		Token: token,
	}, nil

}
func (s *serverApi) UsersInfo(ctx context.Context, _ *emptypb.Empty) (*ssov1.Users, error) {
	userId, err := s.SignIn(ctx, ssoAppId) // Id пользователя, который обращается к серверу
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, "x-user-id", userId)

	isAdmin, err := s.auth.IsAdmin(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, "is admin failed")
	}

	if !isAdmin {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	serverUsers, err := s.auth.UsersAll(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "get users info failed")
	}

	users := make([]*ssov1.User, len(serverUsers))
	for i, userModel := range serverUsers {
		users[i] = &ssov1.User{
			UserId:   userModel.ID,
			Login:    userModel.Login,
			Email:    userModel.Email,
			FullName: userModel.FullName,
			IsAdmin:  userModel.IsAdmin,
			CreateAt: timestamppb.New(userModel.CreatedAt),
			UpdateAt: timestamppb.New(userModel.UpdatedAt),
		}
	}
	return &ssov1.Users{Users: users}, nil
}

func (s *serverApi) UpdatePassword(ctx context.Context, req *ssov1.UpdatePasswordRequest) (*ssov1.UpdatePasswordResponse, error) {
	if err := validateUpdatePassword(req); err != nil {
		return nil, err
	}

	userId, err := s.SignIn(ctx, ssoAppId) // Id пользователя, который обращается к серверу
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, "x-user-id", userId)

	isAdmin, err := s.auth.IsAdmin(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, "is admin failed")
	}

	if !isAdmin && userId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	err = s.auth.UpdatePassword(ctx, req.GetUserId(), req.GetNewPassword())
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "update password failed")
	}

	return &ssov1.UpdatePasswordResponse{
		Message: "password updated"}, nil
}

func (s *serverApi) RemoveUser(ctx context.Context, req *ssov1.RemoveUserRequest) (*ssov1.RemoveUserResponse, error) {
	userId, err := s.SignIn(ctx, ssoAppId) // Id пользователя, который обращается к серверу
	if err != nil {
		return nil, err
	}
	ctx = context.WithValue(ctx, "x-user-id", userId)

	isAdmin, err := s.auth.IsAdmin(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, "is admin failed")
	}

	if !isAdmin && userId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	if err = s.auth.DeleteUser(ctx, req.GetUserId()); err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "delete user failed")
	}
	return &ssov1.RemoveUserResponse{
		Message: "user deleted",
	}, nil
}

func (s *serverApi) SignIn(ctx context.Context, appId int32) (int64, error) {
	userId, err := s.auth.SignIn(ctx, appId)
	if err != nil {
		if errors.Is(err, services.ErrAppNotFound) {
			return 0, status.Error(codes.NotFound, "app not found")
		}
		if errors.Is(err, services.ErrTokenExpired) {
			return 0, status.Error(codes.Unauthenticated, "JWT token expired")
		}
		return 0, status.Error(codes.InvalidArgument, "JWT token invalid")
	}
	return userId, nil
}
