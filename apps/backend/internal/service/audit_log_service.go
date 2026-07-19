package service

import (
	"context"
	"log"

	"cinema-booking/internal/model"
	"cinema-booking/internal/repository"
)

type auditLogService struct {
	repo repository.AuditLogRepository
}

func NewAuditLogService(repo repository.AuditLogRepository) AuditLogger {
	return &auditLogService{repo: repo}
}

// Log writes an audit event to MongoDB.
// It never blocks the caller — errors are only printed to the server log.
func (s *auditLogService) Log(ctx context.Context, eventType model.AuditEventType, userID, showtimeID string, seats []string, message string) {
	entry := &model.AuditLog{
		EventType:   eventType,
		UserID:      userID,
		ShowtimeID:  showtimeID,
		SeatNumbers: seats,
		Message:     message,
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		log.Printf("[AuditLog] Failed to write log: %v", err)
	}
}
