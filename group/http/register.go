package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/group"
)

func RegisterHTTPEndpoints(router *gin.Engine, gc group.UseCase) {
	h := NewHandler(gc)

	router.GET("/groups", h.ListGroups)
}
