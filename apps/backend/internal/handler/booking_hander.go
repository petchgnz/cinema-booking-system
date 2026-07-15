package handler

import (
	"net/http"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/service"

	"github.com/gin-gonic/gin"
)

type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

// User selected seat > locked (temp)
func (h *BookingHandler) LockSeats(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req dto.LockSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.bookingService.LockSeats(c.Request.Context(), userID.(string), req); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "seats locked successfully",
		"expires_in":   "5 minutes",
		"seat_numbers": req.SeatNumbers,
	})
}

// CreateBooking - user confirm booking
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	booking, err := h.bookingService.CreateBooking(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, booking)
}
