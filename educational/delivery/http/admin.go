package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/educational"
	"net/http"
)

func (h *Handler) Overview(c *gin.Context) {
	overview, err := h.groupUseCase.GetOverview(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, overviewResponse{
		Stats: overviewStats{
			Groups:   overview.Groups,
			Subjects: overview.Subjects,
			Tasks:    overview.Tasks,
			Hidden:   overview.HiddenTasks,
		},
		RecentlyUpdated: toOverviewTasks(overview.RecentlyUpdated),
		Issues:          toOverviewTasks(overview.Issues),
	})
}

// Groups

func (h *Handler) ListGroups(c *gin.Context) {
	groups, err := h.groupUseCase.GetAllGroups(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}

	response := adminGroupsResponse{Groups: make([]adminGroup, 0, len(*groups))}
	for i := range *groups {
		response.Groups = append(response.Groups, toAdminGroup(&(*groups)[i]))
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetGroup(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	group, err := h.groupUseCase.GetGroupById(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	response := adminGroupResponse{
		Group:    toAdminGroup(group),
		Subjects: make([]adminSubject, 0, len(group.Subjects)),
	}
	for i := range group.Subjects {
		response.Subjects = append(response.Subjects, toAdminSubject(&group.Subjects[i]))
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) CreateGroup(c *gin.Context) {
	var request groupRequest
	if !bindJSON(c, &request) {
		return
	}

	input := educational.GroupInput{Name: valueOrEmpty(request.Name)}
	if request.Hidden != nil {
		input.Hidden = *request.Hidden
	}
	group, err := h.groupUseCase.CreateGroup(c.Request.Context(), input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toAdminGroup(group))
}

func (h *Handler) UpdateGroup(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var request groupRequest
	if !bindJSON(c, &request) {
		return
	}

	group, err := h.groupUseCase.UpdateGroup(c.Request.Context(), id, educational.GroupPatch{
		Name:   request.Name,
		Hidden: request.Hidden,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAdminGroup(group))
}

func (h *Handler) DeleteGroup(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.groupUseCase.DeleteGroup(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Subjects

func (h *Handler) GetSubject(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	subject, err := h.subjectUseCase.GetSubjectWithSubjectObjects(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	response := adminSubjectResponse{
		Group:   toAdminGroup(&subject.Group),
		Subject: toAdminSubject(subject),
		Tasks:   make([]adminTask, 0, len(subject.SubjectObjects)),
	}
	for i := range subject.SubjectObjects {
		response.Tasks = append(response.Tasks, h.toAdminTask(&subject.SubjectObjects[i]))
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) CreateSubject(c *gin.Context) {
	groupId, ok := idParam(c, "id")
	if !ok {
		return
	}
	var request subjectRequest
	if !bindJSON(c, &request) {
		return
	}

	subject, err := h.subjectUseCase.CreateSubject(c.Request.Context(), groupId, educational.SubjectInput{Name: request.Name})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, toAdminSubject(subject))
}

func (h *Handler) UpdateSubject(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var request subjectRequest
	if !bindJSON(c, &request) {
		return
	}

	subject, err := h.subjectUseCase.UpdateSubject(c.Request.Context(), id, educational.SubjectInput{Name: request.Name})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toAdminSubject(subject))
}

func (h *Handler) DeleteSubject(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.subjectUseCase.DeleteSubject(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Tasks

func (h *Handler) CreateTask(c *gin.Context) {
	subjectId, ok := idParam(c, "id")
	if !ok {
		return
	}
	var request taskRequest
	if !bindJSON(c, &request) {
		return
	}

	input := educational.SubjectObjectInput{
		Name:    valueOrEmpty(request.Name),
		Href:    valueOrEmpty(request.Href),
		Comment: valueOrEmpty(request.Comment),
	}
	if request.Hidden != nil {
		input.Hidden = *request.Hidden
	}
	if request.Categories != nil {
		input.Categories = *request.Categories
	}
	input.Password = valueOrEmpty(request.Password)
	subjectObject, err := h.subjectUseCase.CreateSubjectObject(c.Request.Context(), subjectId, input)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, h.toAdminTask(subjectObject))
}

func (h *Handler) UpdateTask(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var request taskRequest
	if !bindJSON(c, &request) {
		return
	}

	subjectObject, err := h.subjectUseCase.UpdateSubjectObject(c.Request.Context(), id, educational.SubjectObjectPatch{
		Name:       request.Name,
		Href:       request.Href,
		Comment:    request.Comment,
		Hidden:     request.Hidden,
		Categories: request.Categories,
		Password:   request.Password,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.toAdminTask(subjectObject))
}

func (h *Handler) DeleteTask(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if err := h.subjectUseCase.DeleteSubjectObject(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListCategories returns category names in use, for suggestions in the task form.
func (h *Handler) ListCategories(c *gin.Context) {
	names, err := h.subjectUseCase.GetCategoryNames(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	if names == nil {
		names = []string{}
	}
	c.JSON(http.StatusOK, categoriesResponse{Categories: names})
}

// RefreshTask downloads the document again. A failed download is not an error of the request:
// it is stored and returned in the task status.
func (h *Handler) RefreshTask(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	subjectObject, err := h.subjectUseCase.RefreshSubjectObjectContent(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, h.toAdminTask(subjectObject))
}
