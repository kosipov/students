package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/markdown"
	"net/http"
)

// Catalog returns everything students can see: visible groups with their subjects and visible tasks.
func (h *Handler) Catalog(c *gin.Context) {
	groups, err := h.groupUseCase.GetVisibleGroups(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

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
				catalogSubject.Tasks = append(catalogSubject.Tasks, catalogTask{
					ID:         subjectObject.ID,
					Name:       subjectObject.Name,
					Comment:    subjectObject.Comment,
					Href:       safeHref(subjectObject.Href),
					IsDocument: h.subjectUseCase.IsDocument(subjectObject),
					Categories: subjectObject.CategoryNames(),
				})
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
	task, err := h.subjectUseCase.GetTask(c.Request.Context(), id)
	h.respondTask(c, task, err)
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
	isDocument := true
	switch {
	case errors.Is(err, educational.ErrSubjectObjectNotDocument):
		// The task is returned without a document, the client opens its link.
		isDocument = false
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
		Group:      ref{ID: int(task.Group.ID), Name: task.Group.GroupName},
		Subject:    ref{ID: task.Subject.ID, Name: task.Subject.SubjectName},
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
