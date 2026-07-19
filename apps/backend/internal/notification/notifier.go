package notification

import (
	"fmt"
	"log"
	"strings"
)

type Notifier interface {
	SendBookingConfirmed(bookingID, userID string, seats []string)
}

type MockNotifier struct{}

func NewMockNotifier() Notifier {
	return &MockNotifier{}
}

func (n *MockNotifier) SendBookingConfirmed(bookingID, userID string, seats []string) {
	log.Println(fmt.Sprintf(
		"[Notification] Booking confirmed | bookingID = %s | userID = %s | seats = [%s]",
		bookingID, userID, strings.Join(seats, ","),
	))
}
