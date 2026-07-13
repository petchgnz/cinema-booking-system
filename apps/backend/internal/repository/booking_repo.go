package repository

import (
	"context"
	"time"

	"cinema-booking/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *model.Booking) error
	FindByID(ctx context.Context, id string) (*model.Booking, error)
	FindByUserID(ctx context.Context, userID string) ([]model.Booking, error)
	UpdateStatus(ctx context.Context, id string, status model.BookingStatus) error
}

type bookingRepository struct {
	collection *mongo.Collection
}

func NewBookingRepository(db *mongo.Database) BookingRepository {
	return &bookingRepository{
		collection: db.Collection("bookings"),
	}
}

func (r *bookingRepository) Create(ctx context.Context, booking *model.Booking) error {
	booking.ID = primitive.NewObjectID()
	booking.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, booking)
	return err
}

func (r *bookingRepository) FindByID(ctx context.Context, id string) (*model.Booking, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var booking model.Booking
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&booking)
	if err != nil {
		return nil, err
	}

	return &booking, nil
}

func (r *bookingRepository) FindByUserID(ctx context.Context, userID string) ([]model.Booking, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []model.Booking
	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}

	return bookings, nil
}

func (r *bookingRepository) UpdateStatus(ctx context.Context, id string, status model.BookingStatus) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}},
	)
	return err
}
