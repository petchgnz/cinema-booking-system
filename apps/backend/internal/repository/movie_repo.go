package repository

import (
	"context"
	"time"

	"cinema-booking/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// interface
type MovieRepository interface {
	Create(ctx context.Context, movie *model.Movie) error
	FindAll(ctx context.Context) ([]model.Movie, error)
	FindByID(ctx context.Context, id string) (*model.Movie, error)
}

type movieRepository struct {
	collection *mongo.Collection
}

// like a constructor in NestJS
func NewMovieRepository(db *mongo.Database) MovieRepository {
	return &movieRepository{
		collection: db.Collection("movies"),
	}
}

func (r *movieRepository) Create(ctx context.Context, movie *model.Movie) error {
	movie.ID = primitive.NewObjectID()
	movie.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, movie)
	return err
}

func (r *movieRepository) FindAll(ctx context.Context) ([]model.Movie, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var movies []model.Movie
	if err := cursor.All(ctx, &movies); err != nil {
		return nil, err
	}

	return movies, nil
}

func (r *movieRepository) FindByID(ctx context.Context, id string) (*model.Movie, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var movie model.Movie
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&movie)
	if err != nil {
		return nil, err
	}

	return &movie, nil
}
