package auth

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserParams struct {
	Email    string
	Username string
	Password string
}

func (p *CreateUserParams) HashPassword() error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	p.Password = string(hashedBytes)
	return nil
}

type LoginParams struct {
	Email    string
	Password string
}

func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
}

type RefreshToken struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
