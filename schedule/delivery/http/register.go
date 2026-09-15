package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/schedule"
)

// RegisterHTTPEndpoints registers the public presence on api and the sync state on admin.
func RegisterHTTPEndpoints(api *gin.RouterGroup, admin *gin.RouterGroup, useCase schedule.UseCase) {
	h := &Handler{useCase: useCase}

	api.GET("/presence", h.Presence)
	admin.GET("/schedule", h.SyncState)
	admin.POST("/schedule/sync", h.SyncNow)
}
