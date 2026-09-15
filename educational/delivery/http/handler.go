package http

import (
	"github.com/kosipov/students/educational"
	"time"
)

const (
	// Guessing a task password: a student gets 10 tries per 15 minutes, and all students together
	// get 100 per task, so a guesser changing addresses is slowed down too.
	passwordAttemptsPerClient = 10
	passwordAttemptsPerTask   = 100
	passwordAttemptsWindow    = 15 * time.Minute
)

type Handler struct {
	subjectUseCase educational.CommonSubjectUseCase
	groupUseCase   educational.CommonGroupUseCase

	clientAttempts *attemptLimiter
	taskAttempts   *attemptLimiter
}

func NewHandler(subjectUseCase educational.CommonSubjectUseCase, groupUseCase educational.CommonGroupUseCase) *Handler {
	return &Handler{
		subjectUseCase: subjectUseCase,
		groupUseCase:   groupUseCase,
		clientAttempts: newAttemptLimiter(passwordAttemptsPerClient, passwordAttemptsWindow),
		taskAttempts:   newAttemptLimiter(passwordAttemptsPerTask, passwordAttemptsWindow),
	}
}
