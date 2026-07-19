package dto

type LockSeatRequest struct {
	ShowtimeID  string   `json:"showtime_id" binding:"required"`
	SeatNumbers []string `json:"seat_numbers" binding:"required,min=1"`
}

type CreateBookingRequest struct {
	ShowtimeID  string   `json:"showtime_id" binding:"required"`
	SeatNumbers []string `json:"seat_numbers" binding:"required,min=1"`
}
