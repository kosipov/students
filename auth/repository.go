package auth

import (
	"context"
	"github.com/kosipov/students/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmailAndPass(ctx context.Context, username, password string) (*models.User, error)
}
