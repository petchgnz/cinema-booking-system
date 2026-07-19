package repository

import (
	"context"
	"time"

	"cinema-booking/internal/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log *model.AuditLog) error
}

type auditLogRepository struct {
	collection *mongo.Collection
}

func NewAuditLogRepository(db *mongo.Database) AuditLogRepository {
	return &auditLogRepository{
		collection: db.Collection("audit_logs"),
	}
}

func (r *auditLogRepository) Create(ctx context.Context, log *model.AuditLog) error {
	log.ID = primitive.NewObjectID()
	log.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, log)
	return err
}
