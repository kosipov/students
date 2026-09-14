package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/auth"
)

// RegisterHTTPEndpoints registers sign in endpoints on api, which is expected to have the sessions middleware.
func RegisterHTTPEndpoints(api *gin.RouterGroup, uc auth.UseCase) {
	h := NewHandler(uc)

	authEndpoints := api.Group("/auth")
	{
		authEndpoints.POST("/sign-in", h.SignIn)
		authEndpoints.POST("/sign-out", h.SignOut)
		authEndpoints.GET("/me", h.Me)
	}
}
