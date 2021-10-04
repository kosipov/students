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
	result := subjectRepo.db.Limit(1).Find(&subject, id)
	return &subject, result.Error
}

func (subjectRepo *SubjectRepository) GetSubjects(ctx context.Context) (*[]models.Subject, error) {
	var subjects []models.Subject
	result := subjectRepo.db.Preload("Group").Find(&subjects)
	return &subjects, result.Error
}

func (subjectRepo *SubjectRepository) CreateSubject(ctx context.Context, subject *models.Subject) error {
	result := subjectRepo.db.Create(subject)
	return result.Error
}

func (subjectRepo *SubjectRepository) CreateSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error {
	result := subjectRepo.db.Create(subjectObject)
	return result.Error
}

func (subjectRepo *SubjectRepository) GetSubjectObject(ctx context.Context, subjectObjectId int) (*models.SubjectObject, error) {
	var subjectObject models.SubjectObject
	result := subjectRepo.db.Limit(1).Find(&subjectObject, subjectObjectId)
	return &subjectObject, result.Error
}

func (subjectRepo *SubjectRepository) DeleteSubjectObject(ctx context.Context, subjectObject *models.SubjectObject) error {
	return subjectRepo.db.Delete(&subjectObject).Error
}

func (subjectRepo *SubjectRepository) GetGroup(ctx context.Context, groupId int) (*models.Group, error) {
	var group models.Group
	result := subjectRepo.db.Limit(1).Find(&group, groupId)

	return &group, result.Error
}
