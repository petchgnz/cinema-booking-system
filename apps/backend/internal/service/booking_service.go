package service

import (
	"context"
	"fmt"
	"log"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/messaging"
	"cinema-booking/internal/model"
	"cinema-booking/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditLogger is a simple interface to avoid importing the repository directly
type AuditLogger interface {
	Log(ctx context.Context, eventType model.AuditEventType, userID, showtimeID string, seats []string, message string)
}

type SeatBroascaster interface {
	BroadcastSeatUpdate(showtimeID, eventType, seatNumber, status string)
}

type BookingService interface {
	LockSeats(ctx context.Context, userID string, req dto.LockSeatRequest) error
	CreateBooking(ctx context.Context, userID string, req dto.CreateBookingRequest) (*model.Booking, error)
}

type bookingService struct {
	bookingRepo  repository.BookingRepository
	showtimeRepo repository.ShowtimeRepository
	lockService  LockService
	publisher    messaging.BookingPublisher
	broadcaster  SeatBroascaster
	auditLogger  AuditLogger
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	showtimeRepo repository.ShowtimeRepository,
	lockService LockService,
	publisher messaging.BookingPublisher,
	broadcaster SeatBroascaster,
	auditLogger AuditLogger,
) BookingService {
	return &bookingService{
		bookingRepo:  bookingRepo,
		showtimeRepo: showtimeRepo,
		lockService:  lockService,
		publisher:    publisher,
		broadcaster:  broadcaster,
		auditLogger:  auditLogger,
	}
}

func (s *bookingService) LockSeats(ctx context.Context, userID string, req dto.LockSeatRequest) error {
	locked := []string{}

	for _, seatNumber := range req.SeatNumbers {
		ok, err := s.lockService.AcquireLock(ctx, req.ShowtimeID, seatNumber, userID)
		if err != nil {
			s.rollbackLocks(ctx, req.ShowtimeID, locked, userID)
			return err
		}

		if !ok {
			s.rollbackLocks(ctx, req.ShowtimeID, locked, userID)
			s.auditLogger.Log(ctx, model.AuditLockFail, userID, req.ShowtimeID, []string{seatNumber},
				fmt.Sprintf("seat %s is already locked by another user", seatNumber))
			return fmt.Errorf("seat %s is already locked", seatNumber)
		}

		locked = append(locked, seatNumber)

		log.Printf("[BookingService] Broadcasting lock: showtime=%s seat=%s", req.ShowtimeID, seatNumber)
		s.broadcaster.BroadcastSeatUpdate(req.ShowtimeID, "seat_locked", seatNumber, "locked")
	}

	return nil
}

// CreateBooking = create booking after lock seat
func (s *bookingService) CreateBooking(ctx context.Context, userID string, req dto.CreateBookingRequest) (*model.Booking, error) {
	// check that all seats still lock by current user
	for _, seatNumber := range req.SeatNumbers {
		isLocked, err := s.lockService.IsLockedByUser(ctx, req.ShowtimeID, seatNumber, userID)
		if err != nil {
			return nil, err
		}
		if !isLocked {
			return nil, fmt.Errorf("lock expired for seat %s, please select again", seatNumber)
		}
	}

	showtimeID, err := primitive.ObjectIDFromHex(req.ShowtimeID)
	if err != nil {
		return nil, fmt.Errorf("invalid showtime_id")
	}

	// create booking in mongodb
	booking := &model.Booking{
		UserID:      userID,
		ShowtimeID:  showtimeID,
		SeatNumbers: req.SeatNumbers,
		Status:      model.BookingPending,
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		return nil, err
	}

	// updtea seat status in showtime -> booked
	for _, seatNumber := range req.SeatNumbers {
		if err := s.showtimeRepo.UpdateSeatStatus(ctx, req.ShowtimeID, seatNumber, model.SeatBooked); err != nil {
			return nil, err
		}

		s.lockService.ReleaseLock(ctx, req.ShowtimeID, seatNumber, userID)

		log.Printf("[BookingService] Broadcasting Created Booking: showtime=%s seat=%s", req.ShowtimeID, seatNumber)
		s.broadcaster.BroadcastSeatUpdate(req.ShowtimeID, "seat_booked", seatNumber, "booked")
	}

	s.auditLogger.Log(ctx, model.AuditBookingSuccess, userID, req.ShowtimeID, req.SeatNumbers,
		fmt.Sprintf("booking %s confirmed for %d seat(s)", booking.ID.Hex(), len(req.SeatNumbers)))

	// publish event to RabbitMQ
	event := messaging.BookingEvent{
		BookingID:   booking.ID.Hex(),
		UserID:      booking.UserID,
		ShowtimeID:  req.ShowtimeID,
		SeatNumbers: req.SeatNumbers,
		CreatedAt:   booking.CreatedAt,
	}
	if err := s.publisher.PublishBookingCreated(ctx, event); err != nil {
		fmt.Printf("Warning: failed to publish booking event: %v\n", err)
	}

	return booking, nil
}

// helpers
func (s *bookingService) rollbackLocks(ctx context.Context, showtimeID string, seats []string, userID string) {
	for _, seat := range seats {
		s.lockService.ReleaseLock(ctx, showtimeID, seat, userID)
		s.auditLogger.Log(ctx, model.AuditSeatReleased, userID, showtimeID, []string{seat},
			fmt.Sprintf("lock rolled back for seat %s", seat))
	}
}
