package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SeatStatus string

const (
	SeatAvailable SeatStatus = "available"
	SeatLocked    SeatStatus = "locked"
	SeatBooked    SeatStatus = "booked"
)

type Seat struct {
	SeatNumber string     `bson:"seat_number" json:"seat_number"`
	Status     SeatStatus `bson:"status" json:"status"`
}

type Showtime struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MovieID   primitive.ObjectID `bson:"movie_id" json:"movie_id"`
	StartTime time.Time          `bson:"start_time" json:"start_time"`
	EndTime   time.Time          `bson:"end_time" json:"end_time"`
	Hall      string             `bson:"hall" json:"hall"`
	Seats     []Seat             `bson:"seats" json:"seats"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
