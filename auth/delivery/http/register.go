package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/auth"
)

func RegisterHTTPEndpoints(router *gin.Engine, uc auth.UseCase) {
	h := NewHandler(uc)

	authEndpoints := router.Group("/auth")
	{
		authEndpoints.POST("/sign-up", h.SignUp)
	}
}
