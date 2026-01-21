package clients

import (
	"context"

	"github.com/razponn/Resilient-Scatter-Gather/internal/models"
)

type UserService interface {
	GetUser(ctx context.Context, userID string) (models.User, error)
}

type PermissionsService interface {
	CheckAccess(ctx context.Context, userID, chatID string) (models.Permissions, error)
}

type VectorMemory interface {
	GetContext(ctx context.Context, chatID string) (models.VectorContext, error)
}
