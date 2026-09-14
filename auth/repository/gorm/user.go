package gorm

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/models"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	return u.db.Create(user).Error
}

func (u *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := u.db.Where(&models.User{Username: username}).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u *UserRepository) GetUserByEmailAndPass(ctx context.Context, username, password string) (*models.User, error) {
	var user = models.User{}

	if err := u.db.Where(&models.User{Username: username, Password: password}).First(&user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return &user, nil
}
