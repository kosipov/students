package auth

import (
	"context"
	"github.com/kosipov/students/models"
)

const CtxUserKey = "user"

type UseCase interface {
	SignUp(ctx context.Context, username, password string) error
	SignIn(ctx context.Context, username, password string) (*models.User, error)
}
