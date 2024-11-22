package auth

import (
	"context"

	"github.com/HardMax71/syncwrite/backend/pkg/proto/auth/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	params := CreateUserParams{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	user, accessToken, refreshToken, err := h.service.Register(ctx, params)
	if err != nil {
		switch err {
		case ErrUserExists:
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &authv1.RegisterResponse{
		User: &authv1.User{
			Id:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	params := LoginParams{
		Email:    req.Email,
		Password: req.Password,
	}

	user, accessToken, refreshToken, err := h.service.Login(ctx, params)
	if err != nil {
		switch err {
		case ErrInvalidCredentials:
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &authv1.LoginResponse{
		User: &authv1.User{
			Id:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	accessToken, refreshToken, err := h.service.Refresh(ctx, req.RefreshToken)
	if err != nil {
		switch err {
		case ErrInvalidToken:
			return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &authv1.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *Handler) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	err := h.service.Logout(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.LogoutResponse{Success: true}, nil
}

func (h *Handler) VerifyToken(ctx context.Context, req *authv1.VerifyTokenRequest) (*authv1.VerifyTokenResponse, error) {
	user, err := h.service.VerifyToken(ctx, req.Token)
	if err != nil {
		switch err {
		case ErrInvalidToken:
			return &authv1.VerifyTokenResponse{Valid: false}, nil
		case ErrUserNotFound:
			return &authv1.VerifyTokenResponse{Valid: false}, nil
		default:
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	return &authv1.VerifyTokenResponse{
		Valid: true,
		User: &authv1.User{
			Id:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		},
	}, nil
}
