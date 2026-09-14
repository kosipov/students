package http

import (
	"fmt"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	sessionUserIdKey   = "user_id"
	sessionUserNameKey = "user_name"
)

type AuthMiddleware struct {
}

// NewAuthMiddleware lets only signed in users through; others get 401.
func NewAuthMiddleware() gin.HandlerFunc {
	return (&AuthMiddleware{}).Handle
}

func (m *AuthMiddleware) Handle(c *gin.Context) {
	userName, ok := currentUserName(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{Error: "Нужно войти"})
		return
	}
	c.Set(sessionUserNameKey, userName)
	c.Next()
}

func currentUserName(c *gin.Context) (string, bool) {
	session := sessions.Default(c)
	if session.Get(sessionUserIdKey) == nil {
		return "", false
	}
	return fmt.Sprint(session.Get(sessionUserNameKey)), true
}
