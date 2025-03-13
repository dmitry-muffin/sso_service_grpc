package auth

import (
	"context"
	ssov1 "github.com/dmitry-muffin/protos/gen/go/sso"
	"google.golang.org/grpc"
)

type serverAPI struct {
	ssov1.UnimplementedAuthServer
}

func Register(gRPC *grpc.Server) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{})
}

func (s *serverAPI) Login(ctx context.Context, in *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {
	return &ssov1.LoginResponse{
		Token: in.GetEmail(),
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, in *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	panic("implement me")
}

func (s *serverAPI) IsAdmin(ctx context.Context, in *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	panic("implement me")
}
