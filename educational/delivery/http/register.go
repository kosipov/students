package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/educational"
)

func RegisterHTTPEndpoints(router *gin.Engine, uc educational.CommonSubjectUseCase, gc educational.CommonGroupUseCase) {
	h := NewHandler(uc, gc)

	router.GET("/groups", h.ListGroups)
	router.GET("/groups/:group_id/subjects", h.ListSubject)

	router.GET("/subjects/:subject_id", h.ListSubjectObject)

	router.GET("/", h.IndexPage)
}
