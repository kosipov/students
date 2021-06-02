package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"net/http"
	"strconv"
)

type Handler struct {
	useCase      educational.CommonSubjectUseCase
	groupUseCase educational.CommonGroupUseCase
}

func NewHandler(useCase educational.CommonSubjectUseCase, groupUseCase educational.CommonGroupUseCase) *Handler {
	return &Handler{
		useCase:      useCase,
		groupUseCase: groupUseCase,
	}
}

type getSubjectResponse struct {
	Subjects []models.Subject `json:"subjects"`
}

type getGroupResponse struct {
	Groups []models.Group `json:"groups"`
}

type getSubjectObjectResponse struct {
	SubjectObject []models.SubjectObject `json:"subject_object"`
}

func (h *Handler) ListSubject(c *gin.Context) {
	groupId, err := strconv.Atoi(c.Param("group_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	group, err := h.groupUseCase.GetGroupById(c.Request.Context(), groupId)
	subjectList, err := h.useCase.GetSubjectByGroup(c.Request.Context(), group)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &getSubjectResponse{
		Subjects: toSubjects(*subjectList, *group),
	})

}

func (h *Handler) ListGroups(c *gin.Context) {
	listGroups, err := h.groupUseCase.GetAllGroups(c.Request.Context())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &getGroupResponse{
		Groups: toGroups(*listGroups),
	})
}

func (h *Handler) ListSubjectObject(c *gin.Context) {
	subjectId, err := strconv.Atoi(c.Param("subject_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	subject, err := h.useCase.GetSubjectById(c.Request.Context(), subjectId)
	subjectObjectList, err := h.useCase.SubjectObjectListFromSubject(c.Request.Context(), subject)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &getSubjectObjectResponse{
		SubjectObject: toSubjectObjects(*subjectObjectList),
	})
}

func (h *Handler) IndexPage(c *gin.Context) {
	groups, err := h.groupUseCase.GetAllGroups(c.Request.Context())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	subjects, err := h.useCase.GetAllSubject(c.Request.Context())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "index.html", gin.H{
		"index":    "Главная",
		"groups":   groups,
		"subjects": subjects,
	})
}

func toSubjects(subjects []models.Subject, group models.Group) []models.Subject {
	out := make([]models.Subject, len(subjects))

	for i, s := range subjects {
		out[i] = toSubject(s, group)
	}

	return out
}

func toSubject(s models.Subject, g models.Group) models.Subject {
	return models.Subject{
		ID:          s.ID,
		SubjectName: s.SubjectName,
		GroupId:     s.GroupId,
		Group:       g,
	}
}

func toGroups(groups []models.Group) []models.Group {
	out := make([]models.Group, len(groups))

	for i, g := range groups {
		out[i] = toGroup(g)
	}

	return out
}

func toGroup(g models.Group) models.Group {
	return models.Group{
		ID:        g.ID,
		GroupName: g.GroupName,
	}
}

func toSubjectObjects(subjectObjects []models.SubjectObject) []models.SubjectObject {
	out := make([]models.SubjectObject, len(subjectObjects))

	for i, so := range subjectObjects {
		out[i] = toSubjectObject(so)
	}

	return out
}

func toSubjectObject(so models.SubjectObject) models.SubjectObject {
	return models.SubjectObject{
		ID:      so.ID,
		Name:    so.Name,
		Comment: so.Comment,
	}
}
