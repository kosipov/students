package http

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthMiddleware struct {
}

func NewAuthMiddleware() gin.HandlerFunc {
	return (&AuthMiddleware{}).Handle
}

func (m *AuthMiddleware) Handle(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user_id")
	if user == nil {
		c.Redirect(http.StatusMovedPermanently, "/auth/sign-in")
		c.Abort()
		return
	}
	c.Set("user_name", session.Get("user_name"))
	c.Next()
}
