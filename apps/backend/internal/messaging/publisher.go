package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	BookingExchange   = "booking.exchange"
	BookingCreatedKey = "booking.created"
)

// struct of data that send into queue
type BookingEvent struct {
	BookingID   string    `json:"booking_id"`
	UserID      string    `json:"user_id"`
	ShowtimeID  string    `json:"showtime_id"`
	SeatNumbers []string  `json:"seat_numbers"`
	CreatedAt   time.Time `json:"created_at"`
}

// interface for publish event
type BookingPublisher interface {
	PublishBookingCreated(ctx context.Context, event BookingEvent) error
}

type bookingPublisher struct {
	conn *amqp.Connection
}

func NewBookingPublisher(conn *amqp.Connection) (BookingPublisher, error) {
	p := &bookingPublisher{conn: conn}

	if err := p.setupExchange(); err != nil {
		return nil, fmt.Errorf("failed to setup exchange: %w", err)
	}

	return p, nil
}

// declare exchange
func (p *bookingPublisher) setupExchange() error {
	ch, err := p.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return ch.ExchangeDeclare(
		BookingExchange,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
}

func (p *bookingPublisher) PublishBookingCreated(ctx context.Context, event BookingEvent) error {
	ch, err := p.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = ch.PublishWithContext(
		ctx,
		BookingExchange,
		BookingCreatedKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}

	log.Printf("[Publisher] Booking event published: %s", event.BookingID)
	return nil
}
