package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/auth/delivery/http"
	"github.com/kosipov/students/educational"
	http2 "net/http"
)

func RegisterHTTPEndpoints(router *gin.Engine, uc educational.CommonSubjectUseCase, gc educational.CommonGroupUseCase) {
	h := NewHandler(uc, gc)

	router.GET("/groups", h.ListGroups)
	router.GET("/groups/:group_id/subjects", h.ListSubject)

	router.GET("/subjects/:subject_id", h.ListSubjectObject)

	router.GET("/", h.IndexPage)

	router.POST("/admin/groups/:group_id/subjects/", h.CreateSubject)

	adminEndpoints := router.Group("/admin")
	adminEndpoints.Use(http.NewAuthMiddleware())
	{
		adminEndpoints.GET("/groups", h.ListHtmlGroups)
		adminEndpoints.GET("/groups/:group_id/subjects", h.ListHtmlSubjectsGroups)
		adminEndpoints.GET("/groups/:group_id/subject/create", func(c *gin.Context) {
			c.HTML(http2.StatusOK, "admin/form_subject.html", gin.H{
				"groupId": c.Param("group_id"),
			})
		})
		adminEndpoints.POST("groups/:group_id/subject/create", h.CreateSubject)
		adminEndpoints.GET("/subject/:subject_id", h.ListHtmlSubjectObject)
		adminEndpoints.GET("/subject/:subject_id/subject_object/create", func(c *gin.Context) {
			c.HTML(http2.StatusOK, "admin/form_subject_object.html", gin.H{
				"subjectId": c.Param("subject_id"),
			})
		})
		adminEndpoints.POST("/subject/:subject_id/subject_object/create", h.CreateSubjectObject)
		adminEndpoints.DELETE("/subject/:subject_id/subject_object/:subject_object_id", h.DeleteSubjectObject)
	}
}
