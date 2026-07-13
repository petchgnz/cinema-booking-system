package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Movie struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Duration    int                `bson:"duration" json:"duration"`
	PosterURL   string             `bson:"poster_url" json:"poster_url"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}
