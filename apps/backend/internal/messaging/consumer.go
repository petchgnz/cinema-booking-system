package messaging

import (
	"context"
	"encoding/json"
	"log"

	"cinema-booking/internal/model"
	"cinema-booking/internal/notification"
	"cinema-booking/internal/repository"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	BookingQueue = "booking.confirmation"
)

type BookingConsumer struct {
	conn        *amqp.Connection
	bookingRepo repository.BookingRepository
	notifier    notification.Notifier
}

func NewBookingConsumer(conn *amqp.Connection, bookingRepo repository.BookingRepository, notifier notification.Notifier) *BookingConsumer {
	return &BookingConsumer{
		conn:        conn,
		bookingRepo: bookingRepo,
		notifier:    notifier,
	}
}

// Start consume message - for goroutine
func (c *BookingConsumer) Start() {
	ch, err := c.conn.Channel()
	if err != nil {
		log.Fatalf("[Consumer] Failed to open channel: %v", err)
	}

	// declare queue and bind with exchange
	q, err := ch.QueueDeclare(
		BookingQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[Consumer] Failed to declare queue: %v", err)
	}

	err = ch.QueueBind(
		q.Name,
		BookingCreatedKey,
		BookingExchange,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[Consumer] Failed to bind queue: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"booking-worker",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("[Consumer] Failed to register consumer: %v", err)
	}

	log.Printf("[Consumer] Waiting for booking events...")

	for msg := range msgs {
		c.handleMessage(msg)
	}
}

func (c *BookingConsumer) handleMessage(msg amqp.Delivery) {
	var event BookingEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("[Consumer] Failed to parse message: %v", err)
		msg.Nack(false, false) // requeu = false cuz can't parse message (bad message)
		return
	}

	log.Printf("[Consumer] Processing booking: %s", event.BookingID)

	ctx := context.Background()
	if err := c.bookingRepo.UpdateStatus(ctx, event.BookingID, model.BookingConfirmed); err != nil {
		log.Printf("[Consumer] Failed to confirm booking %s: %v", event.BookingID, err)
		msg.Nack(false, true) // requeue = true becuz db error could be retry
		return
	}

	log.Printf("[Consumer] Booking confirmed: %s", event.BookingID)
	c.notifier.SendBookingConfirmed(event.BookingID, event.UserID, event.SeatNumbers)
	msg.Ack(false)
}
