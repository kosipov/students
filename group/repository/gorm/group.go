package gorm

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/models"
)

type GroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (g *GroupRepository) GetGroups(ctx context.Context) (*[]models.Group, error) {
	var groups []models.Group
	result := g.db.Find(&groups)
	return &groups, result.Error
}
