package http

import (
	"errors"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/auth"
	"log"
	"net/http"
	"strings"
)

type Handler struct {
	useCase auth.UseCase
}

func NewHandler(useCase auth.UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

type signInRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type userResponse struct {
	UserName string `json:"userName"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) SignIn(c *gin.Context) {
	var request signInRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{Error: "Некорректный запрос"})
		return
	}

	login := strings.TrimSpace(request.Login)
	if login == "" || strings.TrimSpace(request.Password) == "" {
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, errorResponse{Error: "Введите логин и пароль"})
		return
	}

	user, err := h.useCase.SignIn(c.Request.Context(), login, request.Password)
	if errors.Is(err, auth.ErrUserNotFound) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{Error: "Неверный логин или пароль"})
		return
	}
	if err != nil || user == nil {
		log.Printf("Failed to sign in: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{Error: "Не удалось войти, попробуйте ещё раз"})
		return
	}

	session := sessions.Default(c)
	session.Set(sessionUserIdKey, user.Id)
	session.Set(sessionUserNameKey, user.Username)
	if err := session.Save(); err != nil {
		log.Printf("Failed to save session: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{Error: "Не удалось войти, попробуйте ещё раз"})
		return
	}

	c.JSON(http.StatusOK, userResponse{UserName: user.Username})
}

func (h *Handler) SignOut(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		log.Printf("Failed to save session: %s", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{Error: "Не удалось выйти, попробуйте ещё раз"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Me returns the signed in user, or 401 for a guest.
func (h *Handler) Me(c *gin.Context) {
	userName, ok := currentUserName(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse{Error: "Нужно войти"})
		return
	}
	c.JSON(http.StatusOK, userResponse{UserName: userName})
}
