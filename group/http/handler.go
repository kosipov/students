package http

import (
	"github.com/gin-gonic/gin"
	"github.com/kosipov/students/group"
	"github.com/kosipov/students/models"
	"net/http"
)

type Handler struct {
	useCase group.UseCase
}

func NewHandler(useCase group.UseCase) *Handler {
	return &Handler{
		useCase: useCase,
	}
}

type getResponse struct {
	Groups []models.Group `json:"groups"`
}

func (h *Handler) ListGroups(c *gin.Context) {
	listGroups, err := h.useCase.GetAllGroups(c.Request.Context())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, &getResponse{
		Groups: toGroups(*listGroups),
	})

}

func toGroups(groups []models.Group) []models.Group {
	out := make([]models.Group, len(groups))

	for i, g := range groups {
		out[i] = toGroup(g)
	}

	return out
}

func toGroup(g models.Group) models.Group {
	return models.Group{
		ID:        g.ID,
		GroupName: g.GroupName,
	}
}
