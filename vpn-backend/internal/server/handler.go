package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vpn/backend/internal/wireguard"
)

type Handler struct {
	repo        *Repository
	healthChecker *HealthChecker
}

func NewHandler(repo *Repository, healthChecker *HealthChecker) *Handler {
	return &Handler{repo: repo, healthChecker: healthChecker}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/servers", h.ListServers)
	r.GET("/servers/:id", h.GetServer)
	r.GET("/servers/:id/stats", h.GetServerStats)
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
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server ID"})
		return
	}

	server, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "server not found"})
		return
	}
	c.JSON(http.StatusOK, server)
}

func (h *Handler) GetServerStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid server ID"})
		return
	}

	if h.healthChecker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "health checker not available"})
		return
	}

	stats, err := h.healthChecker.GetServerStats(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}