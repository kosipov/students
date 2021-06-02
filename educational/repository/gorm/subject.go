package gorm

import (
	"context"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/models"
)

type SubjectRepository struct {
	db *gorm.DB
}

func NewSubjectRepository(db *gorm.DB) *SubjectRepository {
	return &SubjectRepository{db: db}
}

func (subjectRepo *SubjectRepository) GetSubjectsByGroup(ctx context.Context, group *models.Group) (*[]models.Subject, error) {
	var subjects []models.Subject
	result := subjectRepo.db.Model(&group).Association("Subjects").Find(&subjects)
	return &subjects, result.Error
}

func (subjectRepo *SubjectRepository) GetSubjectObjectsBySubject(ctx context.Context, subject *models.Subject) (*[]models.SubjectObject, error) {
	var subjectObjects []models.SubjectObject
	result := subjectRepo.db.Model(&subject).Association("SubjectObjects").Find(&subjectObjects)
	return &subjectObjects, result.Error
}

func (subjectRepo *SubjectRepository) GetSubject(ctx context.Context, id int) (*models.Subject, error) {
	var subject models.Subject
	result := subjectRepo.db.First(&subject, id)
	return &subject, result.Error
}

func (subjectRepo *SubjectRepository) GetSubjects(ctx context.Context) (*[]models.Subject, error) {
	var subjects []models.Subject
	result := subjectRepo.db.Preload("Group").Find(&subjects)
	return &subjects, result.Error
}
