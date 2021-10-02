package http

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/auth"
	"net/http"
)

func RegisterHTTPEndpoints(router *gin.Engine, uc auth.UseCase) {
	h := NewHandler(uc)
	store := sessions.NewCookieStore([]byte("secret"))
	router.Use(sessions.Sessions("student-session", store))

	authEndpoints := router.Group("/auth")
	{
		authEndpoints.GET("/sign-out", h.SignOut)
		authEndpoints.POST("/sign-in", h.SignIn)
		authEndpoints.GET("/sign-in", func(context *gin.Context) {
			context.HTML(http.StatusOK, "home/login.html", gin.H{})
		})
	}
}
