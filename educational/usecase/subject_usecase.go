package usecase

import (
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
)

type SubjectUseCase struct {
	subjectRepo educational.CommonSubjectRepository
}

func NewSubjectUseCase(subjectRepo educational.CommonSubjectRepository) *SubjectUseCase {
	return &SubjectUseCase{subjectRepo: subjectRepo}
}

func (subjectUseCase *SubjectUseCase) GetSubjectsByGroup(ctx context.Context, groupId int) (*[]models.Subject, error) {
	group, err := subjectUseCase.subjectRepo.GetGroup(ctx, groupId)
	if err != nil {
		return nil, err
	}
	return subjectUseCase.subjectRepo.GetSubjectsByGroup(ctx, group)
}

func (subjectUseCase *SubjectUseCase) SubjectObjectListFromSubject(ctx context.Context, subjectId int) (*[]models.SubjectObject, error) {
	subject, err := subjectUseCase.GetSubjectById(ctx, subjectId)
	if err != nil {
		return nil, err
	}
	return subjectUseCase.subjectRepo.GetSubjectObjectsBySubject(ctx, subject)
}

func (subjectUseCase *SubjectUseCase) GetSubjectById(ctx context.Context, id int) (*models.Subject, error) {
	subject, err := subjectUseCase.subjectRepo.GetSubject(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, educational.ErrSubjectNotFound
		}
		return nil, err
	}
	return subject, nil
}

func (subjectUseCase *SubjectUseCase) GetAllSubject(ctx context.Context) (*[]models.Subject, error) {
	return subjectUseCase.subjectRepo.GetSubjects(ctx)
}

func (subjectUseCase *SubjectUseCase) CreateSubject(ctx context.Context, name string, groupId int) error {
	subject := &models.Subject{SubjectName: name, GroupId: groupId}

	return subjectUseCase.subjectRepo.CreateSubject(ctx, subject)
}

func (subjectUseCase *SubjectUseCase) CreateSubjectObject(ctx context.Context, name string, subjectId int, href string) (*models.SubjectObject, error) {
	subjectObject := &models.SubjectObject{SubjectId: subjectId, Name: name, Href: href}

	return subjectObject, subjectUseCase.subjectRepo.CreateSubjectObject(ctx, subjectObject)
}

func (subjectUseCase *SubjectUseCase) GetSubjectObject(ctx context.Context, subjectId int, subjectObjectId int) (*models.SubjectObject, error) {
	subjectObject, err := subjectUseCase.subjectRepo.GetSubjectObject(ctx, subjectObjectId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, educational.ErrSubjectObjectNotFound
		}
		return nil, err
	}
	if subjectObject.SubjectId != subjectId {
		return nil, educational.ErrSubjectObjectNotFound
	}
	return subjectObject, nil
}

func (subjectUseCase *SubjectUseCase) UpdateSubjectObject(ctx context.Context, subjectId int, subjectObjectId int, name string, href string) (*models.SubjectObject, error) {
	subjectObject, err := subjectUseCase.GetSubjectObject(ctx, subjectId, subjectObjectId)
	if err != nil {
		return nil, err
	}

	subjectObject.Name = name
	subjectObject.Href = href

	return subjectObject, subjectUseCase.subjectRepo.UpdateSubjectObject(ctx, subjectObject)
}

func (subjectUseCase *SubjectUseCase) DeleteSubjectObject(ctx context.Context, subjectId int, subjectObjectId int) error {
	subjectObject, err := subjectUseCase.GetSubjectObject(ctx, subjectId, subjectObjectId)
	if err != nil {
		return err
	}
	return subjectUseCase.subjectRepo.DeleteSubjectObject(ctx, subjectObject)
}
