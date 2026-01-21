package mocks

import (
	"context"
	"errors"
	"time"

	"github.com/razponn/Resilient-Scatter-Gather/internal/models"
)

var (
	ErrUserService        = errors.New("user service error")
	ErrPermissionsService = errors.New("permissions service error")
	ErrVectorMemory       = errors.New("vector memory error")
)

type UserServiceMock struct {
	Delay time.Duration
	Fail  bool
}

func (m UserServiceMock) GetUser(ctx context.Context, userID string) (models.User, error) {
	start := time.Now()
	if err := sleepCtx(ctx, m.Delay); err != nil {
		return models.User{}, err
	}
	_ = start // на будущее, если захотите писать метрики

	if m.Fail {
		return models.User{}, ErrUserService
	}

	return models.User{
		ID:   userID,
		Name: "Иван",
	}, nil
}

type PermissionsServiceMock struct {
	Delay   time.Duration
	Fail    bool
	Allowed bool
}

func (m PermissionsServiceMock) CheckAccess(ctx context.Context, userID, chatID string) (models.Permissions, error) {
	if err := sleepCtx(ctx, m.Delay); err != nil {
		return models.Permissions{}, err
	}

	if m.Fail {
		return models.Permissions{}, ErrPermissionsService
	}

	return models.Permissions{
		ChatID:  chatID,
		UserID:  userID,
		Allowed: m.Allowed,
	}, nil
}

type VectorMemoryMock struct {
	Delay time.Duration
	Fail  bool
}

func (m VectorMemoryMock) GetContext(ctx context.Context, chatID string) (models.VectorContext, error) {
	start := time.Now()
	if err := sleepCtx(ctx, m.Delay); err != nil {
		return models.VectorContext{}, err
	}

	if m.Fail {
		return models.VectorContext{}, ErrVectorMemory
	}

	return models.VectorContext{
		ChatID:    chatID,
		Snippet:   "Контекст из VectorMemory (пример)",
		Source:    "vector",
		LatencyMs: time.Since(start).Milliseconds(),
	}, nil
}

// sleepCtx — корректная "задержка"
// Это важно: иначе goroutine может продолжать работать после дедлайна
func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
