package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/labs"
)

// RegisterHTTPEndpoints registers the labs occupancy on admin: it is only for the signed in admin.
func RegisterHTTPEndpoints(admin *gin.RouterGroup, useCase *labs.UseCase) {
	h := &Handler{useCase: useCase}

	admin.GET("/labs", h.Occupancy)
}
