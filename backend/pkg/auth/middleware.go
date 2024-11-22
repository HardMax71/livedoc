package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthMiddleware struct {
	service *Service
}

func NewAuthMiddleware(service *Service) *AuthMiddleware {
	return &AuthMiddleware{service: service}
}

type contextKey string

const (
	UserContextKey contextKey = "user"
)

func (m *AuthMiddleware) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for login and register
		if info.FullMethod == "/auth.v1.AuthService/Login" ||
			info.FullMethod == "/auth.v1.AuthService/Register" ||
			info.FullMethod == "/auth.v1.AuthService/Refresh" {
			return handler(ctx, req)
		}

		user, err := m.authorize(ctx)
		if err != nil {
			return nil, err
		}

		newCtx := context.WithValue(ctx, UserContextKey, user)
		return handler(newCtx, req)
	}
}

func (m *AuthMiddleware) StreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		user, err := m.authorize(ss.Context())
		if err != nil {
			return err
		}

		newCtx := context.WithValue(ss.Context(), UserContextKey, user)
		wrappedStream := newWrappedServerStream(ss, newCtx)
		return handler(srv, wrappedStream)
	}
}

func (m *AuthMiddleware) authorize(ctx context.Context) (*User, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing authorization token")
	}

	user, err := m.service.VerifyToken(ctx, values[0])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	return user, nil
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func newWrappedServerStream(ss grpc.ServerStream, ctx context.Context) *wrappedServerStream {
	return &wrappedServerStream{
		ServerStream: ss,
		ctx:          ctx,
	}
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// GetUserFromContext retrieves the user from the context
func GetUserFromContext(ctx context.Context) (*User, error) {
	user, ok := ctx.Value(UserContextKey).(*User)
	if !ok {
		return nil, status.Error(codes.Internal, "could not retrieve user from context")
	}
	return user, nil
}
