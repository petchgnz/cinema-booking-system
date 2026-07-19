package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditEventType string

const (
	AuditBookingSuccess AuditEventType = "booking_success"
	AuditLockFail       AuditEventType = "lock_fail"
	AuditSeatReleased   AuditEventType = "seat_released"
	AuditSystemError    AuditEventType = "system_error"
)

type AuditLog struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EventType   AuditEventType     `bson:"event_type" json:"event_type"`
	UserID      string             `bson:"user_id" json:"user_id"`
	ShowtimeID  string             `bson:"showtime_id" json:"showtime_id"`
	SeatNumbers []string           `bson:"seat_numbers,omitempty" json:"seat_numbers,omitempty"`
	Message     string             `bson:"message" json:"message"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
