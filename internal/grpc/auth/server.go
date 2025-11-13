package authgrpc

import (
	"context"
	"errors"
	"fmt"
	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"github.com/bmstu-itstech/sso/internal/lib/jwt"
	"github.com/bmstu-itstech/sso/internal/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
)

type Auth struct {
	Login           func(ctx context.Context, appId int32, login string, password string) (token string, err error)
	RegisterNewUser func(ctx context.Context, login, password, email, fullName string) (userId int64, err error)
	IsAdmin         func(ctx context.Context, userId int64) (isAdmin bool, err error)
	UserInfo        func(ctx context.Context, id int64) (user models.User, err error)
	UpdatePassword  func(ctx context.Context, id int64, newPassword string) error
	DeleteUser      func(ctx context.Context, userId int64) error
	UsersAll        func(ctx context.Context) ([]models.User, error)
}

type serverApi struct {
	ssov1.UnimplementedAuthServer
	auth *Auth
	cfg  *config.Config
}

func RegisterServer(gRPC *grpc.Server, auth *Auth, cfg *config.Config) {
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
		if errors.Is(err, services.ErrAppNotFoud) {
			return nil, status.Error(codes.NotFound, "app not found")
		}
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
		return nil, status.Error(codes.Internal, "register failed")
	}

	return &ssov1.RegisterResponse{UserId: userId}, nil
}

func (s *serverApi) IsAdmin(ctx context.Context, req *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}

	userId, err := s.getUserId(ctx) // Id пользователя, который обращается к серверу
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Yoy have not jwt token")
	}
	ctx = context.WithValue(ctx, "uid", userId)

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
	userId, err := s.getUserId(ctx) // Id пользователя, который обращается к серверу
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Yoy have not jwt token")
	}
	ctx = context.WithValue(ctx, "uid", userId)

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

func (s *serverApi) UsersInfo(ctx context.Context, _ *emptypb.Empty) (*ssov1.Users, error) {
	userId, err := s.getUserId(ctx) // Id пользователя, который обращается к серверу
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Yoy have not jwt token")
	}
	ctx = context.WithValue(ctx, "uid", userId)

	isAdmin, err := s.auth.IsAdmin(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, "is admin failed")
	}

	if !isAdmin {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	// TODO: провести бенчмарк для проверки на сколько выгодно и безопасно передовать один указатель на слайс
	// TODO: или создавать новый слайс и копировать туда значения
	// TODO: Сейчас реализован второй вариант
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
	userId, err := s.getUserId(ctx) // Id пользователя, который обращается к серверу
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Yoy have not jwt token")
	}
	ctx = context.WithValue(ctx, "uid", userId)

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
	userId, err := s.getUserId(ctx) // Id пользователя, который обращается к серверу
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Yoy have not jwt token")
	}
	ctx = context.WithValue(ctx, "uid", userId)

	isAdmin, err := s.auth.IsAdmin(ctx, userId)
	if err != nil {
		return nil, status.Error(codes.Internal, "is admin failed")
	}

	if !isAdmin && userId != req.GetUserId() {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	fmt.Println(ctx, userId)

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

func (s *serverApi) getUserId(ctx context.Context) (int64, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return 0, status.Error(codes.Unauthenticated, "authorization is required")
	}

	jwtString := authHeaders[0]
	jwtString = strings.TrimPrefix(jwtString, "Bearer ")
	userId, err := jwt.ParseSSOJwtToken(jwtString, s.getTokenJwtSSO())

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, status.Error(codes.Unauthenticated, "token is expired")
		}
		return 0, status.Error(codes.Unauthenticated, "token is invalid")
	}
	return userId, nil

}

func (s *serverApi) getTokenJwtSSO() string {
	return s.cfg.JWT.Secret
}
