package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/educational"
)

// RegisterHTTPEndpoints registers the public API on api and the admin API on admin,
// which is expected to be protected by the auth middleware.
func RegisterHTTPEndpoints(api *gin.RouterGroup, admin *gin.RouterGroup, subjectUseCase educational.CommonSubjectUseCase, groupUseCase educational.CommonGroupUseCase) {
	h := NewHandler(subjectUseCase, groupUseCase)

	api.GET("/catalog", h.Catalog)
	api.GET("/tasks/:id", h.Task)

	admin.GET("/overview", h.Overview)

	admin.GET("/groups", h.ListGroups)
	admin.POST("/groups", h.CreateGroup)
	admin.GET("/groups/:id", h.GetGroup)
	admin.PATCH("/groups/:id", h.UpdateGroup)
	admin.DELETE("/groups/:id", h.DeleteGroup)
	admin.POST("/groups/:id/subjects", h.CreateSubject)

	admin.GET("/subjects/:id", h.GetSubject)
	admin.PATCH("/subjects/:id", h.UpdateSubject)
	admin.DELETE("/subjects/:id", h.DeleteSubject)
	admin.POST("/subjects/:id/tasks", h.CreateTask)

	admin.PATCH("/tasks/:id", h.UpdateTask)
	admin.DELETE("/tasks/:id", h.DeleteTask)
	admin.POST("/tasks/:id/refresh", h.RefreshTask)
	admin.GET("/tasks/:id/preview", h.PreviewTask)
}
