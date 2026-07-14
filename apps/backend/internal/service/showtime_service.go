package service

import (
	"context"
	"fmt"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/model"
	"cinema-booking/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShowtimeService interface {
	CreateShowtime(ctx context.Context, req dto.CreateShowtimeRequest) (*model.Showtime, error)
	GetAllShowtimes(ctx context.Context) ([]model.Showtime, error)
	GetShowtimeByID(ctx context.Context, id string) (*model.Showtime, error)
}

type showtimeService struct {
	showtimeRepo repository.ShowtimeRepository
}

// like a nestjs constructor
func NewShowtimeService(repo repository.ShowtimeRepository) ShowtimeService {
	return &showtimeService{showtimeRepo: repo}
}

func (s *showtimeService) CreateShowtime(ctx context.Context, req dto.CreateShowtimeRequest) (*model.Showtime, error) {
	movieID, err := primitive.ObjectIDFromHex(req.MovieID)
	if err != nil {
		return nil, fmt.Errorf("Invalid movie_id: %w", err)
	}

	seats := generateSeats(req.SeatCount)

	showtime := &model.Showtime{
		MovieID:   movieID,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Hall:      req.Hall,
		Seats:     seats,
	}

	if err := s.showtimeRepo.Create(ctx, showtime); err != nil {
		return nil, err
	}

	return showtime, nil
}

func (s *showtimeService) GetAllShowtimes(ctx context.Context) ([]model.Showtime, error) {
	return s.showtimeRepo.FindAll(ctx)
}

func (s *showtimeService) GetShowtimeByID(ctx context.Context, id string) (*model.Showtime, error) {
	return s.showtimeRepo.FindByID(ctx, id)
}

// helpers
func generateSeats(count int) []model.Seat {
	seats := make([]model.Seat, 0, count)
	rows := "ABCDEFGHIJ"
	cols := 20

	for i := 0; i < count; i++ {
		row := string(rows[i/cols])
		col := (i % cols) + 1
		seats = append(seats, model.Seat{
			SeatNumber: fmt.Sprintf("%s%d", row, col),
			Status:     model.SeatAvailable,
		})
	}

	return seats
}
