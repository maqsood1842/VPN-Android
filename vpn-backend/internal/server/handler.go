package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/servers", h.ListServers)
	r.GET("/servers/:id", h.GetServer)
}

func (h *Handler) ListServers(c *gin.Context) {
	servers, err := h.repo.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, servers)
}

func (h *Handler) GetServer(c *gin.Context) {
	idStr := c.Param("id")
	// UUID parsing would go here
	c.JSON(http.StatusOK, gin.H{"id": idStr})
}