package usecase

import (
	"context"
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
	subject, err := subjectUseCase.subjectRepo.GetSubject(ctx, subjectId)
	if err != nil {
		return nil, err
	}
	return subjectUseCase.subjectRepo.GetSubjectObjectsBySubject(ctx, subject)
}

func (subjectUseCase *SubjectUseCase) GetSubjectById(ctx context.Context, id int) (*models.Subject, error) {
	return subjectUseCase.subjectRepo.GetSubject(ctx, id)
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

func (subjectUseCase *SubjectUseCase) DeleteSubjectObject(ctx context.Context, subjectObjectId int) error {
	subjectObject, err := subjectUseCase.subjectRepo.GetSubjectObject(ctx, subjectObjectId)
	if err != nil {
		return err
	}
	err = subjectUseCase.subjectRepo.DeleteSubjectObject(ctx, subjectObject)
	return err
}
