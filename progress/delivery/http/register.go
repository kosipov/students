package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/progress"
)

// RegisterHTTPEndpoints registers the public API on api and the admin API on admin,
// which is expected to be protected by the auth middleware.
func RegisterHTTPEndpoints(api *gin.RouterGroup, admin *gin.RouterGroup, adminUseCase progress.AdminUseCase, publicUseCase progress.PublicUseCase) {
	h := NewHandler(adminUseCase, publicUseCase)

	api.GET("/groups/:id/students", h.Roster)
	api.GET("/students/:id/progress", h.StudentProgress)

	admin.GET("/groups/:id/progress", h.Matrix)
	admin.POST("/groups/:id/students", h.AddStudent)
	admin.POST("/groups/:id/students/import", h.ImportStudents)
	admin.POST("/groups/:id/students/import-campus", h.ImportStudentsFromCampus)
	admin.PATCH("/students/:id", h.RenameStudent)
	admin.DELETE("/students/:id", h.RemoveStudent)
	admin.PUT("/students/:id/marks/:taskId", h.SetMark)
}

type Handler struct {
	adminUseCase  progress.AdminUseCase
	publicUseCase progress.PublicUseCase
}

func NewHandler(adminUseCase progress.AdminUseCase, publicUseCase progress.PublicUseCase) *Handler {
	return &Handler{adminUseCase: adminUseCase, publicUseCase: publicUseCase}
}
