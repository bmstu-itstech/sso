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
	VerifyToken(ctx context.Context, req *models.VerifyTokenServiceRequest) (models.VerifyTokenServiceResponse, error)
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
	tokenModel, err := s.auth.VerifyToken(ctx, &models.VerifyTokenServiceRequest{Token: req.Token, AppId: req.AppId})
	if err != nil {
		return nil, err
	}
	return &ssov1.VerifyTokenResponse{
		IsAdmin: tokenModel.IsAdmin,
		UserId:  tokenModel.Uid,
	}, nil
}
