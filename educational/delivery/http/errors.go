package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/educational"
	"log"
	"net/http"
	"strconv"
)

type errorResponse struct {
	Error string `json:"error"`
	// Fields holds messages for invalid input fields.
	Fields map[string]string `json:"fields,omitempty"`
}

// respondError writes the error as JSON. Messages are shown to users as is.
func respondError(c *gin.Context, err error) {
	var validationErr *educational.ValidationError
	switch {
	case errors.As(err, &validationErr):
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, errorResponse{
			Error:  "Проверьте введённые данные",
			Fields: validationErr.Fields,
		})
	case errors.Is(err, educational.ErrGroupNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, errorResponse{Error: "Группа не найдена"})
	case errors.Is(err, educational.ErrSubjectNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, errorResponse{Error: "Предмет не найден"})
	case errors.Is(err, educational.ErrSubjectObjectNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, errorResponse{Error: "Задание не найдено"})
	case errors.Is(err, educational.ErrWrongPassword):
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, errorResponse{Error: "Неверный пароль", Fields: map[string]string{"password": "Неверный пароль"}})
	case errors.Is(err, educational.ErrSubjectObjectNotDocument):
		c.AbortWithStatusJSON(http.StatusConflict, errorResponse{Error: "Это задание открывается по ссылке, загружать нечего"})
	default:
		log.Printf("%s %s: %s", c.Request.Method, c.Request.URL.Path, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{Error: "Что-то пошло не так, попробуйте ещё раз"})
	}
}

// idParam reads a numeric path parameter and responds with 404 when it is not a number.
func idParam(c *gin.Context, name string) (int, bool) {
	id, err := strconv.Atoi(c.Param(name))
	if err != nil || id <= 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, errorResponse{Error: "Не найдено"})
		return 0, false
	}
	return id, true
}

func bindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{Error: "Некорректный запрос"})
		return false
	}
	return true
}
