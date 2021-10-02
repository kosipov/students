package http

import (
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/auth"
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

func (h *Handler) SignIn(c *gin.Context) {
	session := sessions.Default(c)
	login := c.PostForm("login")
	password := c.PostForm("password")

	if strings.Trim(login, " ") == "" || strings.Trim(password, " ") == "" {
		c.HTML(http.StatusUnprocessableEntity, "home/login.html", gin.H{
			"message": "Пустой логин или пароль",
		})
		return
	}

	user, err := h.useCase.SignIn(c.Request.Context(), login, password)
	if err != nil || user == nil {
		if err == auth.ErrUserNotFound {
			c.HTML(http.StatusUnauthorized, "home/login.html", gin.H{
				"message": "Неверный логин или пароль",
			})
			return
		}

		c.HTML(http.StatusInternalServerError, "home/login.html", gin.H{
			"message": "Неизвестная ошибка!",
		})
		return
	}

	session.Set("user_id", user.Id)
	session.Set("user_name", user.Username)

	if err := session.Save(); err != nil {
		c.HTML(http.StatusInternalServerError, "home/login.html", gin.H{
			"message": "Неизвестная ошибка!",
		})
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/admin/groups")
}

func (h *Handler) SignOut(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get("user_id")
	if userId == nil {
		c.HTML(http.StatusInternalServerError, "home/login.html", gin.H{
			"message": "Некорректная сессия",
		})
	}
	session.Clear()

	if err := session.Save(); err != nil {
		c.HTML(http.StatusInternalServerError, "home/login.html", gin.H{
			"message": "Ошибка выхода",
		})
	}

	c.Redirect(http.StatusTemporaryRedirect, "/auth/sign-in")
}
