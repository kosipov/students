package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/progress"
)

// Admin

func (h *Handler) Matrix(c *gin.Context) {
	groupId, ok := idParam(c, "id")
	if !ok {
		return
	}
	matrix, err := h.adminUseCase.GetMatrix(c.Request.Context(), groupId)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toMatrixResponse(matrix))
}

func (h *Handler) AddStudent(c *gin.Context) {
	groupId, ok := idParam(c, "id")
	if !ok {
		return
	}
	var request studentRequest
	if !bindJSON(c, &request) {
		return
	}
	student, err := h.adminUseCase.AddStudent(c.Request.Context(), groupId, progress.StudentInput{Name: valueOrEmpty(request.Name)})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toStudentJSON(student, nil))
}

func (h *Handler) ImportStudents(c *gin.Context) {
	groupId, ok := idParam(c, "id")
	if !ok {
		return
	}
	var request importRequest
	if !bindJSON(c, &request) {
		return
	}
	result, err := h.adminUseCase.ImportStudents(c.Request.Context(), groupId, request.Names)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"added": result.Added, "skipped": result.Skipped, "found": result.Found})
}

// ImportStudentsFromCampus loads the roster of the group from the university's campus site.
func (h *Handler) ImportStudentsFromCampus(c *gin.Context) {
	groupId, ok := idParam(c, "id")
	if !ok {
		return
	}
	result, err := h.adminUseCase.ImportFromCampus(c.Request.Context(), groupId)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"added": result.Added, "skipped": result.Skipped, "found": result.Found})
}

func (h *Handler) RenameStudent(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var request studentRequest
	if !bindJSON(c, &request) {
		return
	}
	student, err := h.adminUseCase.RenameStudent(c.Request.Context(), id, progress.StudentInput{Name: valueOrEmpty(request.Name)})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toStudentJSON(student, nil))
}

func (h *Handler) RemoveStudent(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.adminUseCase.RemoveStudent(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) SetMark(c *gin.Context) {
	studentId, ok := idParam(c, "id")
	if !ok {
		return
	}
	taskId, ok := idParam(c, "taskId")
	if !ok {
		return
	}
	var request markRequest
	if !bindJSON(c, &request) {
		return
	}
	err := h.adminUseCase.SetMark(c.Request.Context(), studentId, taskId, progress.MarkInput{
		Status:  valueOrEmpty(request.Status),
		Comment: valueOrEmpty(request.Comment),
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Public

func (h *Handler) Roster(c *gin.Context) {
	groupId, ok := idParam(c, "id")
	if !ok {
		return
	}
	group, students, err := h.publicUseCase.GetRoster(c.Request.Context(), groupId)
	if err != nil {
		respondError(c, err)
		return
	}
	response := rosterResponse{
		Group:    groupJSON{ID: int(group.ID), Name: group.GroupName},
		Students: make([]studentBrief, 0, len(students)),
	}
	for i := range students {
		response.Students = append(response.Students, studentBrief{ID: students[i].ID, Name: students[i].Name})
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) StudentProgress(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	studentProgress, err := h.publicUseCase.GetStudentProgress(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toProgressResponse(studentProgress))
}
