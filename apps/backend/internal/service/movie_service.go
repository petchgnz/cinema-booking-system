package service

import (
	"context"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/model"
	"cinema-booking/internal/repository"
)

type MovieService interface {
	CreateMovie(ctx context.Context, req dto.CreateMovieRequest) (*model.Movie, error)
	GetAllMovies(ctx context.Context) ([]model.Movie, error)
	GetMovieByID(ctx context.Context, id string) (*model.Movie, error)
}

// like a nestjs constructor
type movieService struct {
	movieRepo repository.MovieRepository
}

func NewMovieService(repo repository.MovieRepository) MovieService {
	return &movieService{movieRepo: repo}
}

func (s *movieService) CreateMovie(ctx context.Context, req dto.CreateMovieRequest) (*model.Movie, error) {
	movie := &model.Movie{
		Title:       req.Title,
		Description: req.Description,
		Duration:    req.Duration,
		PosterURL:   req.PosterURL,
	}

	if err := s.movieRepo.Create(ctx, movie); err != nil {
		return nil, err
	}

	return movie, nil
}

func (s *movieService) GetAllMovies(ctx context.Context) ([]model.Movie, error) {
	return s.movieRepo.FindAll(ctx)
}

func (s *movieService) GetMovieByID(ctx context.Context, id string) (*model.Movie, error) {
	return s.movieRepo.FindByID(ctx, id)
}
