package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirebaseUID string             `bson:"firebase_uid" json:"firebase_uid"`
	Email       string             `bson:"email" json:"email"`
	Name        string             `bson:"name" json:"name"`
	PhotoURL    string             `bson:"photo_url" json:"photo_url"`
	Role        UserRole           `bson:"role" json:"role"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
