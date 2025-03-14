package auth

import (
	"context"
	"errors"
	ssov1 "github.com/dmitry-muffin/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/mail"
	"sso/internal/services/auth"
	"sso/internal/storage"
)

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
		appID int32,
	) (token string, err error)

	RegisterNewUser(
		ctx context.Context,
		name string,
		email string,
		password string,
	) (userID int64, err error)

	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth *auth.Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Login(ctx context.Context, in *ssov1.LoginRequest) (*ssov1.LoginResponse, error) {

	// auth service
	if err := validateLogin(in); err != nil {
		return nil, err
	}
	token, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword(), in.GetAppId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &ssov1.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, in *ssov1.RegisterRequest) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(in); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, in.GetName(), in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {

			return nil, status.Error(codes.AlreadyExists, "user already exists")

		}

		return nil, status.Error(codes.Internal, "internal server error")
	}
	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, in *ssov1.IsAdminRequest) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(in); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, in.GetUserId())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func validateLogin(request *ssov1.LoginRequest) error {
	_, err := mail.ParseAddress(request.GetEmail())
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid email")
	}
	if request.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "invalid password")
	}
	//if request.GetAppId() == emptyValue {
	//	return status.Error(codes.InvalidArgument, "invalid app id")
	//}
	return nil
}

func validateRegister(request *ssov1.RegisterRequest) error {
	_, err := mail.ParseAddress(request.GetEmail())
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid email")
	}
	if request.GetName() == "" {
		return status.Error(codes.InvalidArgument, "invalid name")
	}
	if request.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "invalid password")
	}

	return nil
}

func validateIsAdmin(request *ssov1.IsAdminRequest) error {
	if request.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "invalid user id")
	}
	return nil
}
