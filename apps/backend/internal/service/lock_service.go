package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// lock duration = 5 mins
const lockDuration = 5 * time.Minute

type LockService interface {
	AcquireLock(ctx context.Context, showtimeID, seatNumber, userID string) (bool, error)
	ReleaseLock(ctx context.Context, showtimeID, seatNumber, userID string) error
	IsLockedByUser(ctx context.Context, showtimeID, seatNumber, userID string) (bool, error)
}

type lockService struct {
	redis *redis.Client
}

func NewLockService(redis *redis.Client) LockService {
	return &lockService{redis: redis}
}

// a func to create redis key. like "lock:seat:show1:A1"
func generateLockKey(showtimeID, seatNumber string) string {
	return fmt.Sprintf("lock:seat:%s:%s", showtimeID, seatNumber)
}

// try to lock seat. if success, return true; and false if seat is taken
func (s *lockService) AcquireLock(ctx context.Context, showtimeID, seatNumber, userID string) (bool, error) {
	key := generateLockKey(showtimeID, seatNumber)

	ok, err := s.redis.SetNX(ctx, key, userID, lockDuration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return ok, nil
}

func (s *lockService) ReleaseLock(ctx context.Context, showtimeID, seatNumber, userID string) error {
	key := generateLockKey(showtimeID, seatNumber)

	val, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		// if exist key is not found = ok, (already release)
		return nil
	}

	if val != userID {
		return fmt.Errorf("lock belongs to another user")
	}

	return s.redis.Del(ctx, key).Err()
}

// check if this lock is from this user 
func (s *lockService) IsLockedByUser(ctx context.Context, showtimeID, seatNumber, userID string) (bool, error) {
	key := generateLockKey(showtimeID, seatNumber)

	val, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // key is expired
	}

	if err != nil {
		return false, err
	}

	return val == userID, nil
}
