package service

import (
	"context"
	"fmt"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/messaging"
	"cinema-booking/internal/model"
	"cinema-booking/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookingService interface {
	LockSeats(ctx context.Context, userID string, req dto.LockSeatRequest) error
	CreateBooking(ctx context.Context, userID string, req dto.CreateBookingRequest) (*model.Booking, error)
}

type bookingService struct {
	bookingRepo  repository.BookingRepository
	showtimeRepo repository.ShowtimeRepository
	lockService  LockService
	publisher messaging.BookingPublisher
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	showtimeRepo repository.ShowtimeRepository,
	lockService LockService,
	publisher messaging.BookingPublisher,
) BookingService {
	return &bookingService{
		bookingRepo:  bookingRepo,
		showtimeRepo: showtimeRepo,
		lockService:  lockService,
		publisher: publisher,
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
			// if can't lock, rollback locked seats
			s.rollbackLocks(ctx, req.ShowtimeID, locked, userID)
			return fmt.Errorf("seat %s is already locked", seatNumber)
		}

		locked = append(locked, seatNumber)
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
	}

	// publish event to RabbitMQ
	event := messaging.BookingEvent{
		BookingID: booking.ID.Hex(),
		UserID: booking.UserID,
		ShowtimeID: req.ShowtimeID,
		SeatNumbers: req.SeatNumbers,
		CreatedAt: booking.CreatedAt,
	}
	if err := s.publisher.PublishBookingCreated(ctx, event); err != nil {
		fmt.Printf("Warning: failed to publish booking event: %v\n, err")
	}

	return booking, nil
}

// helpers
func (s *bookingService) rollbackLocks(ctx context.Context, showtimeID string, seats []string, userID string) {
	for _, seat := range seats {
		s.lockService.ReleaseLock(ctx, showtimeID, seat, userID)
	}
}
