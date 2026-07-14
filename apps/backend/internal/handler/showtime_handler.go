package handler

import (
	"net/http"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/service"

	"github.com/gin-gonic/gin"
)

type ShowtimeHandler struct {
	showtimeService service.ShowtimeService
}

func NewShowtimeHandler(showtimeService service.ShowtimeService) *ShowtimeHandler {
	return &ShowtimeHandler{showtimeService: showtimeService}
}

func (h *ShowtimeHandler) Create(c *gin.Context) {
	var req dto.CreateShowtimeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	showtime, err := h.showtimeService.CreateShowtime(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, showtime)
}

func (h *ShowtimeHandler) GetAll(c *gin.Context) {
	showtimes, err := h.showtimeService.GetAllShowtimes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, showtimes)
}

func (h *ShowtimeHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	showtime, err := h.showtimeService.GetShowtimeByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"Error": "showtime not found"})
		return
	}

	c.JSON(http.StatusOK, showtime)
}