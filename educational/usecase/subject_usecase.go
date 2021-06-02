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

func (subjectUseCase *SubjectUseCase) GetSubjectByGroup(ctx context.Context, group *models.Group) (*[]models.Subject, error) {
	return subjectUseCase.subjectRepo.GetSubjectsByGroup(ctx, group)
}

func (subjectUseCase *SubjectUseCase) SubjectObjectListFromSubject(ctx context.Context, subject *models.Subject) (*[]models.SubjectObject, error) {
	return subjectUseCase.subjectRepo.GetSubjectObjectsBySubject(ctx, subject)
}

func (subjectUseCase *SubjectUseCase) GetSubjectById(ctx context.Context, id int) (*models.Subject, error) {
	return subjectUseCase.subjectRepo.GetSubject(ctx, id)
}

func (subjectUseCase *SubjectUseCase) GetAllSubject(ctx context.Context) (*[]models.Subject, error) {
	return subjectUseCase.subjectRepo.GetSubjects(ctx)
}
