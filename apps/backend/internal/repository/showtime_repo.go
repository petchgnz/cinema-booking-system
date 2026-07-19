package repository

import (
	"context"
	"time"

	"cinema-booking/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ShowtimeRepository interface {
	Create(ctx context.Context, showtime *model.Showtime) error
	FindAll(ctx context.Context) ([]model.Showtime, error)
	FindByID(ctx context.Context, id string) (*model.Showtime, error)
	UpdateSeatStatus(ctx context.Context, showtimeID string, seatNumber string, status model.SeatStatus) error
}

type showtimeRepository struct {
	collection *mongo.Collection
}

func NewShowtimeRepository(db *mongo.Database) ShowtimeRepository {
	return &showtimeRepository{
		collection: db.Collection("showtimes"),
	}
}

func (r *showtimeRepository) Create(ctx context.Context, showtime *model.Showtime) error {
	showtime.ID = primitive.NewObjectID()
	showtime.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, showtime)
	return err
}

func (r *showtimeRepository) FindAll(ctx context.Context) ([]model.Showtime, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	var showtimes []model.Showtime
	if err := cursor.All(ctx, &showtimes); err != nil {
		return nil, err
	}

	return showtimes, nil
}

func (r *showtimeRepository) FindByID(ctx context.Context, id string) (*model.Showtime, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var showtime model.Showtime
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&showtime)
	if err != nil {
		return nil, err
	}

	return &showtime, nil
}

func (r *showtimeRepository) UpdateSeatStatus(ctx context.Context, showtimeID string, seatNumber string, status model.SeatStatus) error {
	objectID, err := primitive.ObjectIDFromHex(showtimeID)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id":               objectID,
		"seats.seat_number": seatNumber,
	}

	update := bson.M{
		"$set": bson.M{
			"seats.$.status": status,
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	return err
}
