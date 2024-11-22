package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/HardMax71/syncwrite/backend/pkg/config"
	"github.com/HardMax71/syncwrite/backend/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

type Service struct {
	db     *pgxpool.Pool
	config *config.Config
	logger *zap.Logger
}

func NewService(db *pgxpool.Pool, config *config.Config) *Service {
	return &Service{
		db:     db,
		config: config,
		logger: utils.Logger(),
	}
}

func (s *Service) Register(ctx context.Context, params CreateUserParams) (*User, string, string, error) {
	// Check if user exists
	var exists bool
	err := s.db.QueryRow(ctx, `
        SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
    `, params.Email).Scan(&exists)

	if err != nil {
		return nil, "", "", fmt.Errorf("error checking user existence: %w", err)
	}

	if exists {
		return nil, "", "", ErrUserExists
	}

	// Hash password
	if err := params.HashPassword(); err != nil {
		return nil, "", "", fmt.Errorf("error hashing password: %w", err)
	}

	// Create user
	var user User
	err = s.db.QueryRow(ctx, `
        INSERT INTO users (email, username, password_hash)
        VALUES ($1, $2, $3)
        RETURNING id, email, username, password_hash, created_at, updated_at
    `, params.Email, params.Username, params.Password).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, "", "", fmt.Errorf("error creating user: %w", err)
	}

	// Generate tokens
	accessToken, err := utils.GenerateToken(user.ID, user.Username, s.config.JWT.Secret, s.config.JWT.ExpiryDuration)
	if err != nil {
		return nil, "", "", fmt.Errorf("error generating access token: %w", err)
	}

	refreshToken, err := utils.GenerateToken(user.ID, user.Username, s.config.JWT.Secret, s.config.JWT.RefreshDuration)
	if err != nil {
		return nil, "", "", fmt.Errorf("error generating refresh token: %w", err)
	}

	// Store refresh token
	_, err = s.db.Exec(ctx, `
        INSERT INTO refresh_tokens (user_id, token, expires_at)
        VALUES ($1, $2, $3)
    `, user.ID, refreshToken, time.Now().Add(s.config.JWT.RefreshDuration))

	if err != nil {
		return nil, "", "", fmt.Errorf("error storing refresh token: %w", err)
	}

	return &user, accessToken, refreshToken, nil
}

func (s *Service) Login(ctx context.Context, params LoginParams) (*User, string, string, error) {
	var user User
	err := s.db.QueryRow(ctx, `
        SELECT id, email, username, password_hash, created_at, updated_at
        FROM users WHERE email = $1
    `, params.Email).Scan(
		&user.ID, &user.Email, &user.Username, &user.PasswordHash,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	if err := user.ComparePassword(params.Password); err != nil {
		return nil, "", "", ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := utils.GenerateToken(user.ID, user.Username, s.config.JWT.Secret, s.config.JWT.ExpiryDuration)
	if err != nil {
		return nil, "", "", fmt.Errorf("error generating access token: %w", err)
	}

	refreshToken, err := utils.GenerateToken(user.ID, user.Username, s.config.JWT.Secret, s.config.JWT.RefreshDuration)
	if err != nil {
		return nil, "", "", fmt.Errorf("error generating refresh token: %w", err)
	}

	// Store refresh token
	_, err = s.db.Exec(ctx, `
        INSERT INTO refresh_tokens (user_id, token, expires_at)
        VALUES ($1, $2, $3)
    `, user.ID, refreshToken, time.Now().Add(s.config.JWT.RefreshDuration))

	if err != nil {
		return nil, "", "", fmt.Errorf("error storing refresh token: %w", err)
	}

	return &user, accessToken, refreshToken, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := utils.ValidateToken(refreshToken, s.config.JWT.Secret)
	if err != nil {
		return "", "", ErrInvalidToken
	}

	// Check if refresh token exists and is valid
	var tokenExists bool
	err = s.db.QueryRow(ctx, `
        SELECT EXISTS(
            SELECT 1 FROM refresh_tokens 
            WHERE token = $1 AND user_id = $2 AND expires_at > NOW()
        )
    `, refreshToken, claims.UserID).Scan(&tokenExists)

	if err != nil || !tokenExists {
		return "", "", ErrInvalidToken
	}

	// Generate new tokens
	newAccessToken, err := utils.GenerateToken(claims.UserID, claims.Username, s.config.JWT.Secret, s.config.JWT.ExpiryDuration)
	if err != nil {
		return "", "", fmt.Errorf("error generating access token: %w", err)
	}

	newRefreshToken, err := utils.GenerateToken(claims.UserID, claims.Username, s.config.JWT.Secret, s.config.JWT.RefreshDuration)
	if err != nil {
		return "", "", fmt.Errorf("error generating refresh token: %w", err)
	}

	// Delete old refresh token and store new one
	err = s.db.QueryRow(ctx, `
        WITH deleted AS (
            DELETE FROM refresh_tokens WHERE token = $1 RETURNING user_id
        )
        INSERT INTO refresh_tokens (user_id, token, expires_at)
        SELECT user_id, $2, $3 FROM deleted
        RETURNING id
    `, refreshToken, newRefreshToken, time.Now().Add(s.config.JWT.RefreshDuration)).Scan(new(string))

	if err != nil {
		return "", "", fmt.Errorf("error updating refresh token: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	_, err := s.db.Exec(ctx, `
        DELETE FROM refresh_tokens WHERE token = $1
    `, refreshToken)

	if err != nil {
		return fmt.Errorf("error deleting refresh token: %w", err)
	}

	return nil
}

func (s *Service) VerifyToken(ctx context.Context, token string) (*User, error) {
	claims, err := utils.ValidateToken(token, s.config.JWT.Secret)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var user User
	err = s.db.QueryRow(ctx, `
        SELECT id, email, username, created_at, updated_at
        FROM users WHERE id = $1
    `, claims.UserID).Scan(
		&user.ID, &user.Email, &user.Username,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, ErrUserNotFound
	}

	return &user, nil
}
