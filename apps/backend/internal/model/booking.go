package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookingStatus string

const (
	BookingPending   BookingStatus = "pending"
	BookingConfirmed BookingStatus = "confirmed"
	BookingCancelled BookingStatus = "cancelled"
)

type Booking struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	ShowtimeID  primitive.ObjectID `bson:"showtime_id" json:"showtime_id"`
	SeatNumbers []string           `bson:"seat_numbers" json:"seat_numbers"`
	Status      BookingStatus      `bson:"status" json:"status"`
	TotalPrice  float64            `bson:"total_price" json:"total_price"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
