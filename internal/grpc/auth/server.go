package authgrpc

import (
	"context"
	ssov1 "github.com/BOBAvov/protos_sso/gen/go/sso"
	"github.com/bmstu-itstech/sso/internal/config"
	"github.com/bmstu-itstech/sso/internal/domain/models"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AuthGrpc interface {
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
	auth AuthGrpc
	cfg  *config.Config
}

func RegisterServer(gRPC *grpc.Server, auth AuthGrpc, cfg *config.Config) {
	ssov1.RegisterAuthServer(gRPC, &serverApi{auth: auth, cfg: cfg})
}

func (s *serverApi) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (s *serverApi) VerifyToken(ctx context.Context, req *ssov1.VerifyTokenRequest) (*ssov1.VerifyTokenResponse, error) {
	return nil, nil
}
