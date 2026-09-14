package http

import (
	"github.com/kosipov/students/educational"
)

type Handler struct {
	subjectUseCase educational.CommonSubjectUseCase
	groupUseCase   educational.CommonGroupUseCase
}

func NewHandler(subjectUseCase educational.CommonSubjectUseCase, groupUseCase educational.CommonGroupUseCase) *Handler {
	return &Handler{
		subjectUseCase: subjectUseCase,
		groupUseCase:   groupUseCase,
	}
}
