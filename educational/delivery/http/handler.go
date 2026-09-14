package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/educational"
	"github.com/kosipov/students/models"
	"net/http"
	"strconv"
	"strings"
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
	subjectList, err := h.useCase.GetSubjectsByGroup(c.Request.Context(), groupId)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &getSubjectResponse{
		Subjects: toSubjects(*subjectList),
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

	subjectObjectList, err := h.useCase.SubjectObjectListFromSubject(c.Request.Context(), subjectId)
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

	c.HTML(http.StatusOK, "home/index.html", gin.H{
		"index":    "Главная",
		"groups":   groups,
		"subjects": subjects,
	})
}

func (h *Handler) CreateSubject(c *gin.Context) {
	userName, _ := c.Get("user_name")
	groupId, err := strconv.Atoi(c.Param("group_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	subjectName := strings.TrimSpace(c.PostForm("subject_name"))
	if subjectName == "" {
		c.HTML(http.StatusUnprocessableEntity, "admin/form_subject.html", gin.H{
			"userName": userName,
			"groupId":  groupId,
			"message":  "Пустое имя предмета",
		})
		return
	}

	err = h.useCase.CreateSubject(c.Request.Context(), subjectName, groupId)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/groups/"+strconv.Itoa(groupId)+"/subjects")
}

func (h *Handler) CreateGroup(c *gin.Context) {
	userName, _ := c.Get("user_name")
	groupName := strings.TrimSpace(c.PostForm("group_name"))
	if groupName == "" {
		c.HTML(http.StatusUnprocessableEntity, "admin/form_group.html", gin.H{
			"userName": userName,
			"message":  "Пустое имя группы",
		})
		return
	}

	err := h.groupUseCase.CreateGroup(c.Request.Context(), groupName)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/groups")
}

func (h *Handler) ListHtmlGroups(c *gin.Context) {
	userName, _ := c.Get("user_name")
	listGroups, err := h.groupUseCase.GetAllGroups(c.Request.Context())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "admin/groups.html", gin.H{
		"userName": userName,
		"groups":   listGroups,
	})
}

func (h *Handler) ListHtmlSubjectsGroups(c *gin.Context) {
	userName, _ := c.Get("user_name")
	groupId, err := strconv.Atoi(c.Param("group_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	subjectList, err := h.useCase.GetSubjectsByGroup(c.Request.Context(), groupId)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "admin/subjects.html", gin.H{
		"userName":       userName,
		"subjectList":    subjectList,
		"currentGroupId": groupId,
	})
}

func (h *Handler) ListHtmlSubjectObject(c *gin.Context) {
	userName, _ := c.Get("user_name")
	subjectId, err := strconv.Atoi(c.Param("subject_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	subjectObjectList, err := h.useCase.SubjectObjectListFromSubject(c.Request.Context(), subjectId)
	if err != nil {
		if errors.Is(err, educational.ErrSubjectNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "admin/subject_objects.html", gin.H{
		"userName":          userName,
		"subjectObjectList": subjectObjectList,
		"subjectId":         subjectId,
	})
}

func (h *Handler) CreateSubjectObject(c *gin.Context) {
	userName, _ := c.Get("user_name")
	subjectId, err := strconv.Atoi(c.Param("subject_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	subjectObjectName := strings.TrimSpace(c.PostForm("subject_object_name"))
	subjectObjectHref := strings.TrimSpace(c.PostForm("subject_object_href"))

	if subjectObjectName == "" {
		c.HTML(http.StatusUnprocessableEntity, "admin/form_subject_object.html", gin.H{
			"userName":  userName,
			"subjectId": subjectId,
			"message":   "Пустое имя задания",
		})
		return
	}

	_, err = h.useCase.GetSubjectById(c.Request.Context(), subjectId)
	if err != nil {
		if errors.Is(err, educational.ErrSubjectNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	_, err = h.useCase.CreateSubjectObject(c.Request.Context(), subjectObjectName, subjectId, subjectObjectHref)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/subject/"+strconv.Itoa(subjectId))
}

func (h *Handler) EditSubjectObjectForm(c *gin.Context) {
	userName, _ := c.Get("user_name")
	subjectId, err := strconv.Atoi(c.Param("subject_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	subjectObjectId, err := strconv.Atoi(c.Param("subject_object_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	subjectObject, err := h.useCase.GetSubjectObject(c.Request.Context(), subjectId, subjectObjectId)
	if err != nil {
		if errors.Is(err, educational.ErrSubjectObjectNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "admin/form_subject_object.html", gin.H{
		"userName":      userName,
		"subjectId":     subjectId,
		"subjectObject": subjectObject,
	})
}

func (h *Handler) UpdateSubjectObject(c *gin.Context) {
	userName, _ := c.Get("user_name")
	subjectId, err := strconv.Atoi(c.Param("subject_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	subjectObjectId, err := strconv.Atoi(c.Param("subject_object_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	subjectObjectName := strings.TrimSpace(c.PostForm("subject_object_name"))
	subjectObjectHref := strings.TrimSpace(c.PostForm("subject_object_href"))

	if subjectObjectName == "" {
		c.HTML(http.StatusUnprocessableEntity, "admin/form_subject_object.html", gin.H{
			"userName":  userName,
			"subjectId": subjectId,
			"subjectObject": &models.SubjectObject{
				ID:        subjectObjectId,
				SubjectId: subjectId,
				Name:      subjectObjectName,
				Href:      subjectObjectHref,
			},
			"message": "Пустое имя задания",
		})
		return
	}

	_, err = h.useCase.UpdateSubjectObject(c.Request.Context(), subjectId, subjectObjectId, subjectObjectName, subjectObjectHref)
	if err != nil {
		if errors.Is(err, educational.ErrSubjectObjectNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Redirect(http.StatusSeeOther, "/admin/subject/"+strconv.Itoa(subjectId))
}

// DeleteSubjectObject is called via XHR from the subject objects page, so it responds with a status code only.
func (h *Handler) DeleteSubjectObject(c *gin.Context) {
	subjectId, err := strconv.Atoi(c.Param("subject_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	subjectObjectId, err := strconv.Atoi(c.Param("subject_object_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	err = h.useCase.DeleteSubjectObject(c.Request.Context(), subjectId, subjectObjectId)
	if err != nil {
		if errors.Is(err, educational.ErrSubjectObjectNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusNoContent)
}

func toSubjects(subjects []models.Subject) []models.Subject {
	out := make([]models.Subject, len(subjects))

	for i, s := range subjects {
		out[i] = toSubject(s)
	}

	return out
}

func toSubject(s models.Subject) models.Subject {
	return models.Subject{
		ID:          s.ID,
		SubjectName: s.SubjectName,
		GroupId:     s.GroupId,
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
