package repository

import (
	"context"
	"time"

	"cinema-booking/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	FindByFirebaseUID(ctx context.Context, uid string) (*model.User, error)
	Upsert(ctx context.Context, user *model.User, adminEmail string) (*model.User, error)
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

func (r *userRepository) FindByFirebaseUID(ctx context.Context, uid string) (*model.User, error) {
	var user model.User

	err := r.collection.FindOne(ctx, bson.M{"firebase_uid": uid}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Upsert(ctx context.Context, user *model.User, adminEmail string) (*model.User, error) {
	filter := bson.M{"firebase_uid": user.FirebaseUID}

	// determine role: admin if email matches ADMIN_EMAIL, otherwise default to "user"
	// $setOnInsert ensures role is only set on first login — never overwritten after that
	role := model.RoleUser
	if adminEmail != "" && user.Email == adminEmail {
		role = model.RoleAdmin
	}

	update := bson.M{
		"$set": bson.M{
			"email":     user.Email,
			"name":      user.Name,
			"photo_url": user.PhotoURL,
		},
		"$setOnInsert": bson.M{
			"_id":          primitive.NewObjectID(),
			"firebase_uid": user.FirebaseUID,
			"role":         role,
			"created_at":   time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result model.User
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
