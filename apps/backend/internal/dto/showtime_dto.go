package dto

import "time"

type CreateShowtimeRequest struct {
	MovieID   string    `json:"movie_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	Hall      string    `json:"hall" binding:"required"`
	SeatCount int       `json:"seat_count" binding:"required,min=1,max=200"`
}
