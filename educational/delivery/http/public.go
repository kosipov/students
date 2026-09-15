package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/markdown"
	"log"
	"math"
	"net/http"
	"strconv"
)

// Catalog returns everything students can see: visible groups with their subjects and visible tasks.
func (h *Handler) Catalog(c *gin.Context) {
	groups, err := h.groupUseCase.GetVisibleGroups(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	unlocked := readUnlockedTasks(c)
	response := catalogResponse{Groups: make([]catalogGroup, 0, len(*groups))}
	for _, group := range *groups {
		catalogGroup := catalogGroup{
			ID:       int(group.ID),
			Name:     group.GroupName,
			Subjects: make([]catalogSubject, 0, len(group.Subjects)),
		}
		for _, subject := range group.Subjects {
			catalogSubject := catalogSubject{
				ID:    subject.ID,
				Name:  subject.SubjectName,
				Tasks: make([]catalogTask, 0, len(subject.SubjectObjects)),
			}
			for i := range subject.SubjectObjects {
				subjectObject := &subject.SubjectObjects[i]
				task := catalogTask{
					ID:         subjectObject.ID,
					Name:       subjectObject.Name,
					Comment:    subjectObject.Comment,
					Href:       safeHref(subjectObject.Href),
					IsDocument: h.subjectUseCase.IsDocument(subjectObject),
					Categories: subjectObject.CategoryNames(),
				}
				if subjectObject.IsProtected() && !unlocked.IsUnlocked(subjectObject.ID, subjectObject.PasswordVersion) {
					// The link opens the file without the password, so it is given out only after unlocking.
					task.Href = ""
					task.Locked = true
				}
				catalogSubject.Tasks = append(catalogSubject.Tasks, task)
			}
			catalogGroup.Subjects = append(catalogGroup.Subjects, catalogSubject)
		}
		response.Groups = append(response.Groups, catalogGroup)
	}

	c.JSON(http.StatusOK, response)
}

// Task returns a task visible to students with its rendered document.
func (h *Handler) Task(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	task, err := h.subjectUseCase.GetTask(c.Request.Context(), id, readUnlockedTasks(c))
	h.respondTask(c, task, err)
}

// UnlockTask checks the password of a task and remembers in the session that the student entered it.
func (h *Handler) UnlockTask(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var request unlockRequest
	if !bindJSON(c, &request) {
		return
	}

	clientKey := clientIP(c) + " " + strconv.Itoa(id)
	taskKey := strconv.Itoa(id)
	for _, limit := range []struct {
		limiter *attemptLimiter
		key     string
	}{{h.clientAttempts, clientKey}, {h.taskAttempts, taskKey}} {
		if allowed, retryAfter := limit.limiter.Allow(limit.key); !allowed {
			minutes := int(math.Ceil(retryAfter.Minutes()))
			c.Header("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, errorResponse{
				Error: "Слишком много попыток. Попробуйте через " + strconv.Itoa(minutes) + " мин.",
			})
			return
		}
	}

	subjectObject, err := h.subjectUseCase.UnlockTask(c.Request.Context(), id, request.Password)
	if errors.Is(err, educational.ErrWrongPassword) {
		h.clientAttempts.Fail(clientKey)
		h.taskAttempts.Fail(taskKey)
		log.Printf("Wrong password for task %d from %s", id, clientIP(c))
	}
	if err != nil {
		respondError(c, err)
		return
	}

	h.clientAttempts.Reset(clientKey)
	if subjectObject.IsProtected() {
		if err := saveUnlockedTask(c, subjectObject.ID, subjectObject.PasswordVersion); err != nil {
			respondError(c, err)
			return
		}
	}
	c.Status(http.StatusNoContent)
}

// PreviewTask works like Task, but shows hidden tasks to the admin.
func (h *Handler) PreviewTask(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	task, err := h.subjectUseCase.PreviewTask(c.Request.Context(), id)
	h.respondTask(c, task, err)
}

func (h *Handler) respondTask(c *gin.Context, task *educational.Task, err error) {
	isDocument, locked := true, false
	switch {
	case errors.Is(err, educational.ErrSubjectObjectNotDocument):
		// The task is returned without a document, the client opens its link.
		isDocument = false
	case errors.Is(err, educational.ErrTaskLocked):
		// Only what is shown in lists: the client asks for the password.
		locked = true
	case err != nil:
		respondError(c, err)
		return
	}

	subjectObject := task.SubjectObject
	response := taskResponse{
		ID:         subjectObject.ID,
		Name:       subjectObject.Name,
		Comment:    subjectObject.Comment,
		Href:       safeHref(subjectObject.Href),
		Hidden:     subjectObject.Hidden || task.Group.Hidden,
		IsDocument: isDocument,
		Protected:  subjectObject.IsProtected(),
		Locked:     locked,
		Group:      ref{ID: int(task.Group.ID), Name: task.Group.GroupName},
		Subject:    ref{ID: task.Subject.ID, Name: task.Subject.SubjectName},
	}
	if locked {
		response.Href = ""
		response.IsDocument = false
		c.JSON(http.StatusOK, response)
		return
	}

	if isDocument && subjectObject.Content != "" {
		content, err := markdown.Render([]byte(subjectObject.Content))
		if err != nil {
			respondError(c, err)
			return
		}
		html := string(content)
		response.ContentHTML = &html
		response.UpdatedAt = subjectObject.ContentFetchedAt
	}

	c.JSON(http.StatusOK, response)
}
