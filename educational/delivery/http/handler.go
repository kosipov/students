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
	subjectName := c.PostForm("subject_name")
	if subjectName == "" {
		c.HTML(http.StatusUnprocessableEntity, "admin/subjects.html", gin.H{
			"message": "Пустое имя предмета",
		})
	}
	groupId, err := strconv.Atoi(c.PostForm("group_id"))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = h.useCase.CreateSubject(c.Request.Context(), subjectName, groupId)
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
		"message":        "Предмет успешно создан",
		"subjectList":    subjectList,
		"currentGroupId": groupId,
		"userName":       userName,
	})
}

func (h *Handler) CreateGroup(c *gin.Context) {
	groupName := c.PostForm("group_name")
	if groupName == "" {
		c.HTML(http.StatusUnprocessableEntity, "admin/subject.html", gin.H{
			"message": "Пустое имя группы",
		})
	}
	err := h.groupUseCase.CreateGroup(c, groupName)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.HTML(http.StatusOK, "admin/group.html", gin.H{
		"message": "Группа успешно создана",
	})
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

	subjectObjectList, err := h.useCase.SubjectObjectListFromSubject(c.Request.Context(), subjectId)

	if err != nil {
		c.HTML(http.StatusUnprocessableEntity, "admin/subject_objects.html", gin.H{
			"userName":          userName,
			"subjectObjectList": nil,
			"subjectId":         subjectId,
			"message":           "Дисциплина не найдена",
		})
	}

	c.HTML(http.StatusOK, "admin/subject_objects.html", gin.H{
		"userName":          userName,
		"subjectObjectList": subjectObjectList,
		"subjectId":         subjectId,
	})
}

func (h *Handler) CreateSubjectObject(context *gin.Context) {
	userName, _ := context.Get("user_name")
	subjectIdForm := context.PostForm("subject_id")
	subjectObjectName := context.PostForm("subject_object_name")
	subjectObjectHref := context.PostForm("subject_object_href")

	subjectId, _ := strconv.Atoi(subjectIdForm)
	_, err := h.useCase.GetSubjectById(context.Request.Context(), subjectId)
	if err != nil {
		context.HTML(http.StatusUnprocessableEntity, "admin/form_subject_object.html", gin.H{
			"userName":          userName,
			"subjectObjectList": nil,
			"message":           "Дисциплина не найдена",
		})
	}

	_, err = h.useCase.CreateSubjectObject(context.Request.Context(), subjectObjectName, subjectId, subjectObjectHref)
	if err != nil {
		context.HTML(http.StatusUnprocessableEntity, "admin/form_subject_object.html", gin.H{
			"userName":          userName,
			"subjectObjectList": nil,
			"message":           "Задание не создано",
		})
	}

	context.Redirect(http.StatusMovedPermanently, "/admin/subject/"+subjectIdForm)
}

func (h *Handler) DeleteSubjectObject(context *gin.Context) {
	userName, _ := context.Get("user_name")
	subjectObjectId, err := strconv.Atoi(context.Param("subject_object_id"))
	subjectId, err := strconv.Atoi(context.Param("subject_id"))
	if err != nil {
		context.HTML(http.StatusUnprocessableEntity, "admin/subject_objects.html", gin.H{
			"userName":          userName,
			"subjectObjectList": nil,
			"message":           "Задание не найдено",
		})
	}

	err = h.useCase.DeleteSubjectObject(context.Request.Context(), subjectObjectId)

	if err != nil {
		context.HTML(http.StatusUnprocessableEntity, "admin/subject_objects.html", gin.H{
			"userName":          userName,
			"subjectObjectList": nil,
			"message":           "Ошибка удаления",
		})
	}
	subjectObjectList, err := h.useCase.SubjectObjectListFromSubject(context.Request.Context(), subjectId)

	context.HTML(http.StatusOK, "admin/subject_objects.html", gin.H{
		"userName":          userName,
		"subjectObjectList": subjectObjectList,
		"subjectId":         subjectId,
	})
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
