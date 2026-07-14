package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirebaseUID string             `bson:"firebase_uid" json:"firebase_uid"`
	Email       string             `bson:"email" json:"email"`
	Name        string             `bson:"name" json:"name"`
	PhotoURL    string             `bson:"photo_url" json:"photo_url"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
