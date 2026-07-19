package service

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestLockService(t *testing.T) (LockService, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return NewLockService(client), mr
}

func TestAcquireLock_Success(t *testing.T) {
	svc, mr := newTestLockService(t)
	defer mr.Close()

	ctx := context.Background()
	ok, err := svc.AcquireLock(ctx, "show1", "A1", "user1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected lock to be acquired, got false")
	}
}

func TestAcquireLock_AlreadyLocked(t *testing.T) {
	svc, mr := newTestLockService(t)
	defer mr.Close()

	ctx := context.Background()

	// user1 lock ก่อน
	svc.AcquireLock(ctx, "show1", "A1", "user1")

	// user2 พยายาม lock seat เดียวกัน
	ok, err := svc.AcquireLock(ctx, "show1", "A1", "user2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected lock to fail for second user, got true")
	}
}

func TestReleaseLock_Success(t *testing.T) {
	svc, mr := newTestLockService(t)
	defer mr.Close()

	ctx := context.Background()
	svc.AcquireLock(ctx, "show1", "A1", "user1")

	err := svc.ReleaseLock(ctx, "show1", "A1", "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// หลัง release แล้ว user อื่นควร lock ได้
	ok, _ := svc.AcquireLock(ctx, "show1", "A1", "user2")
	if !ok {
		t.Error("expected seat to be available after release, but lock failed")
	}
}

func TestReleaseLock_WrongUser(t *testing.T) {
	svc, mr := newTestLockService(t)
	defer mr.Close()

	ctx := context.Background()
	svc.AcquireLock(ctx, "show1", "A1", "user1")

	// user2 พยายาม release lock ของ user1
	err := svc.ReleaseLock(ctx, "show1", "A1", "user2")
	if err == nil {
		t.Error("expected error when releasing another user's lock, got nil")
	}
}

func TestIsLockedByUser_True(t *testing.T) {
	svc, mr := newTestLockService(t)
	defer mr.Close()

	ctx := context.Background()
	svc.AcquireLock(ctx, "show1", "A1", "user1")

	ok, err := svc.IsLockedByUser(ctx, "show1", "A1", "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected user1 to own the lock")
	}
}

func TestIsLockedByUser_False_OtherUser(t *testing.T) {
	svc, mr := newTestLockService(t)
	defer mr.Close()

	ctx := context.Background()
	svc.AcquireLock(ctx, "show1", "A1", "user1")

	// user2 ถามว่าตัวเองเป็นเจ้าของ lock ไหม
	ok, err := svc.IsLockedByUser(ctx, "show1", "A1", "user2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected user2 to NOT own the lock")
	}
}

func TestIsLockedByUser_False_Expired(t *testing.T) {
	svc, mr := newTestLockService(t)
	defer mr.Close()

	ctx := context.Background()
	svc.AcquireLock(ctx, "show1", "A1", "user1")

	// จำลอง TTL หมด
	mr.FastForward(lockDuration + 1)

	ok, err := svc.IsLockedByUser(ctx, "show1", "A1", "user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected lock to be expired")
	}
}